package service_test

import (
	"context"
	"encoding/json"
	"errors"
	"messaging_service/internal/config"
	"messaging_service/internal/modules/message_broker/mock"
	"messaging_service/internal/modules/message_broker/service"
	"messaging_service/pkg/logger"
	"messaging_service/pkg/observability"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	testifyMock "github.com/stretchr/testify/mock"
)

// Mock implementations
type MockConfigClient struct {
	testifyMock.Mock
}

func (m *MockConfigClient) FetchConfig(ctx context.Context) (*config.Config, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*config.Config), args.Error(1)
}

type MockRedisClient struct {
	testifyMock.Mock
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
	testifyMock.Mock
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

func TestNewConfigManager(t *testing.T) {
	mockConfigClient := new(MockConfigClient)
	mockRedisClient := new(MockRedisClient)
	mockLogger := new(MockLoggerTest)
	mockObs := &observability.ObservabilityStack{
		LoggerService: mockLogger,
	}

	// Create the environment config
	env := &config.Env{
		MESSAGING_SERVICE_REDIS_TTL: 24,
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
	// Setup a temporary value for config.SERVICE_NAME
	originalServiceName := config.SERVICE_NAME
	config.SERVICE_NAME = "test-service"
	defer func() {
		config.SERVICE_NAME = originalServiceName
	}()

	mockConfigClient := new(MockConfigClient)
	mockRedisClient := new(MockRedisClient)

	// Use the existing mock implementations from the mock package
	mockObs := &observability.ObservabilityStack{
		LoggerService:  logger.NewMockLogger(),
		TracerService:  mock.NewMockTracerService(),
		MetricsService: mock.NewMockMetricsService(),
	}

	// Make sure env is properly initialized
	env := &config.Env{
		MESSAGING_SERVICE_REDIS_TTL: 24,
	}

	configManager, _ := service.NewConfigManager(mockConfigClient, mockRedisClient, mockObs, env)

	ctx := context.Background()
	// Create a more complete mock config with all required fields
	mockCfg := &config.Config{
		KafkaBrokers:                  []string{"localhost:9092"},
		KafkaAutoCreateTopics:         "true",
		KafkaNumPartitions:            3,
		KafkaReplicationFactor:        3,
		KafkaBatchSize:                100,
		KafkaBatchBytes:               1048576,
		KafkaBatchTimeoutMs:           500,
		KafkaCompressionCodec:         "snappy",
		KafkaMaxAttempts:              3,
		KafkaRetryBackoffMs:           100,
		KafkaReadTimeoutMs:            5000,
		KafkaWriteTimeoutMs:           5000,
		KafkaRequireActiveListener:    true,
		KafkaRetentionMs:              3000,
		KafkaConsumerMaxWaitMs:        5000,
		KafkaConsumerCommitIntervalMs: 5000,
		KafkaConsumerSessionTimeoutMs: 30000,
		KafkaConsumerHeartbeatMs:      1000,
		KafkaConsumerMaxPollRecords:   1000,
		KafkaConsumerAutoOffsetReset:  "earliest",
		KafkaEnableAutoCommit:         false,
		KafkaIsolationLevel:           "read_committed",
		ObservabilityUrl:              "http://localhost:4317",
	}

	t.Run("Successful API fetch", func(t *testing.T) {
		// Reset mocks
		mockConfigClient = new(MockConfigClient)
		mockRedisClient = new(MockRedisClient)

		// Use the existing mock implementations from the mock package
		mockObs = &observability.ObservabilityStack{
			LoggerService:  logger.NewMockLogger(),
			TracerService:  mock.NewMockTracerService(),
			MetricsService: mock.NewMockMetricsService(),
		}

		configManager, _ = service.NewConfigManager(mockConfigClient, mockRedisClient, mockObs, env)

		// Setup expectations
		mockConfigClient.On("FetchConfig", testifyMock.Anything).Return(mockCfg, nil)

		// Add mock for SetCache
		mockRedisClient.On("SetCache", testifyMock.Anything, testifyMock.Anything, testifyMock.Anything, testifyMock.Anything, testifyMock.Anything, testifyMock.Anything).Return(nil)

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

		// Use the existing mock implementations from the mock package
		mockObs = &observability.ObservabilityStack{
			LoggerService:  logger.NewMockLogger(),
			TracerService:  mock.NewMockTracerService(),
			MetricsService: mock.NewMockMetricsService(),
		}

		configManager, _ = service.NewConfigManager(mockConfigClient, mockRedisClient, mockObs, env)

		apiErr := errors.New("API fetch failed")
		mockConfigClient.On("FetchConfig", testifyMock.Anything).Return(nil, apiErr)

		// Use the same mockCfg from the outer function which now includes ObservabilityUrl
		mockCfgJSON, _ := json.Marshal(mockCfg)
		mockRedisClient.On("GetCache", testifyMock.Anything, testifyMock.Anything, testifyMock.Anything, testifyMock.Anything).Return(string(mockCfgJSON), true, nil)

		// Add mock for SetCache
		mockRedisClient.On("SetCache", testifyMock.Anything, testifyMock.Anything, testifyMock.Anything, testifyMock.Anything, testifyMock.Anything, testifyMock.Anything).Return(nil)

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

		// Use the existing mock implementations from the mock package
		mockObs = &observability.ObservabilityStack{
			LoggerService:  logger.NewMockLogger(),
			TracerService:  mock.NewMockTracerService(),
			MetricsService: mock.NewMockMetricsService(),
		}

		configManager, _ = service.NewConfigManager(mockConfigClient, mockRedisClient, mockObs, env)

		apiErr := errors.New("API fetch failed")
		cacheErr := errors.New("Cache fetch failed")

		mockConfigClient.On("FetchConfig", testifyMock.Anything).Return(nil, apiErr)
		mockRedisClient.On("GetCache", testifyMock.Anything, testifyMock.Anything, testifyMock.Anything, testifyMock.Anything).Return("", false, cacheErr)

		data, err := configManager.GetFromApiConfiguration(ctx)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "CFG001")
		assert.Nil(t, data)
		mockConfigClient.AssertExpectations(t)
		mockRedisClient.AssertExpectations(t)
	})

	t.Run("API fetch succeeds but SetConfig fails", func(t *testing.T) {
		// Reset mocks
		mockConfigClient = new(MockConfigClient)
		mockRedisClient = new(MockRedisClient)

		// Use the existing mock implementations from the mock package
		mockObs = &observability.ObservabilityStack{
			LoggerService:  logger.NewMockLogger(),
			TracerService:  mock.NewMockTracerService(),
			MetricsService: mock.NewMockMetricsService(),
		}

		configManager, _ = service.NewConfigManager(mockConfigClient, mockRedisClient, mockObs, env)

		// Create an invalid config to trigger validation error
		invalidCfg := &config.Config{
			// Missing required fields
		}

		mockConfigClient.On("FetchConfig", testifyMock.Anything).Return(invalidCfg, nil)

		data, err := configManager.GetFromApiConfiguration(ctx)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "CFG002")
		assert.Equal(t, invalidCfg, data)
		mockConfigClient.AssertExpectations(t)
	})

	t.Run("API fetch succeeds but cache fails", func(t *testing.T) {
		// Reset mocks
		mockConfigClient = new(MockConfigClient)
		mockRedisClient = new(MockRedisClient)

		// Use the existing mock implementations from the mock package
		mockObs = &observability.ObservabilityStack{
			LoggerService:  logger.NewMockLogger(),
			TracerService:  mock.NewMockTracerService(),
			MetricsService: mock.NewMockMetricsService(),
		}

		configManager, _ = service.NewConfigManager(mockConfigClient, mockRedisClient, mockObs, env)

		// Setup expectations
		mockConfigClient.On("FetchConfig", testifyMock.Anything).Return(mockCfg, nil)

		// Make SetCache fail
		cacheErr := errors.New("Cache set failed")
		mockRedisClient.On("SetCache", testifyMock.Anything, testifyMock.Anything, testifyMock.Anything, testifyMock.Anything, testifyMock.Anything, testifyMock.Anything).Return(cacheErr)

		data, err := configManager.GetFromApiConfiguration(ctx)

		// Should still succeed since cache failure is not critical
		assert.NoError(t, err)
		assert.Equal(t, mockCfg, data)
		mockConfigClient.AssertExpectations(t)
		mockRedisClient.AssertExpectations(t)
	})
}

