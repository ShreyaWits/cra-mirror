package observability_test

import (
	"context"
	"encryption_microservice/internal/config"
	"encryption_microservice/pkg/logger"
	metrics "encryption_microservice/pkg/matrics"
	"encryption_microservice/pkg/observability"
	"encryption_microservice/pkg/tracer"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/codes"
)

// Mock environment configuration for testing
type mockEnv struct {
	ServiceName string
}

func (m *mockEnv) GetValue(key string) string {
	return ""
}

// Implement the Env interface for testing
func (m *mockEnv) GetServiceName() string {
	return m.ServiceName
}

func TestNewObservabilityStack(t *testing.T) {
	// Create a mock environment config
	mockEnv := &config.EnvConfig{
		// Set minimal required fields
		ServerPort:  "8080",
		GrpcPort:    "50051",
		Environment: "test",
	}

	// Create a new observability stack
	stack := observability.NewObservabilityStack(mockEnv)

	// Check that all components are initialized
	assert.NotNil(t, stack)
	assert.NotNil(t, stack.TracerService)
	assert.NotNil(t, stack.MetricsService)
	assert.NotNil(t, stack.LoggerService)

	// Test that components can be used without panicking
	assert.NotPanics(t, func() {
		ctx := context.Background()

		// Test TracerService
		newCtx, span := stack.TracerService.StartTracer(ctx, "test-span")
		stack.TracerService.SetAttributes(span, map[string]string{"key": "value"})
		stack.TracerService.SetStatus(span, codes.Ok, "success")
		stack.TracerService.StopSpan(span)

		// Test MetricsService
		stack.MetricsService.IncrementCounter(ctx, "test-counter", 1, map[string]string{"key": "value"})
		stack.MetricsService.RecordHistogram(ctx, "test-histogram", 42.0, map[string]string{"key": "value"})

		// Test LoggerService
		stack.LoggerService.Info(ctx, "test info message")
		stack.LoggerService.Error(ctx, "test error message")
		stack.LoggerService.Debug(ctx, "test debug message")
		stack.LoggerService.Warn(ctx, "test warn message")

		// Use the returned context
		assert.NotNil(t, newCtx)
	})
}

func TestMockObservabilityStack(t *testing.T) {
	mockStack := observability.NewMockObservabilityStack()

	// Check that components are initialized
	assert.NotNil(t, mockStack)
	assert.NotNil(t, mockStack.MockTracer)
	assert.NotNil(t, mockStack.MockMetrics)
	assert.NotNil(t, mockStack.MockLogger)

	// Get the components as interfaces
	tracerService, metricsService, loggerService := mockStack.GetObservabilityComponents()

	// Verify interfaces are implemented correctly
	assert.Implements(t, (*tracer.TracerService)(nil), tracerService)
	assert.Implements(t, (*metrics.MetricsService)(nil), metricsService)
	assert.Implements(t, (*logger.Logger)(nil), loggerService)

	// Test using the interfaces
	ctx := context.Background()

	// Tracer
	_, span := tracerService.StartTracer(ctx, "test-span")
	tracerService.SetAttributes(span, map[string]string{"key": "value"})
	tracerService.SetStatus(span, codes.Ok, "success")
	tracerService.StopSpan(span)

	// Metrics
	metricsService.IncrementCounter(ctx, "test-counter", 1, map[string]string{"key": "value"})
	metricsService.RecordHistogram(ctx, "test-histogram", 42.0, map[string]string{"key": "value"})

	// Logger
	loggerService.Info(ctx, "test info message")
	loggerService.Error(ctx, "test error message")

	// Verify the calls were recorded in the mocks
	assert.Equal(t, 1, len(mockStack.MockTracer.StartTracerCalls))
	assert.Equal(t, "test-span", mockStack.MockTracer.StartTracerCalls[0].Name)

	assert.Equal(t, 1, len(mockStack.MockTracer.AttributeCalls))
	assert.Equal(t, 1, len(mockStack.MockTracer.StatusCalls))
	assert.Equal(t, 1, len(mockStack.MockTracer.StopSpanCalls))

	assert.Equal(t, 1, len(mockStack.MockMetrics.CounterCalls))
	assert.Equal(t, "test-counter", mockStack.MockMetrics.CounterCalls[0].Name)
	assert.Equal(t, int64(1), mockStack.MockMetrics.CounterCalls[0].Value)

	assert.Equal(t, 1, len(mockStack.MockMetrics.HistogramCalls))
	assert.Equal(t, "test-histogram", mockStack.MockMetrics.HistogramCalls[0].Name)
}

// Test the formatEndpoint function with different URLs
func TestFormatEndpoint(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "no protocol",
			input:    "example.com:8080",
			expected: "example.com:8080",
		},
		{
			name:     "http protocol",
			input:    "http://example.com:8080",
			expected: "example.com:8080",
		},
		{
			name:     "https protocol",
			input:    "https://example.com:8080",
			expected: "example.com:8080",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := observability.FormatEndpoint(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
