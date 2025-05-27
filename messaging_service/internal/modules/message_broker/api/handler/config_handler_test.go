package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"messaging_service/internal/config"
	"messaging_service/internal/modules/message_broker/api/handler"
	"messaging_service/internal/modules/message_broker/service"
	"messaging_service/pkg/logger"
	"messaging_service/pkg/observability"
)

// Mock implementations for observability components
type MockTracerService struct{}

func (m *MockTracerService) StartTracer(ctx context.Context, name string) (context.Context, trace.Span) {
	return ctx, trace.SpanFromContext(ctx)
}

func (m *MockTracerService) StopSpan(span trace.Span) {}

func (m *MockTracerService) SetAttributes(span trace.Span, attrs map[string]string) {}

func (m *MockTracerService) SetStatus(span trace.Span, code codes.Code, message string) {}

func (m *MockTracerService) RecordError(span trace.Span, err error) {}

type MockMetricsService struct{}

func (m *MockMetricsService) IncrementCounter(ctx context.Context, metricName string, value int64, tags map[string]string) {
}

func (m *MockMetricsService) RecordHistogram(ctx context.Context, name string, value float64, attrs map[string]string) {
}

type MockLoggerService struct{}

func (m *MockLoggerService) Debug(ctx context.Context, args ...interface{})         {}
func (m *MockLoggerService) Info(ctx context.Context, args ...interface{})          {}
func (m *MockLoggerService) Warn(ctx context.Context, args ...interface{})          {}
func (m *MockLoggerService) Error(ctx context.Context, args ...interface{})         {}
func (m *MockLoggerService) WithFields(fields map[string]interface{}) logger.Logger { return m }
func (m *MockLoggerService) Sync() error                                            { return nil }

// Mock ConfigManagerService
type MockConfigManagerService struct {
	ctrl                *gomock.Controller
	setDataToCacheError error
}

// Ensure MockConfigManagerService implements the ConfigManagerServiceInterface
var _ service.ConfigManagerServiceInterface = (*MockConfigManagerService)(nil)

func NewMockConfigManagerService(ctrl *gomock.Controller) *MockConfigManagerService {
	return &MockConfigManagerService{ctrl: ctrl}
}

func (m *MockConfigManagerService) GetFromApiConfiguration(ctx context.Context) (*config.Config, error) {
	return nil, nil
}

func (m *MockConfigManagerService) SetDataToCache(ctx context.Context, key string, cfg *config.Config) error {
	if m.setDataToCacheError != nil {
		return m.setDataToCacheError
	}
	return nil
}

func (m *MockConfigManagerService) GetDataToCache(ctx context.Context, key string) (*config.Config, error) {
	return nil, nil
}

// SetSetDataToCacheError configures the mock to return an error when SetDataToCache is called
func (m *MockConfigManagerService) SetSetDataToCacheError(err error) {
	m.setDataToCacheError = err
}

// Setup function to create a ConfigHandler for testing
func setupConfigHandler(t *testing.T) (*handler.ConfigHandler, *MockConfigManagerService, *gomock.Controller) {
	ctrl := gomock.NewController(t)
	mockSvc := NewMockConfigManagerService(ctrl)

	env := &config.Env{

		ConfigServiceUrl: "http://localhost:8080",
	}

	// Create mock observability stack
	obs := &observability.ObservabilityStack{
		TracerService:  &MockTracerService{},
		MetricsService: &MockMetricsService{},
		LoggerService:  &MockLoggerService{},
	}

	h, _ := handler.NewConfigHandler(mockSvc, env, obs)

	return h, mockSvc, ctrl
}

func TestNewConfigHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := NewMockConfigManagerService(ctrl)
	validEnv := &config.Env{
		ConfigServiceUrl: "http://localhost:8080",
	}

	// Create mock observability stack
	obs := &observability.ObservabilityStack{
		TracerService:  &MockTracerService{},
		MetricsService: &MockMetricsService{},
		LoggerService:  &MockLoggerService{},
	}

	tests := []struct {
		name        string
		svc         *MockConfigManagerService
		env         *config.Env
		obs         *observability.ObservabilityStack
		expectError bool
	}{
		{
			name:        "valid parameters",
			svc:         mockSvc,
			env:         validEnv,
			obs:         obs,
			expectError: false,
		},
		{
			name:        "nil service",
			svc:         nil,
			env:         validEnv,
			obs:         obs,
			expectError: true,
		},
		{
			name:        "nil env",
			svc:         mockSvc,
			env:         nil,
			obs:         obs,
			expectError: true,
		},
		{
			name:        "nil observability stack",
			svc:         mockSvc,
			env:         validEnv,
			obs:         nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var svcInterface service.ConfigManagerServiceInterface
			if tt.svc != nil {
				svcInterface = tt.svc
			}

			h, err := handler.NewConfigHandler(svcInterface, tt.env, tt.obs)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, h)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, h)
			}
		})
	}
}