func TestGetDataToCache(t *testing.T) {
	// Setup a temporary value for config.SERVICE_NAME
	originalServiceName := config.SERVICE_NAME
	config.SERVICE_NAME = "test-service"
	defer func() {
		config.SERVICE_NAME = originalServiceName
	}()

	mockConfigClient := new(MockConfigClient)
	mockRedisClient := new(MockRedisClient)

	// Use the existing mock implementations from the mock package
	mockObs := &observability.ObservabilityStack{
		LoggerService:  logger.NewMockLogger(),
		TracerService:  mock.NewMockTracerService(),
		MetricsService: mock.NewMockMetricsService(),
	}

	env := &config.Env{
		MESSAGING_SERVICE_REDIS_TTL: 24,
	}

	configManager, _ := service.NewConfigManager(mockConfigClient, mockRedisClient, mockObs, env)

	ctx := context.Background()
	// Create a more complete mock config with all required fields
	mockCfg := &config.Config{
		KafkaBrokers:                  []string{"localhost:9092"},
		KafkaAutoCreateTopics:         "true",
		KafkaNumPartitions:            3,
		KafkaReplicationFactor:        3,
		KafkaBatchSize:                100,
		KafkaBatchBytes:               1048576,
		KafkaBatchTimeoutMs:           500,
		KafkaCompressionCodec:         "snappy",
		KafkaMaxAttempts:              3,
		KafkaRetryBackoffMs:           100,
		KafkaReadTimeoutMs:            5000,
		KafkaWriteTimeoutMs:           5000,
		KafkaRequireActiveListener:    true,
		KafkaRetentionMs:              3000,
		KafkaConsumerMaxWaitMs:        5000,
		KafkaConsumerCommitIntervalMs: 5000,
		KafkaConsumerSessionTimeoutMs: 30000,
		KafkaConsumerHeartbeatMs:      1000,
		KafkaConsumerMaxPollRecords:   1000,
		KafkaConsumerAutoOffsetReset:  "earliest",
		KafkaEnableAutoCommit:         false,
		KafkaIsolationLevel:           "read_committed",
		ObservabilityUrl:              "http://localhost:4317",
	}

	// Test successful cache get
	t.Run("Successful cache get", func(t *testing.T) {
		mockCfgJSON, _ := json.Marshal(mockCfg)
		mockRedisClient.On("GetCache", testifyMock.Anything, testifyMock.Anything, testifyMock.Anything, testifyMock.Anything).
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
		mockRedisClient.On("GetCache", testifyMock.Anything, testifyMock.Anything, testifyMock.Anything, testifyMock.Anything).
			Return("", false, cacheErr).Once()

		data, err := configManager.GetDataToCache(ctx, "test-key")

		assert.Error(t, err)
		assert.Nil(t, data)
		assert.Contains(t, err.Error(), "CAC003")
		mockRedisClient.AssertExpectations(t)
	})

	// Test cache not found
	t.Run("Cache not found", func(t *testing.T) {
		mockRedisClient.On("GetCache", testifyMock.Anything, testifyMock.Anything, testifyMock.Anything, testifyMock.Anything).
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
		mockRedisClient.On("GetCache", testifyMock.Anything, testifyMock.Anything, testifyMock.Anything, testifyMock.Anything).
			Return("invalid json", true, nil).Once()

		data, err := configManager.GetDataToCache(ctx, "test-key")

		assert.Error(t, err)
		assert.Nil(t, data)
		assert.Contains(t, err.Error(), "CFG004")
		mockRedisClient.AssertExpectations(t)
	})
}

