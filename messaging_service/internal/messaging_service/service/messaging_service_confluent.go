package service

import (
	"context"
	pb "cra-protos/messaging_service"
	"encoding/json"
	"fmt"
	"messaging_service/internal/config"
	"messaging_service/pkg/confluent"
	"messaging_service/pkg/errors"
	"messaging_service/pkg/logger"
	"sync"
	"time"
)

type MessagingService interface {
	PublishMessage(ctx context.Context, cfg confluent.KafkaConfig, req *pb.PublishRequest) *errors.CustomError
	ConsumeMessage(stream pb.MessagingService_SubscribeV1Server, cfg confluent.KafkaConfig) *errors.CustomError
	CreateTopic(ctx context.Context, req *pb.CreateTopicRequest, cfg confluent.KafkaConfig) *errors.CustomError
}

// ConfluentMessagingService implements MessagingService using confluent-kafka-go
type ConfluentMessagingService struct {
	Factory     confluent.KafkaFactory
	Admin       confluent.KafkaAdmin
	Producers   map[string]confluent.Producer
	Config      *config.Config
	KafkaConfig confluent.KafkaConfig
	mu          sync.RWMutex // Mutex for thread-safe access to producers map
}

// NewConfluentMessagingService creates a new messaging service using confluent-kafka-go
func NewConfluentMessagingService(config *config.Config, factory confluent.KafkaFactory) (MessagingService, error) {
	// Create factory

	// Create a base config
	kafkaConfig := confluent.KafkaConfig{
		Brokers:           config.KafkaBrokers,
		NumPartitions:     config.KafkaNumPartitions,
		ReplicationFactor: config.KafkaReplicationFactor,
		BatchSize:         config.KafkaBatchSize,
		BatchBytes:        config.KafkaBatchBytes,
		BatchTimeout:      time.Duration(config.KafkaBatchTimeoutMs) * time.Millisecond,
		CompressionType:   config.KafkaCompressionCodec,
		MaxAttempts:       config.KafkaMaxAttempts,
		RetryBackoffMs:    config.KafkaRetryBackoffMs,
		ReadTimeout:       time.Duration(config.KafkaReadTimeoutMs) * time.Millisecond,
		WriteTimeout:      time.Duration(config.KafkaWriteTimeoutMs) * time.Millisecond,
		DeliverySemantics: confluent.ExactlyOnce, // Default to exactly-once
		ExactlyOnceConfig: confluent.ExactlyOnceConfig{
			EnableIdempotence:     true,
			EnableTransactions:    true,
			TransactionTimeoutMs:  10000, // 10s
			TransactionalIDPrefix: "txn",
			IsolationLevel:        "read_committed",
		},
		ConsumerConfig: confluent.ConsumerConfig{
			AutoOffsetReset:   config.KafkaConsumerAutoOffsetReset,
			AutoCommit:        config.KafkaEnableAutoCommit,
			CommitInterval:    time.Duration(config.KafkaConsumerCommitIntervalMs) * time.Millisecond,
			SessionTimeout:    time.Duration(config.KafkaConsumerSessionTimeoutMs) * time.Millisecond,
			HeartbeatInterval: time.Duration(config.KafkaConsumerHeartbeatMs) * time.Millisecond,
			MaxPollRecords:    config.KafkaConsumerMaxPollRecords,
		},
	}

	// Create admin client with retry logic
	var admin confluent.KafkaAdmin
	var err error

	// Retry admin creation a few times before falling back to null admin
	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
		admin, err = factory.CreateAdmin(config)
		if err == nil {
			break
		}
		logger.LogWarnEvent("", "confluent_admin_creation_retry", "", "warn",
			fmt.Sprintf("Failed to create Confluent Kafka admin (attempt %d/%d): %v", i+1, maxRetries, err))
		time.Sleep(time.Duration(500*(i+1)) * time.Millisecond) // Exponential backoff
	}

	if err != nil {
		return &ConfluentMessagingService{}, fmt.Errorf("failed to create Confluent Kafka admin after %d attempts: %v", maxRetries, err)
	}

	return &ConfluentMessagingService{
		Factory:     factory,
		Admin:       admin,
		Producers:   make(map[string]confluent.Producer),
		Config:      config,
		KafkaConfig: kafkaConfig,
		mu:          sync.RWMutex{},
	}, nil
}

