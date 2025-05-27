package observability_test

import (
	"context"
	"document_processing/pkg/logger"
	metrics "document_processing/pkg/metrics"
	"document_processing/pkg/observability"
	"document_processing/pkg/tracer"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/codes"
)

// MockEnv implements a minimal environment for testing
type MockEnv struct {
	ServiceNameValue string
}

func (m *MockEnv) GetServiceName() string {
	return m.ServiceNameValue
}

func (m *MockEnv) GetValue(key string) string {
	return ""
}

// TestNewObservabilityStackWithMock tests creating a new observability stack with mocks
func TestNewObservabilityStackWithMock(t *testing.T) {
	// Create a mock stack
	mockStack := observability.NewMockObservabilityStack()

	// Get the stack components
	tracerService, metricsService, loggerService := mockStack.GetObservabilityComponents()

	// Test that they implement the correct interfaces
	assert.Implements(t, (*tracer.TracerService)(nil), tracerService)
	assert.Implements(t, (*metrics.MetricsService)(nil), metricsService)
	assert.Implements(t, (*logger.Logger)(nil), loggerService)
}

// TestCreateObservabilityStack tests creating a new observability stack with real implementations
func TestCreateObservabilityStack(t *testing.T) {
	// Create a direct mock environment instead of trying to use config.Env
	mockEnv := &MockEnv{
		ServiceNameValue: "test-service",
	}

	// Create a new observability stack directly
	stack := &observability.ObservabilityStack{
		TracerService:  tracer.NewTracer(mockEnv.GetServiceName(), true),
		MetricsService: metrics.NewMetricsService(mockEnv.GetServiceName(), true),
		LoggerService:  logger.NewLogger(mockEnv.GetServiceName(), true),
	}

	// Check that all components are initialized
	assert.NotNil(t, stack)
	assert.NotNil(t, stack.TracerService)
	assert.NotNil(t, stack.MetricsService)
	assert.NotNil(t, stack.LoggerService)

	// Test that the components can be used without panicking
	ctx := context.Background()

	assert.NotPanics(t, func() {
		// Test logger
		stack.LoggerService.Info(ctx, "test info message")
		stack.LoggerService.Error(ctx, "test error message")
		stack.LoggerService.Debug(ctx, "test debug message")
		stack.LoggerService.Warn(ctx, "test warn message")

		loggerWithFields := stack.LoggerService.WithFields(map[string]interface{}{
			"key": "value",
		})
		loggerWithFields.Info(ctx, "test with fields")

		// Test metrics
		stack.MetricsService.IncrementCounter(ctx, "test_counter", 1, map[string]string{
			"key": "value",
		})
		stack.MetricsService.RecordHistogram(ctx, "test_histogram", 42.0, map[string]string{
			"key": "value",
		})

		// Test tracer
		_, span := stack.TracerService.StartTracer(ctx, "test-span")
		stack.TracerService.SetAttributes(span, map[string]string{"key": "value"})
		stack.TracerService.SetStatus(span, codes.Ok, "success")
		stack.TracerService.RecordError(span, nil)
		stack.TracerService.StopSpan(span)

		// Test synchronization
		stack.LoggerService.Sync()
	})
}

// TestObservabilityStackWithDisabledOTel tests the stack with OpenTelemetry disabled
func TestObservabilityStackWithDisabledOTel(t *testing.T) {
	// Create a direct mock environment
	mockEnv := &MockEnv{
		ServiceNameValue: "test-service-disabled-otel",
	}

	// Create observability stack with disabled OTel
	stack := &observability.ObservabilityStack{
		TracerService:  tracer.NewTracer(mockEnv.GetServiceName(), false),
		MetricsService: metrics.NewMetricsService(mockEnv.GetServiceName(), false),
		LoggerService:  logger.NewLogger(mockEnv.GetServiceName(), false),
	}

	// Check that all components are initialized
	assert.NotNil(t, stack)
	assert.NotNil(t, stack.TracerService)
	assert.NotNil(t, stack.MetricsService)
	assert.NotNil(t, stack.LoggerService)

	// Test that the components can be used without panicking
	ctx := context.Background()

	assert.NotPanics(t, func() {
		// Test logger
		stack.LoggerService.Info(ctx, "test info message with OTel disabled")
		stack.LoggerService.Error(ctx, "test error message with OTel disabled")

		// Test metrics
		stack.MetricsService.IncrementCounter(ctx, "test_counter", 1, nil)

		// Test tracer
		_, span := stack.TracerService.StartTracer(ctx, "test-span")
		stack.TracerService.StopSpan(span)
	})
}
