package handler

import (
	"context"
	pb "cra-protos/messaging_service"
	"encoding/json"
	"fmt"
	"time"

	"messaging_service/internal/config"
	"messaging_service/pkg/kafka"
	"messaging_service/pkg/logger"
)

type MessagingHandler struct {
	pb.UnimplementedMessagingServiceServer
	config *config.Config
}

// NewMessagingHandler creates a new instance of MessagingHandler
func NewMessagingHandler(config *config.Config) *MessagingHandler {
	return &MessagingHandler{
		config: config,
	}
}

// extractDeliverySemantics extracts delivery semantics from request headers
func extractDeliverySemantics(req *pb.PublishRequest) kafka.DeliverySemantics {
	return kafka.AtLeastOnce
}

func (s *MessagingHandler) PublishMessage(ctx context.Context, req *pb.PublishRequest) (*pb.PublishResponse, error) {

	brokers := s.config.KafkaBrokers
	cfg := kafka.NewDefaultKafkaConfig(brokers, req.Topic)
	cfg.DeliverySemantics = extractDeliverySemantics(req)
	cfg.GroupID = req.GroupId
	cfg.MinBytes = int(req.MinBytes)
	cfg.MaxBytes = int(req.MaxBytes)

	// Log the delivery semantics being used
	logger.LogInfo("kafka_publish_config",
		fmt.Sprintf("Publishing to topic %s with delivery semantics: %s",
			req.Topic, cfg.DeliverySemantics))

	// Create a producer with proper configuration
	producer := kafka.NewProducer(cfg)
	defer producer.Close()

	b, err := MapToBytes(req.Value)
	if err != nil {
		logger.LogErrorEvent("", "serialization_failed", "", "error", fmt.Sprintf("Serialization failed: %v", err))
		return &pb.PublishResponse{
			Status: "failed",
		}, fmt.Errorf("serialization failed: %v", err)
	}

	messageID := fmt.Sprintf("%s-%d", req.Topic, time.Now().UnixNano())

	// Publish with retry for at-least-once or exactly-once semantics
	if cfg.DeliverySemantics == kafka.AtLeastOnce || cfg.DeliverySemantics == kafka.ExactlyOnce {
		err = producer.WriteWithRetry(ctx, []byte(messageID), b, 5)
	} else {
		// Simple write for at-most-once
		err = producer.Write(ctx, []byte(messageID), b)
	}

	if err != nil {
		logger.LogErrorEvent("", "kafka_publish_failed", "", "error", fmt.Sprintf("Failed to publish message: %v", err))
		return &pb.PublishResponse{
			Status: "failed",
		}, err
	}

	logger.LogEvent(messageID, "message_published", "", "success",
		fmt.Sprintf("Message published to topic %s with delivery semantics %s",
			req.Topic, cfg.DeliverySemantics))

	return &pb.PublishResponse{
		Status: "success",
	}, nil
}

func (s *MessagingHandler) SubscribeStream(req *pb.SubscribeRequest, stream pb.MessagingService_SubscribeStreamServer) error {
	const defaultMinBytes = 1
	const defaultMaxBytes = 1048576

	brokers := s.config.KafkaBrokers
	cfg := kafka.NewDefaultKafkaConfig(brokers, req.Topic)
	cfg.GroupID = req.GroupId
	cfg.MinBytes = int(req.MinBytes)
	cfg.MaxBytes = int(req.MaxBytes)
	cfg.DeliverySemantics = kafka.AtLeastOnce

	logger.LogInfo("kafka_subscribe",
		fmt.Sprintf("Subscribing to topic %s with delivery semantics: %s",
			req.Topic, cfg.DeliverySemantics))

	if cfg.MinBytes <= 0 {
		cfg.MinBytes = defaultMinBytes
	}
	if cfg.MaxBytes <= 0 {
		cfg.MaxBytes = defaultMaxBytes
	}

	// Initialize the consumer with error handling
	consumer := kafka.NewConsumer(cfg, func(message []byte) error {
		// Try to deserialize the message
		msg, err := BytesToMap(message)
		if err != nil {
			// Return a permanent error for deserialization failures
			return kafka.NewConsumerError(
				err,
				true,
				"Failed to deserialize message",
			)
		}

		// Example of business logic validation
		// Here we can check message content and return different error types
		// based on whether we want to retry or send directly to DLQ

		// Send the message to the stream
		streamErr := stream.Send(&pb.KafkaMessage{
			Value:     msg,
			Timestamp: time.Now().UnixMilli(),
		})

		if streamErr != nil {
			// Return a retriable error for stream send failures
			return kafka.NewConsumerError(
				streamErr,
				false,
				"Failed to send message to stream",
			)
		}

		return nil
	})

	// Start the consumer
	go func() {
		consumer.Start(stream.Context())
	}()

	// Keep the stream open
	<-stream.Context().Done()

	return nil
}

func MapToBytes(m map[string]string) ([]byte, error) {
	return json.Marshal(m)
}
func BytesToMap(b []byte) (map[string]string, error) {
	var m map[string]string
	err := json.Unmarshal(b, &m)
	return m, err
}

// CreateTopic creates a new Kafka topic with the specified configuration
func (s *MessagingHandler) CreateTopic(ctx context.Context, req *pb.CreateTopicRequest) (*pb.CreateTopicResponse, error) {
	brokers := s.config.KafkaBrokers
	cfg := kafka.NewDefaultKafkaConfig(brokers, "")

	admin := kafka.NewAdmin(cfg)
	defer admin.Close()

	topicConfig := make(map[string]string)
	if req.Config != nil {
		topicConfig = req.Config
	}

	err := admin.CreateTopic(ctx, req.Topic, int(req.NumPartitions), int(req.ReplicationFactor), topicConfig)
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
