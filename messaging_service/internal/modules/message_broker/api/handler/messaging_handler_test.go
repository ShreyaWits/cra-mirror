package handler_test

import (
	"context"
	pb "cra-protos/messaging_service"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"

	"messaging_service/internal/config"
	"messaging_service/internal/modules/message_broker/api/handler"
	mock_service "messaging_service/internal/modules/message_broker/mock"
	"messaging_service/internal/modules/message_broker/service"
	"messaging_service/pkg/errors"
	"messaging_service/pkg/observability"
)

func setupHandler(t *testing.T) (*handler.MessagingHandler, *mock_service.MockMessagingService, *gomock.Controller) {
	ctrl := gomock.NewController(t)
	mockSvc := mock_service.NewMockMessagingService(ctrl)

	cfg := &config.Config{
		KafkaBrokers:           []string{"localhost:9092"},
		KafkaNumPartitions:     3,
		KafkaReplicationFactor: 3,
	}
	h, _ := handler.NewMessagingHandler(cfg, mockSvc, observability.NewObservabilityStack(&config.Env{ServiceName: "test-service", ConfigServiceUrl: "http://localhost:8080"}))

	return h, mockSvc, ctrl
}

func TestNewMessagingHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := mock_service.NewMockMessagingService(ctrl)
	validConfig := &config.Config{
		KafkaBrokers:           []string{"localhost:9092"},
		KafkaNumPartitions:     3,
		KafkaReplicationFactor: 3,
	}
	validObs := observability.NewObservabilityStack(&config.Env{ServiceName: "test-service", ConfigServiceUrl: "http://localhost:8080"})

	tests := []struct {
		name        string
		config      *config.Config
		service     service.MessagingService
		obs         *observability.ObservabilityStack
		expectError bool
	}{
		{
			name:        "nil config",
			config:      nil,
			service:     mockSvc,
			obs:         validObs,
			expectError: true,
		},
		{
			name:        "nil service",
			config:      validConfig,
			service:     nil,
			obs:         validObs,
			expectError: true,
		},
		{
			name:        "nil observability stack",
			config:      validConfig,
			service:     mockSvc,
			obs:         nil,
			expectError: true,
		},
		{
			name: "empty kafka brokers",
			config: &config.Config{
				KafkaBrokers:           []string{},
				KafkaNumPartitions:     3,
				KafkaReplicationFactor: 3,
			},
			service:     mockSvc,
			obs:         validObs,
			expectError: true,
		},
		{
			name:        "valid parameters",
			config:      validConfig,
			service:     mockSvc,
			obs:         validObs,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h, err := handler.NewMessagingHandler(tt.config, tt.service, tt.obs)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, h)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, h)
			}
		})
	}
}

func TestPublishMessageV1(t *testing.T) {
	handler, mockSvc, ctrl := setupHandler(t)
	defer ctrl.Finish()

	tests := []struct {
		name        string
		req         *pb.PublishRequest
		mockSetup   func(req *pb.PublishRequest)
		expectError bool
		expectResp  *pb.PublishResponse
	}{
		{
			name: "valid request",
			req: &pb.PublishRequest{
				Topic: "valid-topic",
				Value: map[string]string{"key": "value"},
			},
			mockSetup: func(req *pb.PublishRequest) {
				mockSvc.EXPECT().
					PublishMessage(gomock.Any(), gomock.Any(), req).
					Return(nil)
			},
			expectError: false,
			expectResp: &pb.PublishResponse{
				Status:  "success",
				Message: "Message published successfully to topic valid-topic",
			},
		},
		{
			name: "valid request with custom key",
			req: &pb.PublishRequest{
				Topic: "transaction-topic",
				Value: map[string]string{"key": "value"},
				Key:   "custom-key-1234",
			},
			mockSetup: func(req *pb.PublishRequest) {
				mockSvc.EXPECT().
					PublishMessage(gomock.Any(), gomock.Any(), req).
					Return(nil)
			},
			expectError: false,
			expectResp: &pb.PublishResponse{
				Status:  "success",
				Message: "Message published successfully to topic transaction-topic",
			},
		},
		{
			name:        "invalid request (missing topic)",
			req:         &pb.PublishRequest{},
			mockSetup:   func(req *pb.PublishRequest) {},
			expectError: true,
			expectResp:  nil,
		},
		{
			name: "invalid request (missing value)",
			req: &pb.PublishRequest{
				Topic: "test-topic",
			},
			mockSetup:   func(req *pb.PublishRequest) {},
			expectError: true,
			expectResp:  nil,
		},
		{
			name: "service error - publish failed",
			req: &pb.PublishRequest{
				Topic: "test-topic",
				Value: map[string]string{"key": "value"},
			},
			mockSetup: func(req *pb.PublishRequest) {
				mockSvc.EXPECT().
					PublishMessage(gomock.Any(), gomock.Any(), req).
					Return(errors.NewCustomError(errors.PUBErrPublishFailed, fmt.Errorf("internal error")))
			},
			expectError: true,
			expectResp:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			// Setup mocks
			tt.mockSetup(tt.req)

			resp, err := handler.PublishMessageV1(ctx, tt.req)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Equal(t, tt.expectResp.Status, resp.Status)
				assert.Equal(t, tt.expectResp.Message, resp.Message)
			}
		})
	}
}

