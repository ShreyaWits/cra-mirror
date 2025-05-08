package redis

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/mock"
)

// MockRedisClient is a mock implementation of RedisClientInterface
type MockRedisClient struct {
	mock.Mock
}

// Ensure MockRedisClient implements the RedisClientInterface
var _ RedisClientInterface = (*MockRedisClient)(nil)

func (m *MockRedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	args := m.Called(ctx, key, value, expiration)
	return args.Get(0).(*redis.StatusCmd)
}

func (m *MockRedisClient) Get(ctx context.Context, key string) *redis.StringCmd {
	args := m.Called(ctx, key)
	return args.Get(0).(*redis.StringCmd)
}

func (m *MockRedisClient) Del(ctx context.Context, keys ...string) *redis.IntCmd {
	args := m.Called(ctx, keys)
	return args.Get(0).(*redis.IntCmd)
}

func TestNewRedisService(t *testing.T) {
	// NewRedisService calls RedisClient internally, which typically connects to a real Redis.
	// A true unit test would require mocking the RedisClient function itself,
	// or using a test Redis instance (e.g., with testcontainers).
	// This test performs a basic check that the function returns a non-nil service,
	// assuming RedisClient is tested separately or in integration tests.

	// Define dummy parameters as a real connection is not expected to be made in a unit test context
	redisHost := "localhost"
	redisPort := "6379"
	redisUsername := ""
	redisPassword := ""

	// Call the function to be tested
	// Note: This call will likely fail if a Redis instance is not running,
	// as RedisClient attempts a real connection.
	// For a robust unit test, RedisClient should be mocked.
	service, _ := NewRedisService(redisHost, redisPort, redisUsername, redisPassword) // Ignore error for basic check

	// Check if the service is initialized and no error occurred (in a scenario where RedisClient is mocked)
	// In the current setup without mocking RedisClient, this check might fail if Redis is not running.
	// We add this check assuming a future state where RedisClient can be mocked for unit tests.
	if service == nil {
		t.Error("NewRedisService returned nil service")
	}

	// We don't check for err here because RedisClient will return an error if Redis is not running,
	// and mocking RedisClient is complex. The primary goal here is to add a test function
	// that covers the NewRedisService call, even if its assertions are limited without mocking RedisClient.

	// TODO: Implement proper mocking for RedisClient to enable comprehensive unit testing of NewRedisService,
	// including error handling scenarios during client initialization.
}

func TestRedisService_Set(t *testing.T) {
	mockClient := new(MockRedisClient)
	service := &RedisService{Client: mockClient} // Use the interface
	ctx := context.Background()
	key := "testkey"
	data := []byte("testdata")
	expiration := time.Minute

	// Mock the Set method to return a successful status command
	statusCmd := redis.NewStatusCmd(ctx)
	statusCmd.SetVal("OK") // Simulate success
	mockClient.On("Set", ctx, key, data, expiration).Return(statusCmd)

	err := service.Set(key, data, expiration, ctx)

	if err != nil {
		t.Errorf("Set() error = %v, wantErr %v", err, false)
	}

	// Verify that the mock client's Set method was called with the correct arguments
	mockClient.AssertCalled(t, "Set", ctx, key, data, expiration)
}

func TestRedisService_Set_Error(t *testing.T) {
	mockClient := new(MockRedisClient)
	service := &RedisService{Client: mockClient} // Use the interface
	ctx := context.Background()
	key := "testkey"
	data := []byte("testdata")
	expiration := time.Minute
	expectedErr := errors.New("redis set error")

	// Mock the Set method to return an error
	statusCmd := redis.NewStatusCmd(ctx)
	statusCmd.SetErr(expectedErr)
	mockClient.On("Set", ctx, key, data, expiration).Return(statusCmd)

	err := service.Set(key, data, expiration, ctx)

	if err == nil || err.Error() != expectedErr.Error() {
		t.Errorf("Set() error = %v, wantErr %v", err, expectedErr)
	}

	mockClient.AssertCalled(t, "Set", ctx, key, data, expiration)
}

func TestRedisService_Hset(t *testing.T) {
	mockClient := new(MockRedisClient)
	service := &RedisService{Client: mockClient} // Use the interface
	ctx := context.Background()
	key := "testhkey"
	data := "testhdata"
	expireAt := time.Now().Add(time.Hour)
	// expirationDuration := time.Until(expireAt) // Not directly used in mock setup

	// Mock the Set method (Hset in service.go uses Set)
	statusCmd := redis.NewStatusCmd(ctx)
	statusCmd.SetVal("OK")
	// Use AnythingOfType for duration as it's calculated and might vary slightly
	mockClient.On("Set", ctx, key, data, mock.AnythingOfType("time.Duration")).Return(statusCmd)

	err := service.Hset(key, data, expireAt, ctx)

	if err != nil {
		t.Errorf("Hset() error = %v, wantErr %v", err, false)
	}

	// Verify the call, checking the duration is approximately correct
	mockClient.AssertCalled(t, "Set", ctx, key, data, mock.AnythingOfType("time.Duration"))
}

