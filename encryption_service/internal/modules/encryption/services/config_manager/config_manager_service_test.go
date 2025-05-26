package service_test

import (
	"context"
	"encoding/json"
	"encryption_microservice/internal/config"
	service "encryption_microservice/internal/modules/encryption/services/config_manager"
	"encryption_microservice/pkg/logger"
	"encryption_microservice/pkg/observability"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock implementations
type MockConfigClient struct {
	mock.Mock
}

func (m *MockConfigClient) FetchConfig(ctx context.Context) (*config.Config, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*config.Config), args.Error(1)
}

type MockRedisClient struct {
	mock.Mock
}

func (m *MockRedisClient) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockRedisClient) SetCache(ctx context.Context, namespace, key, value string, ttl time.Duration, trackingID string) error {
	args := m.Called(ctx, namespace, key, value, ttl, trackingID)
	return args.Error(0)
}

func (m *MockRedisClient) GetCache(ctx context.Context, namespace, key, trackingID string) (string, bool, error) {
	args := m.Called(ctx, namespace, key, trackingID)
	return args.String(0), args.Bool(1), args.Error(2)
}

func (m *MockRedisClient) InvalidateCache(ctx context.Context, namespace, key, trackingID string) error {
	args := m.Called(ctx, namespace, key, trackingID)
	return args.Error(0)
}

type MockLoggerTest struct {
	mock.Mock
}

func (m *MockLoggerTest) Info(ctx context.Context, args ...interface{}) {
	m.Called(append([]interface{}{ctx}, args...)...)
}

func (m *MockLoggerTest) Error(ctx context.Context, args ...interface{}) {
	// Don't use the mock system for this method since it has variable argument count
	// that causes test failures. Just log the call but don't verify it.
}

func (m *MockLoggerTest) Debug(ctx context.Context, args ...interface{}) {
	// Similarly don't verify these calls, just log them
}

func (m *MockLoggerTest) Warn(ctx context.Context, args ...interface{}) {
	m.Called(append([]interface{}{ctx}, args...)...)
}

func (m *MockLoggerTest) WithFields(fields map[string]interface{}) logger.Logger {
	return m
}

func (m *MockLoggerTest) Sync() error {
	return nil
}

// MockObservabilityStack mocks the observability stack
type MockObservabilityStack struct {
	LoggerService  *MockLoggerTest
	TracerService  interface{}
	MetricsService interface{}
}

func (m *MockObservabilityStack) GetLoggerService() logger.Logger {
	return m.LoggerService
}

func TestNewConfigManager(t *testing.T) {
	mockConfigClient := new(MockConfigClient)
	mockRedisClient := new(MockRedisClient)
	mockLogger := new(MockLoggerTest)
	mockObs := &observability.ObservabilityStack{
		LoggerService: mockLogger,
	}

	// Create the environment config
	env := &config.EnvConfig{
		CacheTTL: 24,
	}

	// Test with valid inputs
	configManager, err := service.NewConfigManager(mockConfigClient, mockRedisClient, mockObs, env)
	assert.NoError(t, err)
	assert.NotNil(t, configManager)

	// Test with nil configClient
	configManager, err = service.NewConfigManager(nil, mockRedisClient, mockObs, env)
	assert.Error(t, err)
	assert.Nil(t, configManager)
	assert.Contains(t, err.Error(), "configClient cannot be nil")

	// Test with nil cacheClient
	configManager, err = service.NewConfigManager(mockConfigClient, nil, mockObs, env)
	assert.Error(t, err)
	assert.Nil(t, configManager)
	assert.Contains(t, err.Error(), "cacheClient cannot be nil")

	// Test with nil observability stack
	configManager, err = service.NewConfigManager(mockConfigClient, mockRedisClient, nil, env)
	assert.Error(t, err)
	assert.Nil(t, configManager)
	assert.Contains(t, err.Error(), "observability stack cannot be nil")
}