func TestCreateTopicV1(t *testing.T) {
	handler, mockSvc, ctrl := setupHandler(t)
	defer ctrl.Finish()

	tests := []struct {
		name        string
		req         *pb.CreateTopicRequest
		mockSetup   func(req *pb.CreateTopicRequest)
		expectError bool
		expectResp  *pb.CreateTopicResponse
	}{
		{
			name: "valid request",
			req: &pb.CreateTopicRequest{
				Topic: "new-topic",
			},
			mockSetup: func(req *pb.CreateTopicRequest) {
				mockSvc.EXPECT().
					CreateTopic(gomock.Any(), req, gomock.Any()).
					Return(nil)
			},
			expectError: false,
			expectResp: &pb.CreateTopicResponse{
				Status:  "success",
				Message: "Topic new-topic created successfully with 3 partitions and replication factor 3",
			},
		},
		{
			name:        "invalid request (missing topic)",
			req:         &pb.CreateTopicRequest{},
			mockSetup:   func(req *pb.CreateTopicRequest) {},
			expectError: true,
			expectResp:  nil,
		},
		{
			name: "topic already exists",
			req: &pb.CreateTopicRequest{
				Topic: "existing-topic",
			},
			mockSetup: func(req *pb.CreateTopicRequest) {
				mockSvc.EXPECT().
					CreateTopic(gomock.Any(), req, gomock.Any()).
					Return(errors.NewCustomError(errors.TOPErrTopicExists, fmt.Errorf("topic already exists")))
			},
			expectError: false,
			expectResp: &pb.CreateTopicResponse{
				Status:  "success",
				Message: "Topic existing-topic already exists",
			},
		},
		{
			name: "service error - creation failed",
			req: &pb.CreateTopicRequest{
				Topic: "fail-topic",
			},
			mockSetup: func(req *pb.CreateTopicRequest) {
				mockSvc.EXPECT().
					CreateTopic(gomock.Any(), req, gomock.Any()).
					Return(errors.NewCustomError(errors.TOPErrCreateFailed, fmt.Errorf("topic creation failed")))
			},
			expectError: true,
			expectResp:  nil,
		},
		{
			name: "service error - other error",
			req: &pb.CreateTopicRequest{
				Topic: "connection-fail-topic",
			},
			mockSetup: func(req *pb.CreateTopicRequest) {
				mockSvc.EXPECT().
					CreateTopic(gomock.Any(), req, gomock.Any()).
					Return(errors.NewCustomError(errors.TOPErrCreateFailed, fmt.Errorf("other error")))
			},
			expectError: true,
			expectResp:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			// Setup mocks
			tt.mockSetup(tt.req)

			resp, err := handler.CreateTopicV1(ctx, tt.req)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Equal(t, tt.expectResp.Status, resp.Status)
				assert.Equal(t, tt.expectResp.Message, resp.Message)
			}
		})
	}
}

