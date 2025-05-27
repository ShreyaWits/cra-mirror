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

	// Metric names
	metricPublishTotal         = "publish_message_total"
	metricPublishSuccess       = "publish_message_success"
	metricPublishFailure       = "publish_message_failure"
	metricSubscribeTotal       = "subscribe_total"
	metricSubscribeSuccess     = "subscribe_success"
	metricSubscribeFailure     = "subscribe_failure"
	metricCreateTopicTotal     = "create_topic_total"
	metricCreateTopicSuccess   = "create_topic_success"
	metricCreateTopicFailure   = "create_topic_failure"
	metricUpdateHandlerTotal   = "update_handler_total"
	metricUpdateHandlerSuccess = "update_handler_success"
	metricUpdateHandlerFailure = "update_handler_failure"

	// Error types
	errorTypePublishValidation   = "publish_validation_error"
	errorTypeSubscribeValidation = "subscribe_validation_error"
	errorTypeTopicValidation     = "topic_validation_error"
	errorTypePublish             = "publish_error"
	errorTypeSubscribe           = "subscribe_error"
	errorTypeCreateTopic         = "create_topic_error"
	errorTypeConfig              = "configuration_error"
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
	functionName := "PublishMessageV1"

	// Create timeout context
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	// Start tracing
	tCtx, span := s.obs.TracerService.StartTracer(ctx, functionName)
	defer s.obs.TracerService.StopSpan(span)

	// Generate request ID for correlation
	requestID := fmt.Sprintf("req-%d", time.Now().UnixNano())

	// Set tracing attributes
	s.obs.TracerService.SetAttributes(span, map[string]string{
		"request_id": requestID,
		"operation":  functionName,
		"topic":      req.Topic, // Topic name is not sensitive
	})

	// Increment total requests metric
	s.obs.MetricsService.IncrementCounter(tCtx, metricPublishTotal, 1, nil)

	// Validate request
	if customErr := validation.ValidatePublishRequest(req); customErr != nil {
		s.obs.LoggerService.Error(tCtx, fmt.Sprintf("[%s] Request validation failed", requestID))
		s.obs.MetricsService.IncrementCounter(tCtx, metricPublishFailure, 1, map[string]string{
			"error_type": errorTypePublishValidation,
		})
		return nil, errors.NewGRPCError(customErr)
	}

	s.obs.LoggerService.Info(tCtx, fmt.Sprintf("[%s] Publishing message to topic: %s", requestID, req.Topic))

	// Create Confluent Kafka configuration with exactly-once semantics
	cfg := s.kafkaConfig.CreatePublishConfig(req.Topic, req)

	// Publish message
	customErr := s.messagingService.PublishMessage(tCtx, cfg, req)
	if customErr != nil {
		s.obs.LoggerService.Error(tCtx, fmt.Sprintf("[%s] Failed to publish message to topic: %s", requestID, req.Topic))
		s.obs.MetricsService.IncrementCounter(tCtx, metricPublishFailure, 1, map[string]string{
			"error_type": errorTypePublish,
		})
		return nil, errors.NewGRPCError(customErr)
	}

	s.obs.LoggerService.Info(tCtx, fmt.Sprintf("[%s] Successfully published message to topic: %s", requestID, req.Topic))
	s.obs.MetricsService.IncrementCounter(tCtx, metricPublishSuccess, 1, nil)

	return &pb.PublishResponse{
		Status:  "success",
		Message: fmt.Sprintf("Message published successfully to topic %s", req.Topic),
	}, nil
}

