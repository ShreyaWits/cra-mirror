package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	pb "cra-protos/messaging_service"
	kafkapkg "messaging_service/pkg/kafka"
	"messaging_service/pkg/logger"

	kafka "github.com/segmentio/kafka-go"
)

type MessagingService interface {
	PublishMessage(ctx context.Context, cfg kafkapkg.KafkaConfig, req *pb.PublishRequest) (*pb.PublishResponse, error)
	ConsumeMessage(stream pb.MessagingService_SubscribeV1Server, cfg kafkapkg.KafkaConfig)
	CreateTopic(ctx context.Context, req *pb.CreateTopicRequest, cfg kafkapkg.KafkaConfig) (*pb.CreateTopicResponse, error)
}

type MessagingServiceImpl struct {
	producer *kafkapkg.Producer
	admin    kafkapkg.KafkaAdmin
}

func NewMessagingService(producer *kafkapkg.Producer, admin kafkapkg.KafkaAdmin) MessagingService {
	return &MessagingServiceImpl{
		producer: producer,
		admin:    admin,
	}
}

func (s *MessagingServiceImpl) PublishMessage(ctx context.Context, cfg kafkapkg.KafkaConfig, req *pb.PublishRequest) (res *pb.PublishResponse, err error) {
	// Validate required fields
	if req.Topic == "" {
		return nil, fmt.Errorf("topic is required")
	}
	if req.Value == nil {
		return nil, fmt.Errorf("message content is required")
	}

	// Serialize the message value
	valueBytes, err := json.Marshal(req.Value)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize message: %w", err)
	}

	// Get or create writer
	writer := s.producer.GetWriter(cfg)
	defer func() {
		if err := s.producer.Close(writer); err != nil {
			logger.LogErrorEvent("", "kafka_close_writer_failed", req.Topic, "error", fmt.Sprintf("Failed to close writer: %v", err))
		}
	}()
	// Convert headers
	headers := make([]kafka.Header, 0, len(req.Headers))
	for k, v := range req.Headers {
		headers = append(headers, kafka.Header{
			Key:   k,
			Value: []byte(v),
		})
	}

	// Generate message key if not provided
	key := []byte(req.Key)
	if len(key) == 0 {
		key = []byte(fmt.Sprintf("%s-%d", req.Topic, time.Now().UnixNano()))
	}

	// Publish message with retry
	err = s.producer.WriteWithRetry(ctx, writer, key, valueBytes, headers, 3)
	if err != nil {
		logger.LogErrorEvent("", "kafka_publish_failed", req.Topic, "error", fmt.Sprintf("Failed to publish message: %v", err))
		return &pb.PublishResponse{
			Status:  "failed",
			Message: fmt.Sprintf("Failed to publish message: %v", err),
		}, err
	}

	return &pb.PublishResponse{
		Status:  "success",
		Message: "Message published successfully",
	}, nil
}

func (s *MessagingServiceImpl) ConsumeMessage(stream pb.MessagingService_SubscribeV1Server, cfg kafkapkg.KafkaConfig) {
	// Create a buffered channel for messages
	msgChan := make(chan *pb.KafkaMessage, 1000)

	// Start a goroutine to handle message sending
	go func() {
		for msg := range msgChan {
			if err := stream.Send(msg); err != nil {
				logger.LogErrorEvent("", "stream_send_failed", cfg.Topic, "error", err.Error())
				return
			}
		}
	}()

	// Create consumer with handler
	consumer := kafkapkg.NewConsumer(cfg, func(payload []byte) error {
		// Deserialize the message
		var value map[string]string
		if err := json.Unmarshal(payload, &value); err != nil {
			return fmt.Errorf("failed to deserialize message: %w", err)
		}

		// Create message
		msg := &pb.KafkaMessage{
			Value:     value,
			Timestamp: time.Now().UnixMilli(),
		}

		// Send message to channel
		select {
		case msgChan <- msg:
		default:
			// If channel is full, try to send directly
			if err := stream.Send(msg); err != nil {
				return fmt.Errorf("failed to send message to stream: %w", err)
			}
		}
		return nil
	})

	// Start consuming
	consumer.Start(stream.Context())
	close(msgChan)
}

func (s *MessagingServiceImpl) CreateTopic(ctx context.Context, req *pb.CreateTopicRequest, cfg kafkapkg.KafkaConfig) (*pb.CreateTopicResponse, error) {
	// Validate required fields
	if req.Topic == "" {
		return nil, fmt.Errorf("topic name is required")
	}

	// Set default values if not provided
	if req.NumPartitions <= 0 {
		req.NumPartitions = 1
	}
	if req.ReplicationFactor <= 0 {
		req.ReplicationFactor = 1
	}

	// Initialize config if nil
	if req.Config == nil {
		req.Config = &pb.TopicConfig{}
	}

	// Set default retention settings if not provided
	if req.Config.RetentionMs == 0 {
		req.Config.RetentionMs = 3000
	}
	if req.Config.CleanupPolicy == pb.CleanupPolicy_CLEANUP_POLICY_UNSPECIFIED {
		req.Config.CleanupPolicy = pb.CleanupPolicy_CLEANUP_POLICY_DELETE
	}

	// Convert TopicConfig to map for Kafka admin
	kafkaConfig := map[string]string{
		"retention.ms":                    fmt.Sprintf("%d", req.Config.RetentionMs),
		"cleanup.policy":                  getCleanupPolicyString(req.Config.CleanupPolicy),
		"delete.retention.ms":             "1000",
		"segment.ms":                      "1000",
		"log.retention.check.interval.ms": "1000",
		"log.cleanup.interval.mins":       "1",
		"log.segment.delete.delay.ms":     "1000",
	}

	// Create topic
	err := s.admin.CreateTopic(ctx, req.Topic, int(req.NumPartitions), int(req.ReplicationFactor), kafkaConfig)
	if err != nil {
		logger.LogErrorEvent("kafka", "create_topic_failed", req.Topic, "error", err.Error())
		return &pb.CreateTopicResponse{
			Status:  "failed",
			Message: fmt.Sprintf("Failed to create topic: %v", err),
		}, err
	}

	return &pb.CreateTopicResponse{
		Status:  "success",
		Message: fmt.Sprintf("Topic %s created successfully", req.Topic),
	}, nil
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
