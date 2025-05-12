package service

import (
	"errors"
	"log/slog"
	"redis-service/internal/dto"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockRedisRepository is a mock implementation of RedisRepositoryInterface
type MockRedisRepository struct {
	mock.Mock
}

func (m *MockRedisRepository) GetCache(key string) (string, error) {
	args := m.Called(key)
	return args.String(0), args.Error(1)
}

func (m *MockRedisRepository) SetCache(key string, value string, ttl *time.Duration) error {
	args := m.Called(key, value, ttl)
	return args.Error(0)
}

func (m *MockRedisRepository) InvalidateCache(key string) error {
	args := m.Called(key)
	return args.Error(0)
}

func TestRedisService_GetCache(t *testing.T) {
	mockRepo := new(MockRedisRepository)
	// Create a dummy logger for testing
	logger := slog.Default()
	service := NewRedisService(mockRepo, logger)

	payload := &dto.GetCacheRequest{
		Namespace: "ns",
		Key:       "k1",
	}
	expectedKey := "ns:k1"
	expectedValue := "cached-value"

	mockRepo.On("GetCache", expectedKey).Return(expectedValue, nil)

	val, err := service.GetCache(payload)
	assert.NoError(t, err)
	assert.Equal(t, expectedValue, val)
	mockRepo.AssertExpectations(t)
}

func TestRedisService_SetCache_Success(t *testing.T) {
	mockRepo := new(MockRedisRepository)
	// Create a dummy logger for testing
	logger := slog.Default()
	service := NewRedisService(mockRepo, logger)

	payload := &dto.SetCacheRequest{
		Namespace: "ns",
		Key:       "k2",
		Value:     "val2",
		TTL:       60,
	}
	expectedKey := "ns:k2"
	expectedTTL := time.Duration(60) * time.Second

	mockRepo.On("SetCache", expectedKey, "val2", mock.MatchedBy(func(ttl *time.Duration) bool {
		return ttl != nil && *ttl == expectedTTL
	})).Return(nil)

	ok, err := service.SetCache(payload)
	assert.NoError(t, err)
	assert.True(t, ok)
	mockRepo.AssertExpectations(t)
}

func TestRedisService_SetCache_Error(t *testing.T) {
	mockRepo := new(MockRedisRepository)
	// Create a dummy logger for testing
	logger := slog.Default()
	service := NewRedisService(mockRepo, logger)

	payload := &dto.SetCacheRequest{
		Namespace: "ns",
		Key:       "k3",
		Value:     "val3",
		TTL:       0,
	}
	expectedKey := "ns:k3"

	mockRepo.On("SetCache", expectedKey, "val3", (*time.Duration)(nil)).Return(errors.New("set error"))

	ok, err := service.SetCache(payload)
	assert.Error(t, err)
	assert.False(t, ok)
	mockRepo.AssertExpectations(t)
}

func TestRedisService_InvalidateCache_Success(t *testing.T) {
	mockRepo := new(MockRedisRepository)
	// Create a dummy logger for testing
	logger := slog.Default()
	service := NewRedisService(mockRepo, logger)

	payload := &dto.DeleteCacheRequest{
		Namespace: "ns",
		Key:       "k4",
	}
	expectedKey := "ns:k4"

	mockRepo.On("InvalidateCache", expectedKey).Return(nil)

	ok, err := service.InvalidateCache(payload)
	assert.NoError(t, err)
	assert.True(t, ok)
	mockRepo.AssertExpectations(t)
}

func TestRedisService_InvalidateCache_Error(t *testing.T) {
	mockRepo := new(MockRedisRepository)
	// Create a dummy logger for testing
	logger := slog.Default()
	service := NewRedisService(mockRepo, logger)

	payload := &dto.DeleteCacheRequest{
		Namespace: "ns",
		Key:       "k5",
	}
	expectedKey := "ns:k5"

	mockRepo.On("InvalidateCache", expectedKey).Return(errors.New("delete error"))

	ok, err := service.InvalidateCache(payload)
	assert.Error(t, err)
	assert.False(t, ok)
	mockRepo.AssertExpectations(t)
}