func TestSetDataToCache(t *testing.T) {
	// Setup a temporary value for config.SERVICE_NAME
	originalServiceName := config.SERVICE_NAME
	config.SERVICE_NAME = "test-service"
	defer func() {
		config.SERVICE_NAME = originalServiceName
	}()

	mockConfigClient := new(MockConfigClient)
	mockRedisClient := new(MockRedisClient)

	// Use the existing mock implementations from the mock package
	mockObs := &observability.ObservabilityStack{
		LoggerService:  logger.NewMockLogger(),
		TracerService:  mock.NewMockTracerService(),
		MetricsService: mock.NewMockMetricsService(),
	}

	env := &config.Env{
		MESSAGING_SERVICE_REDIS_TTL: 24,
	}

	configManager, _ := service.NewConfigManager(mockConfigClient, mockRedisClient, mockObs, env)

	ctx := context.Background()
	// Create a more complete mock config with all required fields
	mockCfg := &config.Config{
		KafkaBrokers:                  []string{"localhost:9092"},
		KafkaAutoCreateTopics:         "true",
		KafkaNumPartitions:            3,
		KafkaReplicationFactor:        3,
		KafkaBatchSize:                100,
		KafkaBatchBytes:               1048576,
		KafkaBatchTimeoutMs:           500,
		KafkaCompressionCodec:         "snappy",
		KafkaMaxAttempts:              3,
		KafkaRetryBackoffMs:           100,
		KafkaReadTimeoutMs:            5000,
		KafkaWriteTimeoutMs:           5000,
		KafkaRequireActiveListener:    true,
		KafkaRetentionMs:              3000,
		KafkaConsumerMaxWaitMs:        5000,
		KafkaConsumerCommitIntervalMs: 5000,
		KafkaConsumerSessionTimeoutMs: 30000,
		KafkaConsumerHeartbeatMs:      1000,
		KafkaConsumerMaxPollRecords:   1000,
		KafkaConsumerAutoOffsetReset:  "earliest",
		KafkaEnableAutoCommit:         false,
		KafkaIsolationLevel:           "read_committed",
		ObservabilityUrl:              "http://localhost:4317",
	}

	// Test successful cache set
	t.Run("Successful cache set", func(t *testing.T) {
		mockRedisClient.On("SetCache", testifyMock.Anything, testifyMock.Anything, testifyMock.Anything, testifyMock.Anything, testifyMock.Anything, testifyMock.Anything).
			Return(nil).Once()

		// We don't need Debug or Error mock expectations anymore since we ignore them in the mock

		err := configManager.SetDataToCache(ctx, "test-key", mockCfg)

		assert.NoError(t, err)
		mockRedisClient.AssertExpectations(t)
	})

	// Test cache set error
	t.Run("Cache set error", func(t *testing.T) {
		cacheErr := errors.New("Cache set failed")
		mockRedisClient.On("SetCache", testifyMock.Anything, testifyMock.Anything, testifyMock.Anything, testifyMock.Anything, testifyMock.Anything, testifyMock.Anything).
			Return(cacheErr).Once()

		// We don't need Debug or Error mock expectations anymore since we ignore them in the mock

		err := configManager.SetDataToCache(ctx, "test-key", mockCfg)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "CAC002")
		mockRedisClient.AssertExpectations(t)
	})

	// Test with a full mock of dependencies to cover the marshal error path
	t.Run("Full mock for marshal error", func(t *testing.T) {
		// Create mock implementations with full control
		mockConfigClient := new(MockConfigClient)
		mockRedisClient := new(MockRedisClient)

		// Create observability with our complete control
		mockLogger := new(MockLoggerTest)
		mockTracer := mock.NewMockTracerService()
		mockMetrics := mock.NewMockMetricsService()

		// Log expectations - all of these will be called
		mockLogger.On("Info", testifyMock.Anything, testifyMock.Anything).Return()
		mockLogger.On("Debug", testifyMock.Anything, testifyMock.Anything).Return()
		mockLogger.On("Error", testifyMock.Anything, testifyMock.Anything).Return()

		// Create the stack with our mocks
		mockObs := &observability.ObservabilityStack{
			LoggerService:  mockLogger,
			TracerService:  mockTracer,
			MetricsService: mockMetrics,
		}

		// Create the service with mocks
		service, _ := service.NewConfigManager(mockConfigClient, mockRedisClient, mockObs, env)

		// JSON marshal will succeed with any config (it's hard to make it fail naturally)
		// So we're testing the error path that comes after, the SetCache error
		mockRedisClient.On("SetCache", testifyMock.Anything, testifyMock.Anything, testifyMock.Anything, testifyMock.Anything, testifyMock.Anything, testifyMock.Anything).
			Return(errors.New("mocked cache error"))

		// Simple config
		cfg := &config.Config{
			KafkaBrokers:     []string{"localhost:9092"},
			ObservabilityUrl: "http://localhost:4317",
		}

		// Call SetDataToCache
		err := service.SetDataToCache(context.Background(), "test-key", cfg)

		// Verify the error
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "CAC002")
		mockRedisClient.AssertExpectations(t)
	})

	// Test by mocking the jsonMarshal function
	t.Run("Json Marshal error", func(t *testing.T) {
		// Save the original jsonMarshal function
		originalJsonMarshal := service.GetJsonMarshalFunc()

		// Replace with a mock that returns an error
		service.SetJsonMarshalFunc(func(v interface{}) ([]byte, error) {
			return nil, errors.New("mocked marshal error")
		})

		// Restore the original function when we're done
		defer service.SetJsonMarshalFunc(originalJsonMarshal)

		// Create mocks
		mockConfigClient := new(MockConfigClient)
		mockRedisClient := new(MockRedisClient)

		// Create observability with our complete control
		mockLogger := new(MockLoggerTest)

		// Setup logger expectations - with Anything matchers to handle dynamic values
		mockLogger.On("Info", testifyMock.Anything, testifyMock.Anything).Return()
		mockLogger.On("Debug", testifyMock.Anything, testifyMock.Anything).Return()
		mockLogger.On("Error", testifyMock.Anything, testifyMock.Anything).Return()
		mockLogger.On("Warn", testifyMock.Anything, testifyMock.Anything).Return()

		mockTracer := mock.NewMockTracerService()
		mockMetrics := mock.NewMockMetricsService()

		// Create the stack with our mocks
		mockObs := &observability.ObservabilityStack{
			LoggerService:  mockLogger,
			TracerService:  mockTracer,
			MetricsService: mockMetrics,
		}

		// Create the service with mocks
		configManager, _ := service.NewConfigManager(mockConfigClient, mockRedisClient, mockObs, env)

		// Simple config
		cfg := &config.Config{
			KafkaBrokers:     []string{"localhost:9092"},
			ObservabilityUrl: "http://localhost:4317",
		}

		// Call SetDataToCache - this should fail because of our mocked jsonMarshal
		err := configManager.SetDataToCache(context.Background(), "test-key", cfg)

		// Verify the error
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "CFG003")
	})
}

