package repository

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockRedisInterface mocks the RedisInterface
type MockRedisInterface struct {
	mock.Mock
}

func (m *MockRedisInterface) GetCache(key string) (string, error) {
	args := m.Called(key)
	return args.String(0), args.Error(1)
}

func (m *MockRedisInterface) SetCache(key, value string, expiration *time.Duration) error {
	args := m.Called(key, value, expiration)
	return args.Error(0)
}

func (m *MockRedisInterface) InvalidateCache(key string) error {
	args := m.Called(key)
	return args.Error(0)
}

func TestRedisRepo_GetCache(t *testing.T) {
	mockRedis := new(MockRedisInterface)
	repo := NewRedisRepository(mockRedis)

	key := "test-key"
	expectedValue := "test-value"

	mockRedis.On("GetCache", key).Return(expectedValue, nil)

	val, err := repo.GetCache(key)
	assert.NoError(t, err)
	assert.Equal(t, expectedValue, val)
	mockRedis.AssertExpectations(t)
}

func TestRedisRepo_GetCache_Error(t *testing.T) {
	mockRedis := new(MockRedisInterface)
	repo := NewRedisRepository(mockRedis)

	key := "test-key"
	mockRedis.On("GetCache", key).Return("", errors.New("redis error"))

	val, err := repo.GetCache(key)
	assert.Error(t, err)
	assert.Empty(t, val)
	mockRedis.AssertExpectations(t)
}

func TestRedisRepo_SetCache(t *testing.T) {
	mockRedis := new(MockRedisInterface)
	repo := NewRedisRepository(mockRedis)

	key := "test-key"
	value := "test-value"
	ttl := 10 * time.Second

	mockRedis.On("SetCache", key, value, &ttl).Return(nil)

	err := repo.SetCache(key, value, &ttl)
	assert.NoError(t, err)
	mockRedis.AssertExpectations(t)
}

func TestRedisRepo_SetCache_Error(t *testing.T) {
	mockRedis := new(MockRedisInterface)
	repo := NewRedisRepository(mockRedis)

	key := "test-key"
	value := "test-value"
	ttl := 10 * time.Second

	mockRedis.On("SetCache", key, value, &ttl).Return(errors.New("set error"))

	err := repo.SetCache(key, value, &ttl)
	assert.Error(t, err)
	mockRedis.AssertExpectations(t)
}

func TestRedisRepo_InvalidateCache(t *testing.T) {
	mockRedis := new(MockRedisInterface)
	repo := NewRedisRepository(mockRedis)

	key := "test-key"

	mockRedis.On("InvalidateCache", key).Return(nil)

	err := repo.InvalidateCache(key)
	assert.NoError(t, err)
	mockRedis.AssertExpectations(t)
}

func TestRedisRepo_InvalidateCache_Error(t *testing.T) {
	mockRedis := new(MockRedisInterface)
	repo := NewRedisRepository(mockRedis)

	key := "test-key"

	mockRedis.On("InvalidateCache", key).Return(errors.New("invalidate error"))

	err := repo.InvalidateCache(key)
	assert.Error(t, err)
	mockRedis.AssertExpectations(t)
}