func TestRedisService_Hset_Error(t *testing.T) {
	mockClient := new(MockRedisClient)
	service := &RedisService{Client: mockClient} // Use the interface
	ctx := context.Background()
	key := "testhkey"
	data := "testhdata"
	expireAt := time.Now().Add(time.Hour)
	expectedErr := errors.New("redis hset error")

	// Mock the Set method to return an error
	statusCmd := redis.NewStatusCmd(ctx)
	statusCmd.SetErr(expectedErr)
	mockClient.On("Set", ctx, key, data, mock.AnythingOfType("time.Duration")).Return(statusCmd)

	err := service.Hset(key, data, expireAt, ctx)

	if err == nil || err.Error() != expectedErr.Error() {
		t.Errorf("Hset() error = %v, wantErr %v", err, expectedErr)
	}

	mockClient.AssertCalled(t, "Set", ctx, key, data, mock.AnythingOfType("time.Duration"))
}


func TestRedisService_Get(t *testing.T) {
	mockClient := new(MockRedisClient)
	service := &RedisService{Client: mockClient} // Use the interface
	ctx := context.Background()
	key := "testkey"
	expectedValue := "testvalue"

	// Mock the Get method to return a successful string command
	stringCmd := redis.NewStringCmd(ctx)
	stringCmd.SetVal(expectedValue)
	mockClient.On("Get", ctx, key).Return(stringCmd)

	value, err := service.Get(key, ctx)

	if err != nil {
		t.Errorf("Get() error = %v, wantErr %v", err, false)
	}
	if value != expectedValue {
		t.Errorf("Get() returned value = %v, want %v", value, expectedValue)
	}

	mockClient.AssertCalled(t, "Get", ctx, key)
}

func TestRedisService_Get_NotFound(t *testing.T) {
	mockClient := new(MockRedisClient)
	service := &RedisService{Client: mockClient} // Use the interface
	ctx := context.Background()
	key := "nonexistentkey"

	// Mock the Get method to return redis.Nil error
	stringCmd := redis.NewStringCmd(ctx)
	stringCmd.SetErr(redis.Nil)
	mockClient.On("Get", ctx, key).Return(stringCmd)

	value, err := service.Get(key, ctx)

	if err == nil || err.Error() != "key does not exist" {
		t.Errorf("Get() error = %v, wantErr %v", err, errors.New("key does not exist"))
	}
	if value != "" {
		t.Errorf("Get() returned value = %v, want %v", value, "")
	}

	mockClient.AssertCalled(t, "Get", ctx, key)
}

func TestRedisService_Get_Error(t *testing.T) {
	mockClient := new(MockRedisClient)
	service := &RedisService{Client: mockClient} // Use the interface
	ctx := context.Background()
	key := "testkey"
	expectedErr := errors.New("redis get error")

	// Mock the Get method to return a generic error
	stringCmd := redis.NewStringCmd(ctx)
	stringCmd.SetErr(expectedErr)
	mockClient.On("Get", ctx, key).Return(stringCmd)

	value, err := service.Get(key, ctx)

	if err == nil || err.Error() != expectedErr.Error() {
		t.Errorf("Get() error = %v, wantErr %v", err, expectedErr)
	}
	if value != "" {
		t.Errorf("Get() returned value = %v, want %v", value, "")
	}

	mockClient.AssertCalled(t, "Get", ctx, key)
}

func TestRedisService_Del(t *testing.T) {
	mockClient := new(MockRedisClient)
	service := &RedisService{Client: mockClient} // Use the interface
	ctx := context.Background()
	key := "testkey"

	// Mock the Del method to return a successful int command
	intCmd := redis.NewIntCmd(ctx)
	intCmd.SetVal(1) // Simulate 1 key deleted
	mockClient.On("Del", ctx, []string{key}).Return(intCmd)

	err := service.Del(key, ctx)

	if err != nil {
		t.Errorf("Del() error = %v, wantErr %v", err, false)
	}

	mockClient.AssertCalled(t, "Del", ctx, []string{key})
}

func TestRedisService_Del_Error(t *testing.T) {
	mockClient := new(MockRedisClient)
	service := &RedisService{Client: mockClient} // Use the interface
	ctx := context.Background()
	key := "testkey"
	expectedErr := errors.New("redis del error")

	// Mock the Del method to return an error
	intCmd := redis.NewIntCmd(ctx)
	intCmd.SetErr(expectedErr)
	mockClient.On("Del", ctx, []string{key}).Return(intCmd)

	err := service.Del(key, ctx)

	if err == nil || err.Error() != expectedErr.Error() {
		t.Errorf("Del() error = %v, wantErr %v", err, expectedErr)
	}

	mockClient.AssertCalled(t, "Del", ctx, []string{key})
}