func TestSubscribeV1(t *testing.T) {
	handler, mockSvc, ctrl := setupHandler(t)
	defer ctrl.Finish()

	tests := []struct {
		name        string
		req         *pb.SubscribeRequest
		mockSetup   func(req *pb.SubscribeRequest, stream pb.MessagingService_SubscribeV1Server)
		expectError bool
	}{
		{
			name: "valid subscription",
			req: &pb.SubscribeRequest{
				Topic:   "subscribe-topic",
				GroupId: "group-1",
			},
			mockSetup: func(req *pb.SubscribeRequest, stream pb.MessagingService_SubscribeV1Server) {
				mockSvc.EXPECT().
					ConsumeMessage(stream, gomock.Any()).
					Return(nil)
			},
			expectError: false,
		},
		{
			name: "valid subscription with options",
			req: &pb.SubscribeRequest{
				Topic:   "subscribe-topic-options",
				GroupId: "group-options",
			},
			mockSetup: func(req *pb.SubscribeRequest, stream pb.MessagingService_SubscribeV1Server) {
				mockSvc.EXPECT().
					ConsumeMessage(stream, gomock.Any()).
					Return(nil)
			},
			expectError: false,
		},
		{
			name: "invalid subscription (missing topic)",
			req: &pb.SubscribeRequest{
				GroupId: "group-1",
			},
			mockSetup:   func(req *pb.SubscribeRequest, stream pb.MessagingService_SubscribeV1Server) {},
			expectError: true,
		},
		{
			name: "invalid subscription (missing group ID)",
			req: &pb.SubscribeRequest{
				Topic: "topic-no-group",
			},
			mockSetup:   func(req *pb.SubscribeRequest, stream pb.MessagingService_SubscribeV1Server) {},
			expectError: true,
		},
		{
			name: "consume error - subscription failed",
			req: &pb.SubscribeRequest{
				Topic:   "topic-fail",
				GroupId: "group-1",
			},
			mockSetup: func(req *pb.SubscribeRequest, stream pb.MessagingService_SubscribeV1Server) {
				mockSvc.EXPECT().
					ConsumeMessage(stream, gomock.Any()).
					Return(errors.NewCustomError(errors.SUBErrSubscribeFailed, fmt.Errorf("consume failed")))
			},
			expectError: true,
		},
		{
			name: "consume error - other error",
			req: &pb.SubscribeRequest{
				Topic:   "topic-not-found",
				GroupId: "group-1",
			},
			mockSetup: func(req *pb.SubscribeRequest, stream pb.MessagingService_SubscribeV1Server) {
				mockSvc.EXPECT().
					ConsumeMessage(stream, gomock.Any()).
					Return(errors.NewCustomError(errors.SUBErrSubscribeFailed, fmt.Errorf("topic not found")))
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStream := &mockServerStream{}

			// Setup mocks
			tt.mockSetup(tt.req, mockStream)

			err := handler.SubscribeV1(tt.req, mockStream)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUpdateMessagingService(t *testing.T) {
	handler, mockSvc, ctrl := setupHandler(t)
	defer ctrl.Finish()

	tests := []struct {
		name        string
		config      *config.Config
		service     service.MessagingService
		expectError bool
	}{
		{
			name: "successful update",
			config: &config.Config{
				KafkaBrokers:           []string{"new-broker:9092"},
				KafkaNumPartitions:     5,
				KafkaReplicationFactor: 2,
			},
			service:     mockSvc,
			expectError: false,
		},
		{
			name:        "nil config",
			config:      nil,
			service:     mockSvc,
			expectError: true,
		},
		{
			name: "nil service",
			config: &config.Config{
				KafkaBrokers:           []string{"localhost:9092"},
				KafkaNumPartitions:     3,
				KafkaReplicationFactor: 3,
			},
			service:     nil,
			expectError: true,
		},
		{
			name: "empty kafka brokers",
			config: &config.Config{
				KafkaBrokers:           []string{},
				KafkaNumPartitions:     3,
				KafkaReplicationFactor: 3,
			},
			service:     mockSvc,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := handler.UpdateMessagingService(tt.config, tt.service)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				// We've confirmed no error was returned for valid scenarios
				// We can't directly access the handler's private fields to verify they were updated
			}
		})
	}
}

func TestHelperFunctions(t *testing.T) {
	// We can't test getRequestID directly since it's unexported,
	// so we'll just test that the code with observability works properly.

	// The tests for PublishMessageV1, SubscribeV1, and CreateTopicV1 all use
	// the observability stack, so we have good coverage of the actual functionality.

	// Instead, let's test that our configuration of the observability stack works
	t.Run("observability stack initialization", func(t *testing.T) {
		// Create a new observability stack
		obs := observability.NewObservabilityStack(&config.Env{
			ServiceName:      "test-service",
			ConfigServiceUrl: "http://localhost:8080",
		})

		// Verify it's not nil
		assert.NotNil(t, obs)
		assert.NotNil(t, obs.LoggerService)
		assert.NotNil(t, obs.MetricsService)
		assert.NotNil(t, obs.TracerService)

		// Verify we can use it without panicking
		obs.LoggerService.Info(context.Background(), "Test message")
		obs.MetricsService.IncrementCounter(context.Background(), "test_counter", 1, map[string]string{})
		ctx, span := obs.TracerService.StartTracer(context.Background(), "test_span")
		assert.NotNil(t, ctx)
		assert.NotNil(t, span)
		obs.TracerService.StopSpan(span)
	})
}

// --- Mock stream for SubscribeV1 ---
type mockServerStream struct {
	grpc.ServerStream
}

func (m *mockServerStream) Context() context.Context {
	return context.Background()
}

func (m *mockServerStream) Send(msg *pb.KafkaMessage) error {
	return nil
}
