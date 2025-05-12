package handler

import (
	"context"
	pb "cra-protos/messaging_service"
	"fmt"
	"time"

	appconfig "messaging_service/internal/config"
	"messaging_service/internal/messaging_service/service"
	"messaging_service/internal/validation"
	"messaging_service/pkg/errors"
	"messaging_service/pkg/kafka"
	"messaging_service/pkg/logger"
)

const (
	defaultTimeout = 5 * time.Second
)

type MessagingHandler struct {
	pb.UnimplementedMessagingServiceServer
	config           *appconfig.Config
	messagingService service.MessagingService
	kafkaConfig      *kafka.Configurator
}

// NewMessagingHandler creates a new instance of MessagingHandler
func NewMessagingHandler(config *appconfig.Config, messagingService service.MessagingService) *MessagingHandler {
	return &MessagingHandler{
		config:           config,
		messagingService: messagingService,
		kafkaConfig:      kafka.NewConfigurator(config.KafkaBrokers),
	}
}

func (s *MessagingHandler) PublishMessageV1(ctx context.Context, req *pb.PublishRequest) (*pb.PublishResponse, error) {
	// Create timeout context
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	// Get request ID from context or generate new one
	requestID := getRequestID(ctx)

	// Validate request
	if customErr := validation.ValidatePublishRequest(req); customErr != nil {
		logger.LogEvent(requestID, "validation_failed", req.Topic, "error",
			fmt.Sprintf("Publish validation failed: %v", customErr))
		return nil, errors.NewGRPCError(customErr)
	}

	// Create Kafka configuration using the configurator
	cfg := s.kafkaConfig.CreatePublishConfig(req.Topic, req)

	// Log the publish request
	logger.LogEvent(requestID, "kafka_publish_start", req.Topic, "info",
		fmt.Sprintf("Publishing to topic %s with config: idempotence=%v, transactions=%v",
			req.Topic, cfg.ExactlyOnceConfig.EnableIdempotence, cfg.ExactlyOnceConfig.EnableTransactions))

	// Publish message
	customErr := s.messagingService.PublishMessage(ctx, cfg, req)
	if customErr != nil {
		logger.LogEvent(requestID, "kafka_publish_failed", req.Topic, "error",
			fmt.Sprintf("Failed to publish message: %v", customErr))
		return nil, errors.NewGRPCError(customErr)
	}

	logger.LogEvent(requestID, "kafka_publish_success", req.Topic, "info",
		fmt.Sprintf("Successfully published message to topic %s", req.Topic))

	return &pb.PublishResponse{
		Status:  "success",
		Message: fmt.Sprintf("Message published successfully to topic %s", req.Topic),
	}, nil
}

func (s *MessagingHandler) SubscribeV1(req *pb.SubscribeRequest, stream pb.MessagingService_SubscribeV1Server) error {
	// Get request ID from stream context or generate new one
	requestID := getRequestID(stream.Context())

	// Validate request
	if customErr := validation.ValidateSubscribeRequest(req); customErr != nil {
		logger.LogEvent(requestID, "validation_failed", req.Topic, "error",
			fmt.Sprintf("Subscribe validation failed: %v", customErr))
		return errors.NewGRPCError(customErr)
	}

	// Create Kafka configuration using the configurator
	cfg := s.kafkaConfig.CreateSubscribeConfig(req.Topic, req.GroupId, req)

	// Log the subscribe request
	logger.LogEvent(requestID, "kafka_subscribe_start", req.Topic, "info",
		fmt.Sprintf("Subscribing to topic %s with group %s, config: max_wait=%v, commit_interval=%v, isolation=%v, offset=%v",
			req.Topic, req.GroupId, cfg.ConsumerConfig.MaxWait, cfg.ConsumerConfig.CommitInterval,
			cfg.ConsumerConfig.IsolationLevel, cfg.ConsumerConfig.AutoOffsetReset))

	// Start consuming messages
	customErr := s.messagingService.ConsumeMessage(stream, cfg)
	if customErr != nil {
		logger.LogEvent(requestID, "kafka_subscribe_failed", req.Topic, "error",
			fmt.Sprintf("Failed to subscribe: %v", customErr))
		return errors.NewGRPCError(customErr)
	}

	logger.LogEvent(requestID, "kafka_subscribe_success", req.Topic, "info",
		fmt.Sprintf("Successfully subscribed to topic %s", req.Topic))
	return nil
}

func (s *MessagingHandler) CreateTopicV1(ctx context.Context, req *pb.CreateTopicRequest) (*pb.CreateTopicResponse, error) {
	// Create timeout context
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	// Get request ID from context or generate new one
	requestID := getRequestID(ctx)

	// Validate request
	if customErr := validation.ValidateCreateTopicRequest(req); customErr != nil {
		logger.LogEvent(requestID, "validation_failed", req.Topic, "error",
			fmt.Sprintf("Topic creation validation failed: %v", customErr))
		return nil, errors.NewGRPCError(customErr)
	}

	// Create Kafka configuration using the configurator
	cfg := s.kafkaConfig.CreateTopicConfig(req.Topic, req)

	// Log the create topic request
	logger.LogEvent(requestID, "kafka_create_topic_start", req.Topic, "info",
		fmt.Sprintf("Creating topic %s with %d partitions and replication factor %d",
			req.Topic, req.NumPartitions, req.ReplicationFactor))

	// Create topic
	customErr := s.messagingService.CreateTopic(ctx, req, cfg)
	if customErr != nil {
		logger.LogEvent(requestID, "kafka_create_topic_failed", req.Topic, "error",
			fmt.Sprintf("Failed to create topic: %v", customErr))
		return nil, errors.NewGRPCError(customErr)
	}

	logger.LogEvent(requestID, "kafka_create_topic_success", req.Topic, "info",
		fmt.Sprintf("Successfully created topic %s", req.Topic))

	return &pb.CreateTopicResponse{
		Status: "success",
		Message: fmt.Sprintf("Topic %s created successfully with %d partitions and replication factor %d",
			req.Topic, req.NumPartitions, req.ReplicationFactor),
	}, nil
}

// Helper function to get request ID from context
func getRequestID(ctx context.Context) string {
	// TODO: Implement proper request ID extraction from context
	// For now, return a timestamp-based ID
	return fmt.Sprintf("req-%d", time.Now().UnixNano())
}
