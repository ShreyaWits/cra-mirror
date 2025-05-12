package handler

import (
	"context"
	pb "cra-protos/messaging_service"
	"fmt"
	"time"

	"messaging_service/internal/config"
	"messaging_service/internal/messaging_service/service"
	"messaging_service/internal/validation"
	"messaging_service/pkg/errors"
	"messaging_service/pkg/kafka"
	"messaging_service/pkg/logger"
)

type MessagingHandler struct {
	pb.UnimplementedMessagingServiceServer
	config           *config.Config
	messagingService service.MessagingService
}

// NewMessagingHandler creates a new instance of MessagingHandler
func NewMessagingHandler(config *config.Config, messagingService service.MessagingService) *MessagingHandler {
	return &MessagingHandler{
		config:           config,
		messagingService: messagingService,
	}
}

func (s *MessagingHandler) PublishMessageV1(ctx context.Context, req *pb.PublishRequest) (*pb.PublishResponse, error) {
	// Validate request
	if customErr := validation.ValidatePublishRequest(req); customErr != nil {
		logger.LogEvent("", "validation_failed", req.Topic, "error", customErr.Error())
		return nil, errors.NewGRPCError(customErr)
	}

	// Create Kafka configuration
	cfg := kafka.NewDefaultKafkaConfig(s.config.KafkaBrokers, req.Topic)

	// Set producer configuration if provided
	if req.ProducerConfig != nil {
		cfg.ExactlyOnceConfig.EnableIdempotence = req.ProducerConfig.EnableIdempotence
		cfg.ExactlyOnceConfig.EnableTransactions = req.ProducerConfig.EnableIdempotence
		// Set delivery semantics based on the enum
		if req.ProducerConfig.DeliverySemantics == pb.DeliverySemantics_DELIVERY_SEMANTICS_EXACTLY_ONCE {
			cfg.ExactlyOnceConfig.EnableTransactions = true
		}
	}

	// Log the publish request
	logger.LogEvent("", "kafka_publish_config", req.Topic, "info", fmt.Sprintf("Publishing to topic %s", req.Topic))

	// Publish message
	customErr := s.messagingService.PublishMessage(ctx, cfg, req)
	if customErr != nil {
		logger.LogEvent("", "kafka_publish_failed", req.Topic, "error", customErr.Error())
		return nil, errors.NewGRPCError(customErr)
	}

	return &pb.PublishResponse{
		Status:  "success",
		Message: fmt.Sprintf("Message published successfully to topic %s", req.Topic),
	}, nil
}

func (s *MessagingHandler) SubscribeV1(req *pb.SubscribeRequest, stream pb.MessagingService_SubscribeV1Server) error {
	// Validate request
	if customErr := validation.ValidateSubscribeRequest(req); customErr != nil {
		logger.LogEvent("", "validation_failed", req.Topic, "error", customErr.Error())
		return errors.NewGRPCError(customErr)
	}

	// Create Kafka configuration
	cfg := kafka.NewDefaultKafkaConfig(s.config.KafkaBrokers, req.Topic)
	cfg.GroupID = req.GroupId

	// Set consumer configuration if provided
	if req.ConsumerConfig != nil {
		cfg.ConsumerConfig.MaxWait = time.Duration(req.ConsumerConfig.MaxWaitMs) * time.Millisecond
		cfg.ConsumerConfig.CommitInterval = time.Duration(req.ConsumerConfig.CommitIntervalMs) * time.Millisecond

		// Set isolation level based on enum
		switch req.ConsumerConfig.IsolationLevel {
		case pb.IsolationLevel_ISOLATION_LEVEL_READ_COMMITTED:
			cfg.ConsumerConfig.IsolationLevel = "read_committed"
		case pb.IsolationLevel_ISOLATION_LEVEL_READ_UNCOMMITTED:
			cfg.ConsumerConfig.IsolationLevel = "read_uncommitted"
		default:
			cfg.ConsumerConfig.IsolationLevel = "read_committed"
		}

		// Set auto offset reset based on enum
		switch req.ConsumerConfig.AutoOffsetReset {
		case pb.AutoOffsetReset_AUTO_OFFSET_RESET_LATEST:
			cfg.ConsumerConfig.AutoOffsetReset = "latest"
		case pb.AutoOffsetReset_AUTO_OFFSET_RESET_EARLIEST:
			cfg.ConsumerConfig.AutoOffsetReset = "earliest"
		default:
			cfg.ConsumerConfig.AutoOffsetReset = "latest"
		}
	}

	// Log the subscribe request
	logger.LogEvent("", "kafka_subscribe_config", req.Topic, "info", fmt.Sprintf("Subscribing to topic %s with group %s", req.Topic, req.GroupId))

	// Start consuming messages
	customErr := s.messagingService.ConsumeMessage(stream, cfg)
	if customErr != nil {
		logger.LogEvent("", "kafka_subscribe_failed", req.Topic, "error", customErr.Error())
		return errors.NewGRPCError(customErr)
	}

	return nil
}

func (s *MessagingHandler) CreateTopicV1(ctx context.Context, req *pb.CreateTopicRequest) (*pb.CreateTopicResponse, error) {
	// Validate request
	if customErr := validation.ValidateCreateTopicRequest(req); customErr != nil {
		logger.LogEvent("", "validation_failed", req.Topic, "error", customErr.Error())
		return nil, errors.NewGRPCError(customErr)
	}

	// Create Kafka configuration
	cfg := kafka.NewDefaultKafkaConfig(s.config.KafkaBrokers, req.Topic)

	// Log the create topic request
	logger.LogEvent("", "kafka_create_topic_config", req.Topic, "info", fmt.Sprintf("Creating topic %s with %d partitions and replication factor %d",
		req.Topic, req.NumPartitions, req.ReplicationFactor))

	// Create topic
	customErr := s.messagingService.CreateTopic(ctx, req, cfg)
	if customErr != nil {
		logger.LogEvent("", "kafka_create_topic_failed", req.Topic, "error", customErr.Error())
		return nil, errors.NewGRPCError(customErr)
	}

	return &pb.CreateTopicResponse{
		Status: "success",
		Message: fmt.Sprintf("Topic %s created successfully with %d partitions and replication factor %d",
			req.Topic, req.NumPartitions, req.ReplicationFactor),
	}, nil
}