func TestSetReinitCallback(t *testing.T) {
	h, _, ctrl := setupConfigHandler(t)
	defer ctrl.Finish()

	callbackExecuted := false
	callback := func(configuration *config.Config) error {
		callbackExecuted = true
		return nil
	}

	// Set the callback
	h.SetReinitCallback(callback)

	// Create a fiber app for testing
	app := fiber.New()
	app.Post("/config/update", h.UpdateConfigurations)

	// Create a test request with a valid configuration
	testConfig := &config.Config{
		KafkaBrokers:           []string{"localhost:9092"},
		KafkaNumPartitions:     3,
		KafkaReplicationFactor: 3,
		KafkaBatchSize:         100,
		KafkaBatchBytes:        1048576,
		KafkaBatchTimeoutMs:    500,
		KafkaCompressionCodec:  "snappy",
		KafkaMaxAttempts:       3,
		KafkaRetryBackoffMs:    100,
		KafkaReadTimeoutMs:     5000,
		KafkaWriteTimeoutMs:    5000,
		KafkaRetentionMs:       3000,

		KafkaConsumerMaxWaitMs:        5000,
		KafkaConsumerCommitIntervalMs: 5000,
		KafkaConsumerSessionTimeoutMs: 30000,
		KafkaConsumerHeartbeatMs:      1000,
		KafkaConsumerMaxPollRecords:   1000,
		KafkaConsumerAutoOffsetReset:  "earliest",
		KafkaIsolationLevel:           "read_committed",
	}

	// Wrap the config in the expected structure
	requestWrapper := struct {
		Environment string         `json:"environment"`
		Method      string         `json:"method"`
		ServiceName string         `json:"serviceName"`
		Values      *config.Config `json:"values"`
	}{
		Environment: "test",
		Method:      "update",
		ServiceName: "messaging_service",
		Values:      testConfig,
	}

	reqBody, _ := json.Marshal(requestWrapper)
	req := httptest.NewRequest("POST", "/config/update", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	// Execute the request
	resp, _ := app.Test(req)

	// Verify response
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	assert.True(t, callbackExecuted, "Callback should have been executed")
}

func TestUpdateConfigurations(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name              string
		requestBody       interface{}
		setCallback       bool
		callbackReturns   error
		setupMock         func(*MockConfigManagerService)
		expectedStatus    int
		expectedHasConfig bool
	}{
		{
			name: "valid config update",
			requestBody: struct {
				Environment string         `json:"environment"`
				Method      string         `json:"method"`
				ServiceName string         `json:"serviceName"`
				Values      *config.Config `json:"values"`
			}{
				Environment: "test",
				Method:      "update",
				ServiceName: "messaging_service",
				Values: &config.Config{
					KafkaBrokers:           []string{"localhost:9092"},
					KafkaNumPartitions:     3,
					KafkaReplicationFactor: 3,
					KafkaBatchSize:         100,
					KafkaBatchBytes:        1048576,
					KafkaBatchTimeoutMs:    500,
					KafkaCompressionCodec:  "snappy",
					KafkaMaxAttempts:       3,
					KafkaRetryBackoffMs:    100,
					KafkaReadTimeoutMs:     5000,
					KafkaWriteTimeoutMs:    5000,
					KafkaRetentionMs:       3000,

					KafkaConsumerMaxWaitMs:        5000,
					KafkaConsumerCommitIntervalMs: 5000,
					KafkaConsumerSessionTimeoutMs: 30000,
					KafkaConsumerHeartbeatMs:      1000,
					KafkaConsumerMaxPollRecords:   1000,
					KafkaConsumerAutoOffsetReset:  "earliest",
					KafkaIsolationLevel:           "read_committed",
				},
			},
			setCallback:       false,
			setupMock:         func(mock *MockConfigManagerService) {},
			expectedStatus:    fiber.StatusOK,
			expectedHasConfig: true,
		},
		{
			name: "valid config update with successful callback",
			requestBody: struct {
				Environment string         `json:"environment"`
				Method      string         `json:"method"`
				ServiceName string         `json:"serviceName"`
				Values      *config.Config `json:"values"`
			}{
				Environment: "test",
				Method:      "update",
				ServiceName: "messaging_service",
				Values: &config.Config{
					KafkaBrokers:           []string{"localhost:9092"},
					KafkaNumPartitions:     3,
					KafkaReplicationFactor: 3,
					KafkaBatchSize:         100,
					KafkaBatchBytes:        1048576,
					KafkaBatchTimeoutMs:    500,
					KafkaCompressionCodec:  "snappy",
					KafkaMaxAttempts:       3,
					KafkaRetryBackoffMs:    100,
					KafkaReadTimeoutMs:     5000,
					KafkaWriteTimeoutMs:    5000,
					KafkaRetentionMs:       3000,

					KafkaConsumerMaxWaitMs:        5000,
					KafkaConsumerCommitIntervalMs: 5000,
					KafkaConsumerSessionTimeoutMs: 30000,
					KafkaConsumerHeartbeatMs:      1000,
					KafkaConsumerMaxPollRecords:   1000,
					KafkaConsumerAutoOffsetReset:  "earliest",
					KafkaIsolationLevel:           "read_committed",
				},
			},
			setCallback:       true,
			callbackReturns:   nil,
			setupMock:         func(mock *MockConfigManagerService) {},
			expectedStatus:    fiber.StatusOK,
			expectedHasConfig: true,
		},
		{
			name:              "invalid request body",
			requestBody:       "this is not a valid json object",
			setCallback:       false,
			setupMock:         func(mock *MockConfigManagerService) {},
			expectedStatus:    fiber.StatusBadRequest,
			expectedHasConfig: false,
		},
		{
			name: "missing values in request",
			requestBody: struct {
				Environment string `json:"environment"`
				Method      string `json:"method"`
				ServiceName string `json:"serviceName"`
				// Values is missing
			}{
				Environment: "test",
				Method:      "update",
				ServiceName: "messaging_service",
			},
			setCallback:       false,
			setupMock:         func(mock *MockConfigManagerService) {},
			expectedStatus:    fiber.StatusBadRequest,
			expectedHasConfig: false,
		},
		{
			name: "cache operation failure",
			requestBody: struct {
				Environment string         `json:"environment"`
				Method      string         `json:"method"`
				ServiceName string         `json:"serviceName"`
				Values      *config.Config `json:"values"`
			}{
				Environment: "test",
				Method:      "update",
				ServiceName: "messaging_service",
				Values: &config.Config{
					KafkaBrokers:           []string{"localhost:9092"},
					KafkaNumPartitions:     3,
					KafkaReplicationFactor: 3,
				},
			},
			setCallback: false,
			setupMock: func(mock *MockConfigManagerService) {
				// Configure mock to return an error for SetDataToCache
				mock.SetSetDataToCacheError(errors.New("cache error"))
			},
			expectedStatus:    fiber.StatusOK, // The handler still returns OK even if cache fails
			expectedHasConfig: true,
		},
		{
			name: "failed callback",
			requestBody: struct {
				Environment string         `json:"environment"`
				Method      string         `json:"method"`
				ServiceName string         `json:"serviceName"`
				Values      *config.Config `json:"values"`
			}{
				Environment: "test",
				Method:      "update",
				ServiceName: "messaging_service",
				Values: &config.Config{
					KafkaBrokers:           []string{"localhost:9092"},
					KafkaNumPartitions:     3,
					KafkaReplicationFactor: 3,
					KafkaBatchSize:         100,
					KafkaBatchBytes:        1048576,
					KafkaBatchTimeoutMs:    500,
					KafkaCompressionCodec:  "snappy",
					KafkaMaxAttempts:       3,
					KafkaRetryBackoffMs:    100,
					KafkaReadTimeoutMs:     5000,
					KafkaWriteTimeoutMs:    5000,
					KafkaRetentionMs:       3000,

					KafkaConsumerMaxWaitMs:        5000,
					KafkaConsumerCommitIntervalMs: 5000,
					KafkaConsumerSessionTimeoutMs: 30000,
					KafkaConsumerHeartbeatMs:      1000,
					KafkaConsumerMaxPollRecords:   1000,
					KafkaConsumerAutoOffsetReset:  "earliest",
					KafkaIsolationLevel:           "read_committed",
				},
			},
			setCallback:       true,
			callbackReturns:   errors.New("callback error"),
			setupMock:         func(mock *MockConfigManagerService) {},
			expectedStatus:    fiber.StatusInternalServerError,
			expectedHasConfig: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup handler with mock
			mockSvc := NewMockConfigManagerService(ctrl)

			// Configure mock if needed
			tt.setupMock(mockSvc)

			// Create the environment
			env := &config.Env{
				ConfigServiceUrl: "http://localhost:8080",
			}

			// Create mock observability stack
			obs := &observability.ObservabilityStack{
				TracerService:  &MockTracerService{},
				MetricsService: &MockMetricsService{},
				LoggerService:  &MockLoggerService{},
			}

			// Create the handler
			h, err := handler.NewConfigHandler(mockSvc, env, obs)
			assert.NoError(t, err)

			// Create a fiber app for testing
			app := fiber.New()

			// Set callback if needed
			if tt.setCallback {
				h.SetReinitCallback(func(configuration *config.Config) error {
					return tt.callbackReturns
				})
			} else {
				h.SetReinitCallback(nil)
			}

			app.Post("/config/update", h.UpdateConfigurations)

			// Create request body
			reqBody, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("POST", "/config/update", bytes.NewReader(reqBody))
			req.Header.Set("Content-Type", "application/json")

			// Execute request
			resp, _ := app.Test(req)

			// Verify response
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
		})
	}
}

func TestHealth(t *testing.T) {
	h, _, ctrl := setupConfigHandler(t)
	defer ctrl.Finish()

	// Create a fiber app for testing
	app := fiber.New()
	app.Get("/health", h.Health)

	// Create a test request
	req := httptest.NewRequest("GET", "/health", nil)

	// Execute request
	resp, _ := app.Test(req)

	// Verify response status code
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	// Verify response body
	var result map[string]interface{}
	err := json.NewDecoder(resp.Body).Decode(&result)
	assert.NoError(t, err)
	assert.Equal(t, "OK", result["status"])
	assert.Equal(t, true, result["success"])
}
