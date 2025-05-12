package grpc

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	"redis-service/internal/app"
	"redis-service/internal/dto"
	"redis-service/pkg/message"
	"redis-service/proto"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	tracenoop "go.opentelemetry.io/otel/trace/noop"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// --- Mock RedisServiceInterface ---
type mockRedisService struct {
	mock.Mock
}

func (m *mockRedisService) GetCache(req *dto.GetCacheRequest) (string, error) {
	args := m.Called(req)
	return args.String(0), args.Error(1)
}
func (m *mockRedisService) SetCache(req *dto.SetCacheRequest) (bool, error) {
	args := m.Called(req)
	return args.Bool(0), args.Error(1)
}
func (m *mockRedisService) InvalidateCache(req *dto.DeleteCacheRequest) (bool, error) {
	args := m.Called(req)
	return args.Bool(0), args.Error(1)
}

func setupTestContainer() {
	// Create a no-op tracer
	tracer := tracenoop.NewTracerProvider().Tracer("test")
	// Create a new logger that discards output
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	app.Di = &app.Container{
		Tracer: tracer,
		Logger: logger,
	}
}

func TestGRPCServer_GetCache(t *testing.T) {
	setupTestContainer()
	ctx := context.Background()
	mockSvc := new(mockRedisService)
	server := &GRPCServer{
		service: mockSvc,
		tracer:  app.Di.Tracer,
		logger:  app.Di.Logger,
	}

	validReq := &proto.GetCacheRequest{
		Namespace: "ns",
		Key:       "key",
	}

	// --- Success case ---
	mockSvc.On("GetCache", mock.MatchedBy(func(req *dto.GetCacheRequest) bool {
		return req.Namespace == "ns" && req.Key == "key"
	})).Return("the-value", nil).Once()

	resp, err := server.GetCache(ctx, validReq)
	assert.NoError(t, err)
	assert.True(t, resp.Found)
	assert.Equal(t, "the-value", resp.Value)
	mockSvc.AssertExpectations(t)

	// Reset mock for next test case
	mockSvc.ExpectedCalls = nil
	mockSvc.Calls = nil

	// --- Validation error (empty namespace) ---
	invalidReq := &proto.GetCacheRequest{
		Namespace: "",
		Key:       "key",
	}
	resp, err = server.GetCache(ctx, invalidReq)
	assert.NoError(t, err)
	assert.False(t, resp.Found)
	assert.Equal(t, message.RD0004, resp.Message)
	mockSvc.AssertNotCalled(t, "GetCache")

	// Reset mock for next test case
	mockSvc.ExpectedCalls = nil
	mockSvc.Calls = nil

	// --- Service error ---
	mockSvc.On("GetCache", mock.MatchedBy(func(req *dto.GetCacheRequest) bool {
		return req.Namespace == "ns" && req.Key == "key"
	})).Return("", errors.New("redis error")).Once()

	resp, err = server.GetCache(ctx, validReq)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, codes.Internal, status.Code(err))
	mockSvc.AssertExpectations(t)
}

func TestGRPCServer_SetCache(t *testing.T) {
	setupTestContainer()
	ctx := context.Background()
	mockSvc := new(mockRedisService)
	server := &GRPCServer{
		service: mockSvc,
		tracer:  app.Di.Tracer,
		logger:  app.Di.Logger,
	}

	validReq := &proto.SetCacheRequest{
		Namespace: "abc",
		Key:       "def",
		Value:     "val",
		Ttl:       10,
	}

	// --- Success case ---
	mockSvc.On("SetCache", mock.AnythingOfType("*dto.SetCacheRequest")).Return(true, nil).Once()
	resp, err := server.SetCache(ctx, validReq)
	assert.NoError(t, err)
	assert.True(t, resp.Success)
	assert.Equal(t, message.RD0000, resp.Message)

	// --- Service error ---
	mockSvc.On("SetCache", mock.AnythingOfType("*dto.SetCacheRequest")).Return(false, errors.New("redis error")).Once()
	resp, err = server.SetCache(ctx, validReq)
	assert.NoError(t, err)
	assert.False(t, resp.Success)
	assert.Equal(t, message.RD0002, resp.Message)
}

func TestGRPCServer_InvalidateCache(t *testing.T) {
	setupTestContainer()
	ctx := context.Background()
	mockSvc := new(mockRedisService)
	server := &GRPCServer{
		service: mockSvc,
		tracer:  app.Di.Tracer,
		logger:  app.Di.Logger,
	}

	validReq := &proto.InvalidateCacheRequest{
		Namespace: "ns",
		Key:       "key",
	}

	// --- Success case ---
	mockSvc.On("InvalidateCache", mock.AnythingOfType("*dto.DeleteCacheRequest")).Return(true, nil).Once()
	resp, err := server.InvalidateCache(ctx, validReq)
	assert.NoError(t, err)
	assert.True(t, resp.Success)
	assert.Equal(t, message.RD0000, resp.Message)

	// Reset mock for next test case
	mockSvc.ExpectedCalls = nil
	mockSvc.Calls = nil

	// --- Validation error (empty namespace) ---
	invalidReq := &proto.InvalidateCacheRequest{
		Namespace: "",
		Key:       "key",
	}
	resp, err = server.InvalidateCache(ctx, invalidReq)
	assert.NoError(t, err)
	assert.False(t, resp.Success)
	assert.Equal(t, message.RD0004, resp.Message)
	mockSvc.AssertNotCalled(t, "InvalidateCache")

	// Reset mock for next test case
	mockSvc.ExpectedCalls = nil
	mockSvc.Calls = nil

	// --- Service error ---
	mockSvc.On("InvalidateCache", mock.AnythingOfType("*dto.DeleteCacheRequest")).Return(false, errors.New("redis error")).Once()
	resp, err = server.InvalidateCache(ctx, validReq)
	assert.NoError(t, err)
	assert.False(t, resp.Success)
	assert.Equal(t, message.RD0002, resp.Message)
	mockSvc.AssertExpectations(t)
}
