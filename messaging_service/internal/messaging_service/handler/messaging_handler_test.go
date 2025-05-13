package handler_test

import (
	"context"
	pb "cra-protos/messaging_service"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"google.golang.org/grpc"

	"messaging_service/internal/config"
	"messaging_service/internal/messaging_service/handler"
	mock_service "messaging_service/internal/messaging_service/mock"
	"messaging_service/pkg/errors"
)

func setupHandler(t *testing.T) (*handler.MessagingHandler, *mock_service.MockMessagingService) {
	ctrl := gomock.NewController(t)
	mockSvc := mock_service.NewMockMessagingService(ctrl)

	cfg := &config.Config{
		KafkaBrokers: []string{"localhost:9092"},
	}
	h := handler.NewMessagingHandler(cfg, mockSvc)

	return h, mockSvc
}

func TestPublishMessageV1(t *testing.T) {
	handler, mockSvc := setupHandler(t)

	tests := []struct {
		name        string
		req         *pb.PublishRequest
		mockError   *errors.CustomError
		expectError bool
	}{
		{
			name: "valid request",
			req: &pb.PublishRequest{
				Topic: "valid-topic",
				Value: map[string]string{"key": "value"},
			},
			mockError:   nil,
			expectError: false,
		},
		{
			name:        "invalid request (missing topic)",
			req:         &pb.PublishRequest{},
			expectError: true,
		},
		{
			name: "service error",
			req: &pb.PublishRequest{
				Topic: "test-topic",
				Value: map[string]string{"key": "value"},
			},
			mockError:   errors.NewCustomError(errors.PUBErrPublishFailed, fmt.Errorf("internal error")),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			if tt.req.Topic != "" && tt.req.Value != nil {
				mockSvc.EXPECT().
					PublishMessage(gomock.Any(), gomock.Any(), tt.req).
					Return(tt.mockError)
			}

			resp, err := handler.PublishMessageV1(ctx, tt.req)

			if tt.expectError && err == nil {
				t.Fatalf("expected error but got nil")
			}
			if !tt.expectError && err != nil {
				t.Fatalf("did not expect error but got %v", err)
			}
			if !tt.expectError && resp == nil {
				t.Fatal("expected non-nil response")
			}
		})
	}
}

func TestCreateTopicV1(t *testing.T) {
	handler, mockSvc := setupHandler(t)

	tests := []struct {
		name        string
		req         *pb.CreateTopicRequest
		mockError   *errors.CustomError
		expectError bool
	}{
		{
			name: "valid request",
			req: &pb.CreateTopicRequest{
				Topic: "new-topic",
			},
			mockError:   nil,
			expectError: false,
		},
		{
			name:        "invalid request (missing topic)",
			req:         &pb.CreateTopicRequest{},
			expectError: true,
		},
		{
			name: "service error",
			req: &pb.CreateTopicRequest{
				Topic: "fail-topic",
			},
			mockError:   errors.NewCustomError(errors.TOPErrCreateFailed, fmt.Errorf("topic creation failed")),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			if tt.req.Topic != "" {
				mockSvc.EXPECT().
					CreateTopic(gomock.Any(), tt.req, gomock.Any()).
					Return(tt.mockError)
			}

			resp, err := handler.CreateTopicV1(ctx, tt.req)

			if tt.expectError && err == nil {
				t.Fatal("expected error but got nil")
			}
			if !tt.expectError && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !tt.expectError && resp == nil {
				t.Fatal("expected response, got nil")
			}
		})
	}
}

func TestSubscribeV1(t *testing.T) {
	handler, mockSvc := setupHandler(t)

	tests := []struct {
		name        string
		req         *pb.SubscribeRequest
		mockError   *errors.CustomError
		expectError bool
	}{
		{
			name: "valid subscription",
			req: &pb.SubscribeRequest{
				Topic:   "subscribe-topic",
				GroupId: "group-1",
			},
			mockError:   nil,
			expectError: false,
		},
		{
			name: "invalid subscription (missing topic)",
			req: &pb.SubscribeRequest{
				GroupId: "group-1",
			},
			expectError: true,
		},
		{
			name: "consume error",
			req: &pb.SubscribeRequest{
				Topic:   "topic-fail",
				GroupId: "group-1",
			},
			mockError:   errors.NewCustomError(errors.SUBErrSubscribeFailed, fmt.Errorf("consume failed")),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStream := &mockServerStream{}

			if tt.req.Topic != "" {
				mockSvc.EXPECT().
					ConsumeMessage(mockStream, gomock.Any()).
					Return(tt.mockError)
			}

			err := handler.SubscribeV1(tt.req, mockStream)

			if tt.expectError && err == nil {
				t.Fatal("expected error but got nil")
			}
			if !tt.expectError && err != nil {
				t.Fatalf("unexpected error: %v", err)
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
