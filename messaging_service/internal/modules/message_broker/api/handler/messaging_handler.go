package handler

import (
	"context"
	pb "cra-protos/messaging_service"
	"fmt"
	"time"

	appconfig "messaging_service/internal/config"
	"messaging_service/internal/modules/message_broker/service"
	"messaging_service/internal/utils/validation"
	"messaging_service/pkg/confluent"
	"messaging_service/pkg/errors"
	"messaging_service/pkg/observability"
)

const (
	defaultTimeout = 5 * time.Second
)

type MessagingHandler struct {
	pb.UnimplementedMessagingServiceServer
	config           *appconfig.Config
	messagingService service.MessagingService
	kafkaConfig      *confluent.Configurator
	obs              *observability.ObservabilityStack
}

// NewMessagingHandler creates a new instance of MessagingHandler
func NewMessagingHandler(config *appconfig.Config, messagingService service.MessagingService, obs *observability.ObservabilityStack) (*MessagingHandler, error) {
	// Validate all arguments
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil in NewMessagingHandler")
	}
	if messagingService == nil {
		return nil, fmt.Errorf("messagingService cannot be nil in NewMessagingHandler")
	}
	if obs == nil {
		return nil, fmt.Errorf("observability stack cannot be nil in NewMessagingHandler")
	}

	// Validate kafka brokers
	if config.KafkaBrokers == nil || len(config.KafkaBrokers) == 0 {
		return nil, fmt.Errorf("kafka brokers must be configured in NewMessagingHandler")
	}

	// Initialize the configurator
	kafkaConfig := confluent.NewConfigurator(config.KafkaBrokers, config)
	if kafkaConfig == nil {
		return nil, fmt.Errorf("failed to create Kafka configurator")
	}

	return &MessagingHandler{
		config:           config,
		messagingService: messagingService,
		kafkaConfig:      kafkaConfig,
		obs:              obs,
	}, nil
}

func (s *MessagingHandler) PublishMessageV1(ctx context.Context, req *pb.PublishRequest) (*pb.PublishResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()
	// Get request ID from context or generate new one
	// requestID := getRequestID(ctx)
	functionName := "PublishMessageV1"
	functionFailed := "PublishMessageV1_Failed"

	s.obs.MetricsService.IncrementCounter(ctx, functionName, 1, map[string]string{})

	// Create timeout context

	tCtx, span := s.obs.TracerService.StartTracer(ctx, functionName)
	defer s.obs.TracerService.StopSpan(span)
	// Validate request
	if customErr := validation.ValidatePublishRequest(req); customErr != nil {
		loggerdata := fmt.Sprintf("Publish validation failed: %v", customErr.Error())
		s.obs.LoggerService.Error(tCtx, loggerdata)
		s.obs.MetricsService.IncrementCounter(ctx, functionFailed, 1, map[string]string{"error": loggerdata})
		return nil, errors.NewGRPCError(customErr)
	}

	// Create Confluent Kafka configuration with exactly-once semantics
	cfg := s.kafkaConfig.CreatePublishConfig(req.Topic, req)

	// Publish message
	customErr := s.messagingService.PublishMessage(tCtx, cfg, req)
	if customErr != nil {
		loggerdata := fmt.Sprintf("Failed to publish message: %v", customErr.Error())
		s.obs.LoggerService.Error(tCtx, loggerdata)
		s.obs.MetricsService.IncrementCounter(tCtx, functionFailed, 1, map[string]string{"error": loggerdata})
		return nil, errors.NewGRPCError(customErr)
	}

	return &pb.PublishResponse{
		Status:  "success",
		Message: fmt.Sprintf("Message published successfully to topic %s", req.Topic),
	}, nil
}

func (s *MessagingHandler) SubscribeV1(req *pb.SubscribeRequest, stream pb.MessagingService_SubscribeV1Server) error {
	functionName := "SubscribeV1"
	functionFailed := "SubscribeV1_Failed"

	s.obs.MetricsService.IncrementCounter(stream.Context(), functionName, 1, map[string]string{})

	// Create timeout context
	ctx, cancel := context.WithTimeout(stream.Context(), defaultTimeout)
	defer cancel()

	tCtx, span := s.obs.TracerService.StartTracer(ctx, functionName)
	defer s.obs.TracerService.StopSpan(span)

	// Validate request
	if customErr := validation.ValidateSubscribeRequest(req); customErr != nil {
		loggerdata := fmt.Sprintf("Subscribe validation failed: %v", customErr.Error())
		s.obs.LoggerService.Error(tCtx, loggerdata)
		s.obs.MetricsService.IncrementCounter(tCtx, functionFailed, 1, map[string]string{"error": loggerdata})
		return errors.NewGRPCError(customErr)
	}

	// Create Confluent Kafka configuration with read committed and latest offset
	cfg := s.kafkaConfig.CreateSubscribeConfig(req.Topic, req.GroupId, req)

	// Log the subscribe request
	s.obs.LoggerService.Info(tCtx, fmt.Sprintf("Subscribing to topic %s with group %s (read committed, latest offset)",
		req.Topic, req.GroupId))

	// Log successful subscription before starting consumer
	s.obs.LoggerService.Info(tCtx, fmt.Sprintf("Successfully subscribed to topic %s", req.Topic))

	// Start consuming messages
	customErr := s.messagingService.ConsumeMessage(stream, cfg)
	if customErr != nil {
		loggerdata := fmt.Sprintf("Failed to subscribe: %v", customErr.Error())
		s.obs.LoggerService.Error(tCtx, loggerdata)
		s.obs.MetricsService.IncrementCounter(tCtx, functionFailed, 1, map[string]string{"error": loggerdata})
		return errors.NewGRPCError(customErr)
	}

	return nil
}

