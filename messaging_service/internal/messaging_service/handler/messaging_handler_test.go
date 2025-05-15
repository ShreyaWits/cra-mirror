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
	"messaging_service/internal/messaging_service/handler"
	mock_service "messaging_service/internal/messaging_service/mock"
	"messaging_service/pkg/errors"
	"messaging_service/pkg/logger"
)

func init() {
	// Initialize logger to avoid nil pointer dereference
	logger.InitLogger()
}

func setupHandler(t *testing.T) (*handler.MessagingHandler, *mock_service.MockMessagingService, *gomock.Controller) {
	ctrl := gomock.NewController(t)
	mockSvc := mock_service.NewMockMessagingService(ctrl)

	cfg := &config.Config{
		KafkaBrokers:           []string{"localhost:9092"},
		KafkaNumPartitions:     3,
		KafkaReplicationFactor: 3,
	}
	h := handler.NewMessagingHandler(cfg, mockSvc)

	return h, mockSvc, ctrl
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
