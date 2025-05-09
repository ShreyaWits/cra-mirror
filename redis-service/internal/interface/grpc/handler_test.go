package grpc

import (
	"context"
	"errors"
	"testing"

	"redis-service/internal/dto"
	"redis-service/pkg/message"
	"redis-service/proto"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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

func TestGRPCServer_GetCache(t *testing.T) {
	ctx := context.Background()
	mockSvc := new(mockRedisService)
	server := &GRPCServer{service: mockSvc}

	validReq := &proto.GetCacheRequest{
		Namespace:  "ns",
		Key:        "key",
		TrackingId: uuid.NewString(),
	}

	// --- Success case ---
	mockSvc.On("GetCache", mock.AnythingOfType("*dto.GetCacheRequest")).Return("the-value", nil).Once()
	resp, err := server.GetCache(ctx, validReq)
	assert.NoError(t, err)
	assert.True(t, resp.Found)
	assert.Equal(t, "the-value", resp.Value)
	assert.Equal(t, message.RD0000, resp.Message)
	mockSvc.AssertExpectations(t)

	// --- Validation error (invalid TrackingId) ---
	invalidReq := &proto.GetCacheRequest{
		Namespace:  "ns",
		Key:        "key",
		TrackingId: "not-a-uuid",
	}
	resp, err = server.GetCache(ctx, invalidReq)
	assert.NoError(t, err)
	assert.False(t, resp.Found)
	assert.Equal(t, message.RD0001, resp.Message)
	assert.NotEmpty(t, resp.Error)
	// No mockSvc.On or AssertExpectations here, since service should NOT be called

	// --- Service error ---
	mockSvc.On("GetCache", mock.AnythingOfType("*dto.GetCacheRequest")).Return("", errors.New("redis error")).Once()
	resp, err = server.GetCache(ctx, validReq)
	assert.NoError(t, err)
	assert.False(t, resp.Found)
	assert.Equal(t, "", resp.Value)
	assert.Equal(t, message.RD0002, resp.Message)
	mockSvc.AssertExpectations(t)
}

func TestGRPCServer_SetCache(t *testing.T) {
	ctx := context.Background()
	mockSvc := new(mockRedisService)
	server := &GRPCServer{service: mockSvc}

	validReq := &proto.SetCacheRequest{
		Namespace:  "abc",
		Key:        "def",
		TrackingId: uuid.NewString(),
		Value:      "val",
		Ttl:        10,
	}

	// --- Success case ---
	mockSvc.On("SetCache", mock.AnythingOfType("*dto.SetCacheRequest")).Return(true, nil).Once()
	resp, err := server.SetCache(ctx, validReq)
	assert.NoError(t, err)
	assert.True(t, resp.Success)
	assert.Equal(t, message.RD0000, resp.Message)

	// --- Validation error (invalid TrackingId) ---
	invalidReq := &proto.SetCacheRequest{
		Namespace:  "abc",
		Key:        "def",
		TrackingId: "not-a-uuid",
		Value:      "val",
		Ttl:        10,
	}
	resp, err = server.SetCache(ctx, invalidReq)
	assert.NoError(t, err)
	assert.False(t, resp.Success)
	assert.Equal(t, message.RD0004, resp.Message)

	// --- Service error ---
	mockSvc.On("SetCache", mock.AnythingOfType("*dto.SetCacheRequest")).Return(false, errors.New("redis error")).Once()
	resp, err = server.SetCache(ctx, validReq)
	assert.NoError(t, err)
	assert.False(t, resp.Success)
	assert.Equal(t, message.RD0002, resp.Message)
}

func TestGRPCServer_InvalidateCache(t *testing.T) {
	ctx := context.Background()
	mockSvc := new(mockRedisService)
	server := &GRPCServer{service: mockSvc}

	validReq := &proto.InvalidateCacheRequest{
		Namespace:  "ns",
		Key:        "key",
		TrackingId: uuid.NewString(),
	}

	// --- Success case ---
	mockSvc.On("InvalidateCache", mock.AnythingOfType("*dto.DeleteCacheRequest")).Return(true, nil).Once()
	resp, err := server.InvalidateCache(ctx, validReq)
	assert.NoError(t, err)
	assert.True(t, resp.Success)
	assert.Equal(t, message.RD0000, resp.Message)

	// --- Validation error (invalid TrackingId) ---
	invalidReq := &proto.InvalidateCacheRequest{
		Namespace:  "ns",
		Key:        "key",
		TrackingId: "not-a-uuid",
	}
	resp, err = server.InvalidateCache(ctx, invalidReq)
	assert.NoError(t, err)
	assert.False(t, resp.Success)
	assert.Equal(t, message.RD0004, resp.Message)

	// --- Service error ---
	mockSvc.On("InvalidateCache", mock.AnythingOfType("*dto.DeleteCacheRequest")).Return(false, errors.New("redis error")).Once()
	resp, err = server.InvalidateCache(ctx, validReq)
	assert.NoError(t, err)
	assert.False(t, resp.Success)
	assert.Equal(t, message.RD0002, resp.Message)
}