func (s *MessagingHandler) CreateTopicV1(ctx context.Context, req *pb.CreateTopicRequest) (*pb.CreateTopicResponse, error) {
	// Create timeout context
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	functionName := "CreateTopicV1"
	functionFailed := "CreateTopicV1_Failed"

	s.obs.MetricsService.IncrementCounter(ctx, functionName, 1, map[string]string{})

	tCtx, span := s.obs.TracerService.StartTracer(ctx, functionName)
	defer s.obs.TracerService.StopSpan(span)

	// Validate request
	if customErr := validation.ValidateCreateTopicRequest(req); customErr != nil {
		loggerdata := fmt.Sprintf("Topic creation validation failed: %v", customErr.Error())
		s.obs.LoggerService.Error(tCtx, loggerdata)
		s.obs.MetricsService.IncrementCounter(tCtx, functionFailed, 1, map[string]string{"error": loggerdata})
		return nil, errors.NewGRPCError(customErr)
	}

	// Create Confluent Kafka configuration with default values
	cfg := s.kafkaConfig.CreateTopicConfig(req.Topic, req)

	// Log the create topic request using configured values
	s.obs.LoggerService.Info(tCtx, fmt.Sprintf("Creating topic %s with %d partitions and replication factor %d",
		req.Topic, s.config.KafkaNumPartitions, s.config.KafkaReplicationFactor))

	// Create topic
	customErr := s.messagingService.CreateTopic(tCtx, req, cfg)
	if customErr != nil {
		// Check if this is a "topic already exists" error, which is not actually an error
		if customErr.ErrorCode == errors.TOPErrTopicExists {
			s.obs.LoggerService.Info(tCtx, fmt.Sprintf("Topic %s already exists", req.Topic))

			// Return success response with a message indicating the topic already exists
			return &pb.CreateTopicResponse{
				Status:  "success",
				Message: fmt.Sprintf("Topic %s already exists", req.Topic),
			}, nil
		}

		// For any other error, handle as before
		loggerdata := fmt.Sprintf("Failed to create topic: %v", customErr.Error())
		s.obs.LoggerService.Error(tCtx, loggerdata)
		s.obs.MetricsService.IncrementCounter(tCtx, functionFailed, 1, map[string]string{"error": loggerdata})
		return nil, errors.NewGRPCError(customErr)
	}

	s.obs.LoggerService.Info(tCtx, fmt.Sprintf("Successfully created topic %s", req.Topic))

	return &pb.CreateTopicResponse{
		Status: "success",
		Message: fmt.Sprintf("Topic %s created successfully with %d partitions and replication factor %d",
			req.Topic, s.config.KafkaNumPartitions, s.config.KafkaReplicationFactor),
	}, nil
}

// Helper function to get request ID from context
func getRequestID(ctx context.Context) string {
	// Check if request ID is in the context
	if id, ok := ctx.Value("request_id").(string); ok && id != "" {
		return id
	}

	// Fallback to timestamp-based ID if not found
	return fmt.Sprintf("req-%d", time.Now().UnixNano())
}

// UpdateMessagingService updates the messaging service with a new instance
func (h *MessagingHandler) UpdateMessagingService(config *appconfig.Config, messagingService service.MessagingService) error {
	// Validate inputs
	if config == nil {
		return fmt.Errorf("config cannot be nil when updating messaging service")
	}
	if messagingService == nil {
		return fmt.Errorf("messagingService cannot be nil when updating messaging service")
	}

	// Validate kafka brokers
	if config.KafkaBrokers == nil || len(config.KafkaBrokers) == 0 {
		return fmt.Errorf("kafka brokers must be configured")
	}

	// Create new configurator
	kafkaConfig, err := confluent.SafeNewConfigurator(config.KafkaBrokers, config)
	if err != nil {
		h.obs.LoggerService.Error(context.Background(), "Failed to create Kafka configurator: %v", err)
		return err
	}

	// Update handler fields
	h.config = config
	h.messagingService = messagingService
	h.kafkaConfig = kafkaConfig

	h.obs.LoggerService.Info(context.Background(), "MessagingHandler updated with new configuration")
	return nil
}