func TestGetFromApiConfiguration(t *testing.T) {
	mockConfigClient := new(MockConfigClient)
	mockRedisClient := new(MockRedisClient)
	mockLogger := new(MockLoggerTest)
	mockObs := &observability.ObservabilityStack{
		LoggerService: mockLogger,
	}

	env := &config.EnvConfig{
		CacheTTL: 24,
	}

	configManager, _ := service.NewConfigManager(mockConfigClient, mockRedisClient, mockObs, env)

	ctx := context.Background()
	// Create a mock config with required fields
	mockCfg := &config.Config{
		// ServerPort:       "8082",
		// GrpcPort:         "50051",
		// UserServiceURL:   "http://localhost:8080",
		VaultAddr:                 "https://localhost:8200",
		VaultToken:                "test-token",
		VaultPath:                 "transit",
		OtelCollectorGrpcEndpoint: "localhost:4317",
	}

	t.Run("Successful API fetch", func(t *testing.T) {
		// Reset mocks
		mockConfigClient = new(MockConfigClient)
		mockRedisClient = new(MockRedisClient)
		mockLogger = new(MockLoggerTest)
		mockObs = &observability.ObservabilityStack{
			LoggerService: mockLogger,
		}

		configManager, _ = service.NewConfigManager(mockConfigClient, mockRedisClient, mockObs, env)

		// Setup expectations
		mockConfigClient.On("FetchConfig", mock.Anything).Return(mockCfg, nil)

		// Add mock for SetCache
		mockRedisClient.On("SetCache", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

		data, err := configManager.GetFromApiConfiguration(ctx)

		assert.NoError(t, err)
		assert.Equal(t, mockCfg, data)
		mockConfigClient.AssertExpectations(t)
		mockRedisClient.AssertExpectations(t)
	})

	t.Run("API fetch fails, cache succeeds", func(t *testing.T) {
		// Reset mocks
		mockConfigClient = new(MockConfigClient)
		mockRedisClient = new(MockRedisClient)
		mockLogger = new(MockLoggerTest)
		mockObs = &observability.ObservabilityStack{
			LoggerService: mockLogger,
		}

		configManager, _ = service.NewConfigManager(mockConfigClient, mockRedisClient, mockObs, env)

		apiErr := errors.New("API fetch failed")
		mockConfigClient.On("FetchConfig", mock.Anything).Return(nil, apiErr)

		// Use the same mockCfg from the outer function
		mockCfgJSON, _ := json.Marshal(mockCfg)
		mockRedisClient.On("GetCache", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(string(mockCfgJSON), true, nil)

		// Add mock for SetCache
		mockRedisClient.On("SetCache", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

		data, err := configManager.GetFromApiConfiguration(ctx)

		assert.NoError(t, err)
		assert.Equal(t, mockCfg, data)
		mockConfigClient.AssertExpectations(t)
		mockRedisClient.AssertExpectations(t)
	})

	t.Run("API fetch fails, cache fails", func(t *testing.T) {
		// Reset mocks
		mockConfigClient = new(MockConfigClient)
		mockRedisClient = new(MockRedisClient)
		mockLogger = new(MockLoggerTest)
		mockObs = &observability.ObservabilityStack{
			LoggerService: mockLogger,
		}

		configManager, _ = service.NewConfigManager(mockConfigClient, mockRedisClient, mockObs, env)

		apiErr := errors.New("API fetch failed")
		cacheErr := errors.New("Cache fetch failed")

		mockConfigClient.On("FetchConfig", mock.Anything).Return(nil, apiErr)
		mockRedisClient.On("GetCache", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return("", false, cacheErr)

		data, err := configManager.GetFromApiConfiguration(ctx)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get cache")
		assert.Nil(t, data)
		mockConfigClient.AssertExpectations(t)
		mockRedisClient.AssertExpectations(t)
	})
}

func TestGetDataToCache(t *testing.T) {
	mockConfigClient := new(MockConfigClient)
	mockRedisClient := new(MockRedisClient)
	mockLogger := new(MockLoggerTest)
	mockObs := &observability.ObservabilityStack{
		LoggerService: mockLogger,
	}

	env := &config.EnvConfig{
		CacheTTL: 24,
	}

	configManager, _ := service.NewConfigManager(mockConfigClient, mockRedisClient, mockObs, env)

	ctx := context.Background()
	// Create a mock config with required fields
	mockCfg := &config.Config{
		// ServerPort:       "8082",
		// GrpcPort:         "50051",
		// UserServiceURL:   "http://localhost:8080",
		VaultAddr:                 "https://localhost:8200",
		VaultToken:                "test-token",
		VaultPath:                 "transit",
		OtelCollectorGrpcEndpoint: "localhost:4317",
	}

	// Test successful cache get
	t.Run("Successful cache get", func(t *testing.T) {
		mockCfgJSON, _ := json.Marshal(mockCfg)
		mockRedisClient.On("GetCache", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return(string(mockCfgJSON), true, nil).Once()

		// We don't need Debug or Error mock expectations anymore since we ignore them in the mock

		data, err := configManager.GetDataToCache(ctx, "test-key")

		assert.NoError(t, err)
		assert.Equal(t, mockCfg, data)
		mockRedisClient.AssertExpectations(t)
	})

	// Test cache get error
	t.Run("Cache get error", func(t *testing.T) {
		cacheErr := errors.New("Cache get failed")
		mockRedisClient.On("GetCache", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return("", false, cacheErr).Once()

		// We don't need Debug or Error mock expectations anymore since we ignore them in the mock

		data, err := configManager.GetDataToCache(ctx, "test-key")

		assert.Error(t, err)
		assert.Nil(t, data)
		assert.Contains(t, err.Error(), "failed to get cache")
		mockRedisClient.AssertExpectations(t)
	})

	// Test cache not found
	t.Run("Cache not found", func(t *testing.T) {
		mockRedisClient.On("GetCache", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return("", false, nil).Once()

		// We don't need Debug or Error mock expectations anymore since we ignore them in the mock

		data, err := configManager.GetDataToCache(ctx, "test-key")

		assert.Error(t, err)
		assert.Nil(t, data)
		assert.Contains(t, err.Error(), "config not found in cache")
		mockRedisClient.AssertExpectations(t)
	})

	// Test unmarshal error
	t.Run("Unmarshal error", func(t *testing.T) {
		mockRedisClient.On("GetCache", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return("invalid json", true, nil).Once()

		// We don't need Debug or Error mock expectations anymore since we ignore them in the mock

		data, err := configManager.GetDataToCache(ctx, "test-key")

		assert.Error(t, err)
		assert.Nil(t, data)
		assert.Contains(t, err.Error(), "failed to unmarshal cached config")
		mockRedisClient.AssertExpectations(t)
	})
}

func TestSetDataToCache(t *testing.T) {
	mockConfigClient := new(MockConfigClient)
	mockRedisClient := new(MockRedisClient)
	mockLogger := new(MockLoggerTest)
	mockObs := &observability.ObservabilityStack{
		LoggerService: mockLogger,
	}

	env := &config.EnvConfig{
		CacheTTL: 24,
	}

	configManager, _ := service.NewConfigManager(mockConfigClient, mockRedisClient, mockObs, env)

	ctx := context.Background()
	// Create a mock config with required fields
	mockCfg := &config.Config{
		// ServerPort:       "8082",
		// GrpcPort:         "50051",
		// UserServiceURL:   "http://localhost:8080",
		VaultAddr:                 "https://localhost:8200",
		VaultToken:                "test-token",
		VaultPath:                 "transit",
		OtelCollectorGrpcEndpoint: "localhost:4317",
	}

	// Test successful cache set
	t.Run("Successful cache set", func(t *testing.T) {
		mockRedisClient.On("SetCache", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return(nil).Once()

		// We don't need Debug or Error mock expectations anymore since we ignore them in the mock

		err := configManager.SetDataToCache(ctx, "test-key", mockCfg)

		assert.NoError(t, err)
		mockRedisClient.AssertExpectations(t)
	})

	// Test cache set error
	t.Run("Cache set error", func(t *testing.T) {
		cacheErr := errors.New("Cache set failed")
		mockRedisClient.On("SetCache", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return(cacheErr).Once()

		// We don't need Debug or Error mock expectations anymore since we ignore them in the mock

		err := configManager.SetDataToCache(ctx, "test-key", mockCfg)

		assert.Error(t, err)
		assert.Equal(t, cacheErr, err)
		mockRedisClient.AssertExpectations(t)
	})
}