// PublishMessage publishes a message to Kafka
func (s *ConfluentMessagingService) PublishMessage(ctx context.Context, confluentConfig confluent.KafkaConfig, req *pb.PublishRequest) *errors.CustomError {

	confluentConfig.Topic = req.Topic

	// First check if topic exists
	topicExists, err := s.TopicExists(ctx, req.Topic)
	if err != nil {
		logger.LogWarnEvent("", "kafka_check_topic_failed", req.Topic, "warn",
			fmt.Sprintf("Failed to check if topic exists: %v", err))
		return errors.NewCustomError(
			errors.PUBErrTopicNotExists,
			fmt.Errorf("topic able to get existing topic: %s", req.Topic),
		)
		// Continue anyway with a warning
	} else if !topicExists {
		return errors.NewCustomError(
			errors.PUBErrTopicNotExists,
			fmt.Errorf("topic %s does not exist", req.Topic),
		)
	}

	// Check if topic has active consumers
	if topicExists {
		hasConsumers, err := s.Admin.HasActiveConsumers(ctx, req.Topic)
		if err != nil {
			logger.LogWarnEvent("", "kafka_check_consumers_failed", req.Topic, "warn",
				fmt.Sprintf("Failed to check for active consumers: %v", err))
			// Continue anyway, don't fail the publish
		} else if !hasConsumers && s.Config.KafkaRequireActiveListener {
			// Only return an error if active listeners are required by configuration
			return errors.NewCustomError(
				errors.PUBErrTopicNotExists,
				fmt.Errorf("no active consumers found for topic: %s", req.Topic),
			)
		} else if !hasConsumers {
			// Just log a warning if active listeners aren't required
			logger.LogWarnEvent("", "kafka_no_active_consumers", req.Topic, "warn",
				fmt.Sprintf("No active consumers found for topic %s, but continuing with publish due to configuration", req.Topic))
		}
	}

	// Get or create producer
	producer, err := s.GetOrCreateProducer(confluentConfig)
	if err != nil {
		return errors.NewCustomError(
			errors.PUBErrProducerNotReady,
			fmt.Errorf("failed to create producer: %v", err),
		)
	}

	// Parse value
	var value []byte
	// Convert map to JSON
	jsonBytes, err := json.Marshal(req.Value)
	if err != nil {
		return errors.NewCustomError(
			errors.PUBErrInvalidMessage,
			fmt.Errorf("failed to serialize message value: %v", err),
		)
	}
	value = jsonBytes

	var key []byte
	if req.Key != "" {
		key = []byte(req.Key)
	} else {
		// Generate a default key if none provided
		key = []byte(fmt.Sprintf("%s-%d", confluentConfig.Topic, time.Now().UnixNano()))
	}

	// No headers in the protobuf definition, so we'll skip that part

	// Track transaction start time
	var txStartTime time.Time

	// Begin transaction if using exactly-once
	if confluentConfig.DeliverySemantics == confluent.ExactlyOnce &&
		confluentConfig.ExactlyOnceConfig.EnableTransactions {

		txStartTime = time.Now()
		err = producer.BeginTransaction()
		if err != nil {
			// Log the error and try to recreate the producer
			logger.LogErrorEvent("", "transaction_begin_failed", confluentConfig.Topic, "error",
				fmt.Sprintf("Failed to begin transaction, will retry with new producer: %v", err))

			// Remove the producer from the cache to force creation of a new one
			s.mu.Lock()
			delete(s.Producers, confluentConfig.Topic)
			s.mu.Unlock()

			// Try to create a new producer
			producer, err = s.GetOrCreateProducer(confluentConfig)
			if err != nil {
				return errors.NewCustomError(
					errors.PUBErrProducerNotReady,
					fmt.Errorf("failed to recreate producer after transaction failure: %v", err),
				)
			}

			// Try to begin transaction again
			err = producer.BeginTransaction()
			if err != nil {
				return errors.NewCustomError(
					errors.KAFErrConnectionFailed,
					fmt.Errorf("failed to begin transaction after producer recreation: %v", err),
				)
			}

			// Log successful retry
			logger.LogEvent("", "transaction_begin_retry_success", confluentConfig.Topic, "info",
				fmt.Sprintf("Successfully began transaction after producer recreation (took %v)", time.Since(txStartTime)))
		}
	}

	// Track publish time
	pubStartTime := time.Now()

	// Publish message with retry logic already built into the WriteWithRetry method
	err = producer.WriteWithRetry(ctx, key, value, nil)

	// Log publish timing
	pubDuration := time.Since(pubStartTime)
	if pubDuration > 500*time.Millisecond {
		logger.LogWarnEvent("", "kafka_publish_slow", confluentConfig.Topic, "warn",
			fmt.Sprintf("Slow message publish: %v", pubDuration))
	}

	if err != nil {
		// Abort transaction if it was started
		if confluentConfig.DeliverySemantics == confluent.ExactlyOnce &&
			confluentConfig.ExactlyOnceConfig.EnableTransactions {
			abortErr := producer.AbortTransaction(ctx)
			if abortErr != nil {
				logger.LogErrorEvent("", "transaction_abort_failed", confluentConfig.Topic, "error",
					fmt.Sprintf("failed to abort transaction: %v", abortErr))
			}
		}

		logger.LogEvent("", "kafka_publish_failed", req.Topic, "error",
			fmt.Sprintf("Failed to publish message: %v", err))

		return errors.NewCustomError(
			errors.PUBErrPublishFailed,
			err,
		)
	}

	// Commit transaction if using exactly-once
	if confluentConfig.DeliverySemantics == confluent.ExactlyOnce &&
		confluentConfig.ExactlyOnceConfig.EnableTransactions {

		// Track commit time
		commitStartTime := time.Now()

		err = producer.CommitTransaction(ctx)

		// Log commit timing
		commitDuration := time.Since(commitStartTime)
		if commitDuration > 500*time.Millisecond {
			logger.LogWarnEvent("", "kafka_commit_slow", confluentConfig.Topic, "warn",
				fmt.Sprintf("Slow transaction commit: %v", commitDuration))
		}

		if err != nil {
			return errors.NewCustomError(
				errors.KAFErrConnectionFailed,
				err,
			)
		}

		// Log total transaction time
		totalTxDuration := time.Since(txStartTime)
		if totalTxDuration > 1*time.Second {
			logger.LogWarnEvent("", "kafka_transaction_slow", confluentConfig.Topic, "warn",
				fmt.Sprintf("Slow transaction (total time: %v, publish: %v, commit: %v)",
					totalTxDuration, pubDuration, commitDuration))
		}
	}

	// Log successful publish
	logger.LogEvent("", "kafka_publish_success", req.Topic, "info",
		fmt.Sprintf("Successfully published message to topic %s", req.Topic))

	return nil
}