func (s *MessagingHandler) SubscribeV1(req *pb.SubscribeRequest, stream pb.MessagingService_SubscribeV1Server) error {
	functionName := "SubscribeV1"

	// Create timeout context
	ctx, cancel := context.WithTimeout(stream.Context(), defaultTimeout)
	defer cancel()

	// Start tracing
	tCtx, span := s.obs.TracerService.StartTracer(ctx, functionName)
	defer s.obs.TracerService.StopSpan(span)

	// Generate request ID for correlation
	requestID := fmt.Sprintf("req-%d", time.Now().UnixNano())

	// Set tracing attributes
	s.obs.TracerService.SetAttributes(span, map[string]string{
		"request_id": requestID,
		"operation":  functionName,
		"topic":      req.Topic,
		"group_id":   req.GroupId,
	})

	// Increment total requests metric
	s.obs.MetricsService.IncrementCounter(tCtx, metricSubscribeTotal, 1, nil)

	// Validate request
	if customErr := validation.ValidateSubscribeRequest(req); customErr != nil {
		s.obs.LoggerService.Error(tCtx, fmt.Sprintf("[%s] Subscription validation failed", requestID))
		s.obs.MetricsService.IncrementCounter(tCtx, metricSubscribeFailure, 1, map[string]string{
			"error_type": errorTypeSubscribeValidation,
		})
		return errors.NewGRPCError(customErr)
	}

	// Create Confluent Kafka configuration with read committed and latest offset
	cfg := s.kafkaConfig.CreateSubscribeConfig(req.Topic, req.GroupId, req)

	// Log the subscribe request
	s.obs.LoggerService.Info(tCtx, fmt.Sprintf("[%s] Initiating subscription to topic: %s with group: %s",
		requestID, req.Topic, req.GroupId))

	// Start consuming messages
	customErr := s.messagingService.ConsumeMessage(stream, cfg)
	if customErr != nil {
		s.obs.LoggerService.Error(tCtx, fmt.Sprintf("[%s] Failed to establish subscription to topic: %s", requestID, req.Topic))
		s.obs.MetricsService.IncrementCounter(tCtx, metricSubscribeFailure, 1, map[string]string{
			"error_type": errorTypeSubscribe,
		})
		return errors.NewGRPCError(customErr)
	}

	s.obs.LoggerService.Info(tCtx, fmt.Sprintf("[%s] Successfully subscribed to topic: %s", requestID, req.Topic))
	s.obs.MetricsService.IncrementCounter(tCtx, metricSubscribeSuccess, 1, nil)

	return nil
}

func (s *MessagingHandler) CreateTopicV1(ctx context.Context, req *pb.CreateTopicRequest) (*pb.CreateTopicResponse, error) {
	functionName := "CreateTopicV1"

	// Create timeout context
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	// Start tracing
	tCtx, span := s.obs.TracerService.StartTracer(ctx, functionName)
	defer s.obs.TracerService.StopSpan(span)

	// Generate request ID for correlation
	requestID := fmt.Sprintf("req-%d", time.Now().UnixNano())

	// Set tracing attributes
	s.obs.TracerService.SetAttributes(span, map[string]string{
		"request_id": requestID,
		"operation":  functionName,
		"topic":      req.Topic,
	})

	// Increment total requests metric
	s.obs.MetricsService.IncrementCounter(tCtx, metricCreateTopicTotal, 1, nil)

	// Validate request
	if customErr := validation.ValidateCreateTopicRequest(req); customErr != nil {
		s.obs.LoggerService.Error(tCtx, fmt.Sprintf("[%s] Topic creation validation failed", requestID))
		s.obs.MetricsService.IncrementCounter(tCtx, metricCreateTopicFailure, 1, map[string]string{
			"error_type": errorTypeTopicValidation,
		})
		return nil, errors.NewGRPCError(customErr)
	}

	// Create Confluent Kafka configuration with default values
	cfg := s.kafkaConfig.CreateTopicConfig(req.Topic, req)

	// Log the create topic request using configured values
	s.obs.LoggerService.Info(tCtx, fmt.Sprintf("[%s] Creating topic: %s with partitions: %d, replication factor: %d",
		requestID, req.Topic, s.config.KafkaNumPartitions, s.config.KafkaReplicationFactor))

	// Create topic
	customErr := s.messagingService.CreateTopic(tCtx, req, cfg)
	if customErr != nil {
		// Check if this is a "topic already exists" error, which is not actually an error
		if customErr.ErrorCode == errors.TOPErrTopicExists {
			s.obs.LoggerService.Info(tCtx, fmt.Sprintf("[%s] Topic already exists: %s", requestID, req.Topic))
			s.obs.MetricsService.IncrementCounter(tCtx, metricCreateTopicSuccess, 1, nil)

			// Return success response with a message indicating the topic already exists
			return &pb.CreateTopicResponse{
				Status:  "success",
				Message: fmt.Sprintf("Topic %s already exists", req.Topic),
			}, nil
		}

		// For any other error, handle as before
		s.obs.LoggerService.Error(tCtx, fmt.Sprintf("[%s] Failed to create topic: %s", requestID, req.Topic))
		s.obs.MetricsService.IncrementCounter(tCtx, metricCreateTopicFailure, 1, map[string]string{
			"error_type": errorTypeCreateTopic,
		})
		return nil, errors.NewGRPCError(customErr)
	}

	s.obs.LoggerService.Info(tCtx, fmt.Sprintf("[%s] Successfully created topic: %s", requestID, req.Topic))
	s.obs.MetricsService.IncrementCounter(tCtx, metricCreateTopicSuccess, 1, nil)

	return &pb.CreateTopicResponse{
		Status: "success",
		Message: fmt.Sprintf("Topic %s created successfully with %d partitions and replication factor %d",
			req.Topic, s.config.KafkaNumPartitions, s.config.KafkaReplicationFactor),
	}, nil
}

