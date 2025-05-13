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

// ConfluentMessagingService implements MessagingService using confluent-kafka-go
type ConfluentMessagingService struct {
	factory   confluent.KafkaFactory
	admin     confluent.KafkaAdmin
	producers map[string]confluent.Producer
	config    *config.Config
	mu        sync.RWMutex // Mutex for thread-safe access to producers map
}

// NewConfluentMessagingService creates a new messaging service using confluent-kafka-go
func NewConfluentMessagingService(config *config.Config) MessagingService {
	// Create factory
	factory := confluent.NewFactory()

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
		admin, err = factory.CreateAdmin(kafkaConfig)
		if err == nil {
			break
		}
		logger.LogWarnEvent("", "confluent_admin_creation_retry", "", "warn",
			fmt.Sprintf("Failed to create Confluent Kafka admin (attempt %d/%d): %v", i+1, maxRetries, err))
		time.Sleep(time.Duration(500*(i+1)) * time.Millisecond) // Exponential backoff
	}

	if err != nil {
		logger.LogErrorEvent("", "confluent_admin_creation_failed", "", "error",
			fmt.Sprintf("Failed to create Confluent Kafka admin after %d attempts: %v", maxRetries, err))
		// Create a null admin that logs errors but doesn't fail
		admin = &ConfluentAdmin{}
	}

	return &ConfluentMessagingService{
		factory:   factory,
		admin:     admin,
		producers: make(map[string]confluent.Producer),
		config:    config,
		mu:        sync.RWMutex{},
	}
}