// ConsumeMessage implements the consumer streaming RPC
func (s *ConfluentMessagingService) ConsumeMessage(stream pb.MessagingService_SubscribeV1Server, confluentConfig confluent.KafkaConfig) *errors.CustomError {
	// Extract topic and group information
	ctx := stream.Context()

	groupID := confluentConfig.ConsumerConfig.GroupID

	// Check if topic exists
	topicExists, err := s.TopicExists(ctx, confluentConfig.Topic)
	if err != nil {
		logger.LogErrorEvent("", "kafka_check_topic_failed", confluentConfig.Topic, "error",
			fmt.Sprintf("Failed to check if topic exists: %v", err))
		return errors.NewCustomError(
			errors.SUBErrTopicNotExists,
			err,
		)
	}

	if !topicExists {
		logger.LogErrorEvent("", "kafka_topic_not_found", confluentConfig.Topic, "error",
			fmt.Sprintf("Topic %s does not exist", confluentConfig.Topic))
		return errors.NewCustomError(
			errors.SUBErrTopicNotExists,
			fmt.Errorf(confluentConfig.Topic),
		)
	}

	// Create a new consumer with retry logic
	var consumer confluent.Consumer
	var errConsumer error
	maxRetries := 3

	for i := 0; i < maxRetries; i++ {
		consumer, errConsumer = s.Factory.CreateConsumer(
			confluentConfig,
			groupID,
			func(data []byte) error {
				// Parse message
				var value map[string]interface{}
				if err := json.Unmarshal(data, &value); err != nil {
					return confluent.NewConsumerError(err, true, "failed to parse message")
				}

				// Convert to string map for compatibility
				strValue := make(map[string]string)
				for k, v := range value {
					strValue[k] = fmt.Sprintf("%v", v)
				}

				// Create response
				msg := &pb.KafkaMessage{
					Value:     strValue,
					Timestamp: time.Now().UnixMilli(),
				}

				// Send the message to the stream
				if err := stream.Send(msg); err != nil {
					logger.LogEvent("", "stream_send_failed", confluentConfig.Topic, "error", err.Error())
					return confluent.NewConsumerError(err, false, "failed to send message to stream")
				}

				return nil
			},
		)

		if errConsumer == nil {
			break
		}

		logger.LogWarnEvent("", "consumer_creation_retry", confluentConfig.Topic, "warn",
			fmt.Sprintf("Failed to create consumer (attempt %d/%d): %v", i+1, maxRetries, errConsumer))

		// Wait before retrying
		select {
		case <-ctx.Done():
			return errors.NewCustomError(
				errors.SUBErrConsumerNotReady,
				fmt.Errorf("context cancelled while creating consumer: %v", ctx.Err()),
			)
		case <-time.After(time.Duration(500*(i+1)) * time.Millisecond):
			// Continue with retry
		}
	}

	if errConsumer != nil {
		return errors.NewCustomError(
			errors.SUBErrConsumerNotReady,
			fmt.Errorf("failed to create consumer after %d attempts: %v", maxRetries, errConsumer),
		)
	}

	// Start consuming
	consumer.Start(ctx)

	// Wait for the stream to be closed
	<-ctx.Done()

	// Close the consumer with a timeout to ensure proper cleanup
	closeCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	closeErr := consumer.Close()
	if closeErr != nil {
		logger.LogErrorEvent("", "consumer_close_error", confluentConfig.Topic, "error",
			fmt.Sprintf("Error while closing consumer: %v", closeErr))
	}

	// Wait for close to complete or timeout
	<-closeCtx.Done()

	return nil
}