// Test SetEnvironment function
func TestSetEnvironment(t *testing.T) {
	mockConfigClient := new(MockConfigClient)
	mockRedisClient := new(MockRedisClient)

	// Use the existing mock implementations from the mock package
	mockObs := &observability.ObservabilityStack{
		LoggerService:  logger.NewMockLogger(),
		TracerService:  mock.NewMockTracerService(),
		MetricsService: mock.NewMockMetricsService(),
	}

	env := &config.Env{
		MESSAGING_SERVICE_REDIS_TTL: 24,
	}

	configManager, _ := service.NewConfigManager(mockConfigClient, mockRedisClient, mockObs, env)

	// Set a new environment config
	newEnv := &config.Env{
		MESSAGING_SERVICE_REDIS_TTL: 48,
	}
	configManager.SetEnvironment(newEnv)

	// Verify that the new environment was set by making a call that uses it
	mockRedisClient.On("SetCache", testifyMock.Anything, testifyMock.Anything, testifyMock.Anything, testifyMock.Anything, testifyMock.Anything, testifyMock.Anything).
		Return(nil).Once()

	// Create a simple config for the test
	simpleConfig := &config.Config{
		KafkaBrokers: []string{"localhost:9092"},
	}

	// Setup a temporary value for config.SERVICE_NAME
	originalServiceName := config.SERVICE_NAME
	config.SERVICE_NAME = "test-service"
	defer func() {
		config.SERVICE_NAME = originalServiceName
	}()

	// The TTL should now be 48 hours instead of 24
	err := configManager.SetDataToCache(context.Background(), "test-key", simpleConfig)
	assert.NoError(t, err)

	// Verify the mock was called with the expected TTL (48 hours)
	mockRedisClient.AssertCalled(t, "SetCache", testifyMock.Anything, testifyMock.Anything, testifyMock.Anything, testifyMock.Anything,
		time.Duration(48)*time.Hour, testifyMock.Anything)
}
