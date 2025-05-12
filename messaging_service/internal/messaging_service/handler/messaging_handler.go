package handler

import (
	"context"
	pb "cra-protos/messaging_service"
	"fmt"
	"time"

	"messaging_service/internal/config"
	"messaging_service/internal/messaging_service/service"
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

func (s *MessagingHandler) PublishMessage(ctx context.Context, req *pb.PublishRequest) (*pb.PublishResponse, error) {
	// Create Kafka configuration
	cfg := kafka.NewDefaultKafkaConfig(s.config.KafkaBrokers, req.Topic)

	// Set producer configuration if provided
	if req.ProducerConfig != nil {
		cfg.ExactlyOnceConfig.EnableIdempotence = req.ProducerConfig.EnableIdempotence
		cfg.ExactlyOnceConfig.EnableTransactions = req.ProducerConfig.EnableIdempotence
		cfg.ExactlyOnceConfig.TransactionTimeoutMs = int(req.ProducerConfig.RequestTimeoutMs)
	}

	// Log the publish request
	logger.LogInfo("kafka_publish_config",
		fmt.Sprintf("Publishing to topic %s", req.Topic))

	return s.messagingService.PublishMessage(ctx, cfg, req)
}

func (s *MessagingHandler) Subscribe(req *pb.SubscribeRequest, stream pb.MessagingService_SubscribeServer) error {
	// Create Kafka configuration
	cfg := kafka.NewDefaultKafkaConfig(s.config.KafkaBrokers, req.Topic)

	// Set consumer configuration
	cfg.GroupID = req.GroupId

	// Set consumer configuration if provided
	if req.ConsumerConfig != nil {
		cfg.ConsumerConfig.MaxWait = time.Duration(req.ConsumerConfig.MaxWaitMs) * time.Millisecond
		cfg.ConsumerConfig.MaxAttempts = int(req.ConsumerConfig.MaxAttempts)
		cfg.ConsumerConfig.IsolationLevel = req.ConsumerConfig.IsolationLevel
		cfg.ConsumerConfig.CommitInterval = time.Duration(req.ConsumerConfig.CommitIntervalMs) * time.Millisecond
	}

	logger.LogInfo("kafka_subscribe",
		fmt.Sprintf("Subscribing to topic %s", req.Topic))

	s.messagingService.ConsumeMessage(stream, cfg)

	// Keep the stream open
	<-stream.Context().Done()
	return nil
}

func (s *MessagingHandler) CreateTopic(ctx context.Context, req *pb.CreateTopicRequest) (*pb.CreateTopicResponse, error) {
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