func (s *ConfluentMessagingService) CreateTopic(ctx context.Context, req *pb.CreateTopicRequest, cfg confluent.KafkaConfig) *errors.CustomError {
	// Check if topic already exists
	topicExists, err := s.TopicExists(ctx, req.Topic)
	if err != nil {
		logger.LogWarnEvent("", "kafka_check_topic_failed", req.Topic, "warn",
			fmt.Sprintf("Failed to check if topic exists: %v", err))
	} else if topicExists {
		logger.LogEvent("", "kafka_topic_exists", req.Topic, "info",
			fmt.Sprintf("Topic %s already exists", req.Topic))
		return errors.NewCustomError(
			errors.TOPErrTopicExists,
			fmt.Errorf("Topic %s already exists", req.Topic),
		)
	}

	// Create the topic with retry logic
	var lastErr error
	for i := 0; i < 3; i++ {
		if ctx.Err() != nil {
			return errors.NewCustomError(errors.TOPErrCreateFailed, fmt.Errorf("Context cancelled: %v", ctx.Err()))
		}

		err := s.Admin.CreateTopic(
			ctx,
			req.Topic,
			cfg.NumPartitions,
			s.Config.KafkaReplicationFactor,
			nil,
		)

		if err == nil {
			return nil
		}

		lastErr = err
		logger.LogWarnEvent("", "create_topic_retry", req.Topic, "warn",
			fmt.Sprintf("Failed to create topic (attempt %d/3): %v", i+1, err))

		select {
		case <-ctx.Done():
			return errors.NewCustomError(errors.TOPErrCreateFailed, fmt.Errorf("Context cancelled: %v", ctx.Err()))
		case <-time.After(time.Duration(500*(i+1)) * time.Millisecond):
		}
	}

	logger.LogEvent("kafka", "create_topic_failed", req.Topic, "error", lastErr.Error())
	return errors.NewCustomError(
		errors.TOPErrCreateFailed,
		fmt.Errorf("Failed to create topic after 3 attempts: %v", lastErr),
	)
}

// GetOrCreateProducer gets an existing producer for a topic or creates a new one
func (s *ConfluentMessagingService) GetOrCreateProducer(cfg confluent.KafkaConfig) (confluent.Producer, error) {

	// Check if we already have a producer for this topic
	s.mu.RLock()
	producer, exists := s.Producers[cfg.Topic]
	s.mu.RUnlock()

	if exists {
		return producer, nil
	}

	// Create a new producer with thread safety
	s.mu.Lock()
	defer s.mu.Unlock()

	// Double-check if another goroutine created the producer while we were waiting for the lock
	if producer, exists = s.Producers[cfg.Topic]; exists {
		return producer, nil
	}

	// Create a new producer
	producer, err := s.Factory.CreateProducer(cfg)
	if err != nil {
		return nil, err
	}

	s.Producers[cfg.Topic] = producer
	return producer, nil
}

// TopicExists checks if a topic exists in Kafka
func (s *ConfluentMessagingService) TopicExists(ctx context.Context, topic string) (bool, error) {
	// Get list of topics
	topics, err := s.Admin.ListTopics(ctx)
	if err != nil {
		return false, err
	}

	// Check if our topic is in the list
	for _, t := range topics {
		if t == topic {
			return true, nil
		}
	}

	return false, nil
}
