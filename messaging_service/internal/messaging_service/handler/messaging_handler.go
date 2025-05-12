package handler

import (
	"context"
	pb "cra-protos/messaging_service"
	"fmt"
	"time"

	"messaging_service/internal/config"
	"messaging_service/internal/messaging_service/service"
	"messaging_service/internal/validation"
	"messaging_service/pkg/kafka"
	"messaging_service/pkg/logger"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
	if err := validation.ValidatePublishRequest(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
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
	logger.LogInfo("kafka_publish_config",
		fmt.Sprintf("Publishing to topic %s", req.Topic))

	return s.messagingService.PublishMessage(ctx, cfg, req)
}

func (s *MessagingHandler) SubscribeV1(req *pb.SubscribeRequest, stream pb.MessagingService_SubscribeV1Server) error {
	if err := validation.ValidateSubscribeRequest(req); err != nil {
		return status.Error(codes.InvalidArgument, err.Error())
	}

	// Create Kafka configuration
	cfg := kafka.NewDefaultKafkaConfig(s.config.KafkaBrokers, req.Topic)

	// Set consumer configuration
	cfg.GroupID = req.GroupId

	// Set consumer configuration if provided
	if req.ConsumerConfig != nil {
		cfg.ConsumerConfig.MaxWait = time.Duration(req.ConsumerConfig.MaxWaitMs) * time.Millisecond
		cfg.ConsumerConfig.CommitInterval = time.Duration(req.ConsumerConfig.CommitIntervalMs) * time.Millisecond

		// Convert isolation level enum to string
		switch req.ConsumerConfig.IsolationLevel {
		case pb.IsolationLevel_ISOLATION_LEVEL_READ_COMMITTED:
			cfg.ConsumerConfig.IsolationLevel = "read_committed"
		case pb.IsolationLevel_ISOLATION_LEVEL_READ_UNCOMMITTED:
			cfg.ConsumerConfig.IsolationLevel = "read_uncommitted"
		default:
			cfg.ConsumerConfig.IsolationLevel = "read_uncommitted"
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

	logger.LogInfo("kafka_subscribe",
		fmt.Sprintf("Subscribing to topic %s", req.Topic))

	s.messagingService.ConsumeMessage(stream, cfg)

	// Keep the stream open
	<-stream.Context().Done()
	return nil
}

func (s *MessagingHandler) CreateTopicV1(ctx context.Context, req *pb.CreateTopicRequest) (*pb.CreateTopicResponse, error) {
	if err := validation.ValidateCreateTopicRequest(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Create Kafka configuration
	cfg := kafka.NewDefaultKafkaConfig(s.config.KafkaBrokers, req.Topic)

	// Set default values if not provided
	if req.NumPartitions <= 0 {
		req.NumPartitions = 3
	}
	if req.ReplicationFactor <= 0 {
		req.ReplicationFactor = 3
	}

	return s.messagingService.CreateTopic(ctx, req, cfg)
}
