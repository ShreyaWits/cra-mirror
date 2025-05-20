package redis

import (
	"errors"
	"testing"
	"time"

	"context"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockRedisClient implements RedisClientInterface for testing
type MockRedisClient struct {
	mock.Mock
}

func (m *MockRedisClient) Get(ctx context.Context, key string) *redis.StringCmd {
	args := m.Called(ctx, key)
	cmd := redis.NewStringResult(args.String(0), args.Error(1))
	return cmd
}

func (m *MockRedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	args := m.Called(ctx, key, value, expiration)
	cmd := redis.NewStatusResult(args.String(0), args.Error(1))
	return cmd
}

func (m *MockRedisClient) Del(ctx context.Context, keys ...string) *redis.IntCmd {
	args := m.Called(ctx, keys)
	cmd := redis.NewIntResult(int64(args.Int(0)), args.Error(1))
	return cmd
}

func NewMockRedis(mockClient *MockRedisClient) *RedisClient {
	return &RedisClient{
		client: mockClient,
	}
}

func TestGetCache_KeyExists(t *testing.T) {
	mockClient := new(MockRedisClient)
	key := "foo"
	value := "bar"
	mockClient.On("Get", mock.Anything, key).Return(value, nil)

	redisClient := NewMockRedis(mockClient)
	got, err := redisClient.GetCache(key)
	assert.NoError(t, err)
	assert.Equal(t, value, got)
}

func TestGetCache_KeyDoesNotExist(t *testing.T) {
	mockClient := new(MockRedisClient)
	key := "missing"
	mockClient.On("Get", mock.Anything, key).Return("", redis.Nil)

	redisClient := NewMockRedis(mockClient)
	got, err := redisClient.GetCache(key)
	assert.NoError(t, err)
	assert.Equal(t, "", got)
}

func TestGetCache_Error(t *testing.T) {
	mockClient := new(MockRedisClient)
	key := "foo"
	mockClient.On("Get", mock.Anything, key).Return("", errors.New("connection error"))

	redisClient := NewMockRedis(mockClient)
	got, err := redisClient.GetCache(key)
	assert.Error(t, err)
	assert.Equal(t, "", got)
}

func TestSetCache_WithExpiration(t *testing.T) {
	mockClient := new(MockRedisClient)
	key := "foo"
	value := "bar"
	exp := 10 * time.Second
	mockClient.On("Set", mock.Anything, key, value, exp).Return("OK", nil)

	redisClient := NewMockRedis(mockClient)
	err := redisClient.SetCache(key, value, &exp)
	assert.NoError(t, err)
}

func TestSetCache_WithoutExpiration(t *testing.T) {
	mockClient := new(MockRedisClient)
	key := "foo"
	value := "bar"
	exp := 0 * time.Second
	mockClient.On("Set", mock.Anything, key, value, exp).Return("OK", nil)

	redisClient := NewMockRedis(mockClient)
	err := redisClient.SetCache(key, value, nil)
	assert.NoError(t, err)
}

func TestSetCache_Error(t *testing.T) {
	mockClient := new(MockRedisClient)
	key := "foo"
	value := "bar"
	exp := 0 * time.Second
	mockClient.On("Set", mock.Anything, key, value, exp).Return("", errors.New("set error"))

	redisClient := NewMockRedis(mockClient)
	err := redisClient.SetCache(key, value, nil)
	assert.Error(t, err)
}

func TestInvalidateCache_Success(t *testing.T) {
	mockClient := new(MockRedisClient)
	key := "foo"
	mockClient.On("Del", mock.Anything, []string{key}).Return(1, nil)

	redisClient := NewMockRedis(mockClient)
	err := redisClient.InvalidateCache(key)
	assert.NoError(t, err)
}

func TestInvalidateCache_Error(t *testing.T) {
	mockClient := new(MockRedisClient)
	key := "foo"
	mockClient.On("Del", mock.Anything, []string{key}).Return(0, errors.New("del error"))

	redisClient := NewMockRedis(mockClient)
	err := redisClient.InvalidateCache(key)
	assert.Error(t, err)
}