// UpdateMessagingService updates the messaging service with a new instance
func (h *MessagingHandler) UpdateMessagingService(config *appconfig.Config, messagingService service.MessagingService) error {
	// Create context for tracing and metrics
	ctx := context.Background()
	functionName := "UpdateMessagingService"

	// Start tracing
	tCtx, span := h.obs.TracerService.StartTracer(ctx, functionName)
	defer h.obs.TracerService.StopSpan(span)

	// Generate request ID for correlation
	requestID := fmt.Sprintf("req-%d", time.Now().UnixNano())

	// Set tracing attributes
	h.obs.TracerService.SetAttributes(span, map[string]string{
		"request_id": requestID,
		"operation":  functionName,
	})

	// Increment total requests metric
	h.obs.MetricsService.IncrementCounter(tCtx, metricUpdateHandlerTotal, 1, nil)

	// Validate inputs
	if config == nil {
		h.obs.LoggerService.Error(tCtx, fmt.Sprintf("[%s] Missing configuration for handler update", requestID))
		h.obs.MetricsService.IncrementCounter(tCtx, metricUpdateHandlerFailure, 1, map[string]string{
			"error_type": errorTypeConfig,
		})
		return fmt.Errorf("config cannot be nil when updating messaging service")
	}
	if messagingService == nil {
		h.obs.LoggerService.Error(tCtx, fmt.Sprintf("[%s] Missing messaging service for handler update", requestID))
		h.obs.MetricsService.IncrementCounter(tCtx, metricUpdateHandlerFailure, 1, map[string]string{
			"error_type": errorTypeConfig,
		})
		return fmt.Errorf("messagingService cannot be nil when updating messaging service")
	}

	// Validate kafka brokers
	if config.KafkaBrokers == nil || len(config.KafkaBrokers) == 0 {
		h.obs.LoggerService.Error(tCtx, fmt.Sprintf("[%s] Missing Kafka brokers configuration", requestID))
		h.obs.MetricsService.IncrementCounter(tCtx, metricUpdateHandlerFailure, 1, map[string]string{
			"error_type": errorTypeConfig,
		})
		return fmt.Errorf("kafka brokers must be configured")
	}

	h.obs.LoggerService.Info(tCtx, fmt.Sprintf("[%s] Updating messaging handler configuration", requestID))

	// Create new configurator
	kafkaConfig, err := confluent.SafeNewConfigurator(config.KafkaBrokers, config)
	if err != nil {
		h.obs.LoggerService.Error(tCtx, fmt.Sprintf("[%s] Failed to create Kafka configurator", requestID))
		h.obs.MetricsService.IncrementCounter(tCtx, metricUpdateHandlerFailure, 1, map[string]string{
			"error_type": errorTypeConfig,
		})
		return err
	}

	// Update handler fields
	h.config = config
	h.messagingService = messagingService
	h.kafkaConfig = kafkaConfig

	h.obs.LoggerService.Info(tCtx, fmt.Sprintf("[%s] Successfully updated messaging handler configuration", requestID))
	h.obs.MetricsService.IncrementCounter(tCtx, metricUpdateHandlerSuccess, 1, nil)

	return nil
}