// PublishMessage publishes a message to Kafka
func (s *ConfluentMessagingService) PublishMessage(ctx context.Context, cfg interface{}, req *pb.PublishRequest) *errors.CustomError {
	// Input validation
	if req.Topic == "" {
		return errors.NewCustomError(
			errors.MSGErrInvalidTopic,
			fmt.Errorf("Topic is required"),
		)
	}

	// Convert to confluent config if needed
	var confluentConfig confluent.KafkaConfig
	switch c := cfg.(type) {
	case confluent.KafkaConfig:
		// Already the right type
		confluentConfig = c
		// Set the topic from the request
		confluentConfig.Topic = req.Topic
	default:
		// Unknown config type
		return errors.NewCustomError(
			errors.PUBErrInvalidConfig,
			fmt.Errorf("Unsupported config type"),
		)
	}

	// First check if topic exists
	topicExists, err := s.topicExists(ctx, req.Topic)
	if err != nil {
		logger.LogWarnEvent("", "kafka_check_topic_failed", req.Topic, "warn",
			fmt.Sprintf("Failed to check if topic exists: %v", err))
		// Continue anyway with a warning
	} else if !topicExists {
		return errors.NewCustomError(
			errors.PUBErrTopicNotExists,
			fmt.Errorf("Topic does not exist: %s", req.Topic),
		)
	}

	// Check if topic has active consumers
	if topicExists {
		hasConsumers, err := s.admin.HasActiveConsumers(ctx, req.Topic)
		if err != nil {
			logger.LogWarnEvent("", "kafka_check_consumers_failed", req.Topic, "warn",
				fmt.Sprintf("Failed to check for active consumers: %v", err))
			// Continue anyway, don't fail the publish
		} else if !hasConsumers && s.config.KafkaRequireActiveListener {
			// Only return an error if active listeners are required by configuration
			return errors.NewCustomError(
				errors.PUBErrTopicNotExists,
				fmt.Errorf("No active consumers found for topic: %s", req.Topic),
			)
		} else if !hasConsumers {
			// Just log a warning if active listeners aren't required
			logger.LogWarnEvent("", "kafka_no_active_consumers", req.Topic, "warn",
				fmt.Sprintf("No active consumers found for topic %s, but continuing with publish due to configuration", req.Topic))
		}
	}

	// Get or create producer
	producer, err := s.getOrCreateProducer(confluentConfig)
	if err != nil {
		return errors.NewCustomError(
			errors.PUBErrProducerNotReady,
			fmt.Errorf("Failed to create producer: %v", err),
		)
	}

	// Parse value
	var value []byte
	if len(req.Value) > 0 {
		// Convert map to JSON
		jsonBytes, err := json.Marshal(req.Value)
		if err != nil {
			return errors.NewCustomError(
				errors.PUBErrInvalidMessage,
				fmt.Errorf("Failed to serialize message value: %v", err),
			)
		}
		value = jsonBytes
	} else {
		return errors.NewCustomError(
			errors.PUBErrInvalidMessage,
			fmt.Errorf("Message value is required"),
		)
	}

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
			delete(s.producers, confluentConfig.Topic)
			s.mu.Unlock()

			// Try to create a new producer
			producer, err = s.getOrCreateProducer(confluentConfig)
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
					fmt.Sprintf("Failed to abort transaction: %v", abortErr))
			}
		}

		logger.LogEvent("", "kafka_publish_failed", req.Topic, "error",
			fmt.Sprintf("Failed to publish message: %v", err))

		return errors.NewCustomError(
			errors.PUBErrPublishFailed,
			fmt.Errorf("Failed to publish message: %v", err),
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
				fmt.Errorf("Failed to commit transaction: %v", err),
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
func (s *ConfluentMessagingService) ConsumeMessage(stream pb.MessagingService_SubscribeV1Server, cfg interface{}) *errors.CustomError {
	// Extract topic and group information
	ctx := stream.Context()

	// Convert to confluent config if needed
	var confluentConfig confluent.KafkaConfig
	var groupID string

	// Handle different config types
	switch c := cfg.(type) {
	case confluent.KafkaConfig:
		// Already the right type
		confluentConfig = c
		groupID = c.ConsumerConfig.GroupID
	default:
		// Unknown config type
		return errors.NewCustomError(
			errors.SUBErrInvalidConfig,
			fmt.Errorf("Unsupported config type"),
		)
	}

	// Input validation
	if confluentConfig.Topic == "" {
		return errors.NewCustomError(
			errors.MSGErrInvalidTopic,
			fmt.Errorf("Topic is required"),
		)
	}

	if groupID == "" {
		return errors.NewCustomError(
			errors.SUBErrInvalidConfig,
			fmt.Errorf("Consumer group ID is required"),
		)
	}

	// Check if topic exists
	topicExists, err := s.topicExists(ctx, confluentConfig.Topic)
	if err != nil {
		logger.LogErrorEvent("", "kafka_check_topic_failed", confluentConfig.Topic, "error",
			fmt.Sprintf("Failed to check if topic exists: %v", err))
		return errors.NewCustomError(
			errors.SUBErrTopicNotExists,
			fmt.Errorf("Failed to check if topic exists: %v", err),
		)
	}

	if !topicExists {
		logger.LogErrorEvent("", "kafka_topic_not_found", confluentConfig.Topic, "error",
			fmt.Sprintf("Topic %s does not exist", confluentConfig.Topic))
		return errors.NewCustomError(
			errors.SUBErrTopicNotExists,
			fmt.Errorf("Topic does not exist: %s", confluentConfig.Topic),
		)
	}

	// Create a new consumer with retry logic
	var consumer confluent.Consumer
	var errConsumer error
	maxRetries := 3

	for i := 0; i < maxRetries; i++ {
		consumer, errConsumer = s.factory.CreateConsumer(
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
				fmt.Errorf("Context cancelled while creating consumer: %v", ctx.Err()),
			)
		case <-time.After(time.Duration(500*(i+1)) * time.Millisecond):
			// Continue with retry
		}
	}

	if errConsumer != nil {
		return errors.NewCustomError(
			errors.SUBErrConsumerNotReady,
			fmt.Errorf("Failed to create consumer after %d attempts: %v", maxRetries, errConsumer),
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

// CreateTopic creates a new Kafka topic
func (s *ConfluentMessagingService) CreateTopic(ctx context.Context, req *pb.CreateTopicRequest, cfg interface{}) *errors.CustomError {
	// Input validation
	if req.Topic == "" {
		return errors.NewCustomError(
			errors.MSGErrInvalidTopic,
			fmt.Errorf("Topic name is required"),
		)
	}

	// Check if topic already exists
	topicExists, err := s.topicExists(ctx, req.Topic)
	if err != nil {
		logger.LogWarnEvent("", "kafka_check_topic_failed", req.Topic, "warn",
			fmt.Sprintf("Failed to check if topic exists: %v", err))
		// Continue anyway with topic creation
	} else if topicExists {
		// Topic already exists, return success with a specific error code
		logger.LogEvent("", "kafka_topic_exists", req.Topic, "info",
			fmt.Sprintf("Topic %s already exists", req.Topic))
		return errors.NewCustomError(
			errors.TOPErrTopicExists,
			fmt.Errorf("Topic %s already exists", req.Topic),
		)
	}

	// Check if we have a null admin client
	if _, isNull := s.admin.(*ConfluentAdmin); isNull {
		logger.LogWarnEvent("", "create_topic_null_admin", req.Topic, "warn",
			"Attempting to create topic with null admin client. Operation may not be performed.")
	}

	// Create the topic with retry logic
	maxRetries := 3
	var createErr error

	for i := 0; i < maxRetries; i++ {
		createErr = s.admin.CreateTopic(
			ctx,
			req.Topic,
			s.config.KafkaNumPartitions,
			s.config.KafkaReplicationFactor,
			nil, // Use default config
		)

		if createErr == nil {
			break
		}

		logger.LogWarnEvent("", "create_topic_retry", req.Topic, "warn",
			fmt.Sprintf("Failed to create topic (attempt %d/%d): %v", i+1, maxRetries, createErr))

		// Wait before retrying
		select {
		case <-ctx.Done():
			return errors.NewCustomError(
				errors.TOPErrCreateFailed,
				fmt.Errorf("Context cancelled while creating topic: %v", ctx.Err()),
			)
		case <-time.After(time.Duration(500*(i+1)) * time.Millisecond):
			// Continue with retry
		}
	}

	if createErr != nil {
		logger.LogEvent("kafka", "create_topic_failed", req.Topic, "error", createErr.Error())
		return errors.NewCustomError(
			errors.TOPErrCreateFailed,
			fmt.Errorf("Failed to create topic after %d attempts: %v", maxRetries, createErr),
		)
	}

	return nil
}

// getOrCreateProducer gets an existing producer for a topic or creates a new one
func (s *ConfluentMessagingService) getOrCreateProducer(cfg confluent.KafkaConfig) (confluent.Producer, error) {
	if cfg.Topic == "" {
		return nil, fmt.Errorf("topic is required")
	}

	// Check if we already have a producer for this topic
	s.mu.RLock()
	producer, exists := s.producers[cfg.Topic]
	s.mu.RUnlock()

	if exists {
		return producer, nil
	}

	// Create a new producer with thread safety
	s.mu.Lock()
	defer s.mu.Unlock()

	// Double-check if another goroutine created the producer while we were waiting for the lock
	if producer, exists = s.producers[cfg.Topic]; exists {
		return producer, nil
	}

	// Create a new producer
	producer, err := s.factory.CreateProducer(cfg)
	if err != nil {
		return nil, err
	}

	s.producers[cfg.Topic] = producer
	return producer, nil
}

// Close closes all resources
func (s *ConfluentMessagingService) Close() error {
	// Close all producers with proper locking
	var lastErr error

	// Get a copy of the producers map to avoid long lock
	s.mu.RLock()
	producersCopy := make(map[string]confluent.Producer)
	for topic, producer := range s.producers {
		producersCopy[topic] = producer
	}
	s.mu.RUnlock()

	// Close each producer
	for topic, producer := range producersCopy {
		if err := producer.Close(); err != nil {
			logger.LogErrorEvent("", "producer_close_error", topic, "error",
				fmt.Sprintf("Failed to close producer: %v", err))
			lastErr = err
		}

		// Remove from map
		s.mu.Lock()
		delete(s.producers, topic)
		s.mu.Unlock()
	}

	// Close admin client
	if err := s.admin.Close(); err != nil {
		logger.LogErrorEvent("", "admin_close_error", "", "error",
			fmt.Sprintf("Failed to close admin client: %v", err))
		lastErr = err
	}

	return lastErr
}

// ConfluentAdmin is a no-op implementation of KafkaAdmin that logs errors but doesn't fail
type ConfluentAdmin struct{}

func (a *ConfluentAdmin) CreateTopic(ctx context.Context, topic string, numPartitions, replicationFactor int, configs map[string]string) error {
	logger.LogWarnEvent("", "null_admin_create_topic", topic, "warn",
		"Using null admin client, topic creation may not be performed")
	return nil
}

func (a *ConfluentAdmin) DeleteTopic(ctx context.Context, topic string) error {
	logger.LogWarnEvent("", "null_admin_delete_topic", topic, "warn",
		"Using null admin client, topic deletion may not be performed")
	return nil
}

func (a *ConfluentAdmin) ListTopics(ctx context.Context) ([]string, error) {
	logger.LogWarnEvent("", "null_admin_list_topics", "", "warn",
		"Using null admin client, topic listing may not be performed")
	return nil, nil
}

func (a *ConfluentAdmin) HasActiveConsumers(ctx context.Context, topic string) (bool, error) {
	logger.LogWarnEvent("", "null_admin_check_consumers", topic, "warn",
		"Using null admin client, consumer check may not be performed")
	// In null implementation, assume there are consumers (safer)
	return true, nil
}

func (a *ConfluentAdmin) Close() error {
	return nil
}

// topicExists checks if a topic exists in Kafka
func (s *ConfluentMessagingService) topicExists(ctx context.Context, topic string) (bool, error) {
	// Get list of topics
	topics, err := s.admin.ListTopics(ctx)
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
