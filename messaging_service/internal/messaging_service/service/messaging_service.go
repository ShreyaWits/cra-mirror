package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	pb "cra-protos/messaging_service"
	errors "messaging_service/pkg/errors"
	kafkapkg "messaging_service/pkg/kafka"
	"messaging_service/pkg/logger"
)

type MessagingService interface {
	PublishMessage(ctx context.Context, cfg kafkapkg.KafkaConfig, req *pb.PublishRequest) *errors.CustomError
	ConsumeMessage(stream pb.MessagingService_SubscribeV1Server, cfg kafkapkg.KafkaConfig) *errors.CustomError
	CreateTopic(ctx context.Context, req *pb.CreateTopicRequest, cfg kafkapkg.KafkaConfig) *errors.CustomError
}

type MessagingServiceImpl struct {
	producerFactory ProducerFactory
	admin           kafkapkg.KafkaAdmin
	// Map to store topic-specific producers
	producers map[string]kafkapkg.Producer
}

// ProducerFactory defines an interface for creating producers
type ProducerFactory interface {
	CreateProducer(cfg kafkapkg.KafkaConfig) kafkapkg.Producer
}

// DefaultProducerFactory implements ProducerFactory
type DefaultProducerFactory struct{}

// CreateProducer creates a new producer
func (f *DefaultProducerFactory) CreateProducer(cfg kafkapkg.KafkaConfig) kafkapkg.Producer {
	return kafkapkg.NewProducer(cfg)
}

func NewMessagingService(admin kafkapkg.KafkaAdmin) MessagingService {
	return &MessagingServiceImpl{
		producerFactory: &DefaultProducerFactory{},
		admin:           admin,
		producers:       make(map[string]kafkapkg.Producer),
	}
}

// getOrCreateProducer gets an existing producer for a topic or creates a new one
func (s *MessagingServiceImpl) getOrCreateProducer(cfg kafkapkg.KafkaConfig) (kafkapkg.Producer, error) {
	if cfg.Topic == "" {
		return nil, fmt.Errorf("topic is required")
	}

	// Check if we already have a producer for this topic
	if producer, exists := s.producers[cfg.Topic]; exists {
		return producer, nil
	}

	// Create a new producer
	producer := s.producerFactory.CreateProducer(cfg)
	s.producers[cfg.Topic] = producer
	return producer, nil
}

func (s *MessagingServiceImpl) PublishMessage(ctx context.Context, cfg kafkapkg.KafkaConfig, req *pb.PublishRequest) *errors.CustomError {
	if req.Topic == "" {
		return errors.NewCustomError(errors.MSGErrInvalidTopic, fmt.Errorf("topic is required"))
	}
	if req.Value == nil || len(req.Value) == 0 {
		return errors.NewCustomError(errors.PUBErrInvalidMessage, fmt.Errorf("message content is required"))
	}

	// Ensure topic in config matches request
	cfg.Topic = req.Topic

	valueBytes, err := json.Marshal(req.Value)
	if err != nil {
		return errors.NewCustomError(errors.PUBErrInvalidMessage, fmt.Errorf("failed to serialize message: %w", err))
	}

	// Get or create a producer for this topic
	producer, err := s.getOrCreateProducer(cfg)
	if err != nil {
		return errors.NewCustomError(errors.PUBErrProducerNotReady,
			fmt.Errorf("failed to get producer: %w", err))
	}

	key := []byte(req.Key)
	if len(key) == 0 {
		key = []byte(fmt.Sprintf("%s-%d", req.Topic, time.Now().UnixNano()))
	}

	// Use WriteWithRetry for better reliability
	err = producer.WriteWithRetry(ctx, key, valueBytes, nil)
	if err != nil {
		logger.LogEvent("", "kafka_publish_failed", req.Topic, "error",
			fmt.Sprintf("Failed to publish message: %v", err))
		return errors.NewCustomError(errors.PUBErrPublishFailed,
			fmt.Errorf("failed to publish message: %w", err))
	}

	return nil
}

func (s *MessagingServiceImpl) ConsumeMessage(stream pb.MessagingService_SubscribeV1Server, cfg kafkapkg.KafkaConfig) *errors.CustomError {
	msgChan := make(chan *pb.KafkaMessage, 1000)

	go func() {
		for msg := range msgChan {
			if err := stream.Send(msg); err != nil {
				logger.LogEvent("", "stream_send_failed", cfg.Topic, "error", err.Error())
				// Not returning error here, as this is a goroutine
			}
		}
	}()

	consumer := kafkapkg.NewConsumer(cfg, func(payload []byte) error {
		var value map[string]string
		if err := json.Unmarshal(payload, &value); err != nil {
			return fmt.Errorf("failed to deserialize message: %w", err)
		}

		// Create a new KafkaMessage with the updated fields from the protobuf
		msg := &pb.KafkaMessage{
			Value:     value,
			Timestamp: time.Now().UnixMilli(),
			Key:       "",                      // Default to empty key
			Headers:   make(map[string]string), // Initialize empty headers
		}

		select {
		case msgChan <- msg:
		default:
			if err := stream.Send(msg); err != nil {
				return fmt.Errorf("failed to send message to stream: %w", err)
			}
		}
		return nil
	})

	consumer.Start(stream.Context())
	close(msgChan)
	return nil
}

func (s *MessagingServiceImpl) CreateTopic(ctx context.Context, req *pb.CreateTopicRequest, cfg kafkapkg.KafkaConfig) *errors.CustomError {
	if req.Topic == "" {
		return errors.NewCustomError(errors.MSGErrInvalidTopic, fmt.Errorf("topic name is required"))
	}

	// Use configuration from KafkaConfig passed from higher layers
	// which now contains values from environment variables
	kafkaConfig := map[string]string{
		"retention.ms":                    fmt.Sprintf("%d", cfg.ConsumerConfig.RetentionTime.Milliseconds()),
		"cleanup.policy":                  "delete", // Maps to CLEANUP_POLICY_DELETE from proto
		"delete.retention.ms":             "1000",
		"segment.ms":                      "1000",
		"log.retention.check.interval.ms": "1000",
		"log.cleanup.interval.mins":       "1",
		"log.segment.delete.delay.ms":     "1000",
	}

	err := s.admin.CreateTopic(ctx, req.Topic, cfg.NumPartitions, cfg.ReplicationFactor, kafkaConfig)
	if err != nil {
		logger.LogEvent("kafka", "create_topic_failed", req.Topic, "error", err.Error())
		return errors.NewCustomError(errors.TOPErrCreateFailed, fmt.Errorf("failed to create topic: %w", err))
	}

	return nil
}

// Close closes all producers and frees resources
func (s *MessagingServiceImpl) Close() error {
	var lastErr error
	for topic, producer := range s.producers {
		if err := producer.Close(); err != nil {
			logger.LogErrorEvent("", "producer_close_error", topic, "error",
				fmt.Sprintf("Failed to close producer: %v", err))
			lastErr = err
		}
	}
	return lastErr
}

// Helper function to convert CleanupPolicy enum to string
func getCleanupPolicyString(policy pb.CleanupPolicy) string {
	switch policy {
	case pb.CleanupPolicy_CLEANUP_POLICY_DELETE:
		return "delete"
	case pb.CleanupPolicy_CLEANUP_POLICY_COMPACT:
		return "compact"
	default:
		return "delete"
	}
}
