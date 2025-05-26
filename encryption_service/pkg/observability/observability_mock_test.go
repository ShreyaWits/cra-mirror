package observability_test

import (
	"context"
	"encryption_microservice/pkg/logger"
	metrics "encryption_microservice/pkg/matrics"
	"encryption_microservice/pkg/observability"
	"encryption_microservice/pkg/tracer"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/codes"
)

func TestNewMockObservabilityStack(t *testing.T) {
	mockStack := observability.NewMockObservabilityStack()

	// Verify the structure is initialized correctly
	assert.NotNil(t, mockStack)
	assert.NotNil(t, mockStack.MockTracer)
	assert.NotNil(t, mockStack.MockMetrics)
	assert.NotNil(t, mockStack.MockLogger)

	// Test GetObservabilityComponents
	tracerService, metricsService, loggerService := mockStack.GetObservabilityComponents()

	// Verify that the components implement the correct interfaces
	assert.Implements(t, (*tracer.TracerService)(nil), tracerService)
	assert.Implements(t, (*metrics.MetricsService)(nil), metricsService)
	assert.Implements(t, (*logger.Logger)(nil), loggerService)

	// Cast the components back to their mock types
	_, okTracer := tracerService.(*tracer.MockTracerService)
	_, okMetrics := metricsService.(*metrics.MockMetricsService)
	_, okLogger := loggerService.(*logger.MockLogger)

	// Verify the casts were successful
	assert.True(t, okTracer)
	assert.True(t, okMetrics)
	assert.True(t, okLogger)
}

func TestMockObservabilityStackUsage(t *testing.T) {
	mockStack := observability.NewMockObservabilityStack()
	ctx := context.Background()

	// Get components
	tracerService, metricsService, loggerService := mockStack.GetObservabilityComponents()

	// Use the tracer
	_, span := tracerService.StartTracer(ctx, "test-span")
	tracerService.SetAttributes(span, map[string]string{"key1": "value1"})
	tracerService.SetStatus(span, codes.Ok, "ok")
	tracerService.RecordError(span, errors.New("test error"))
	tracerService.StopSpan(span)

	// Verify tracer calls
	assert.Equal(t, 1, len(mockStack.MockTracer.StartTracerCalls))
	assert.Equal(t, "test-span", mockStack.MockTracer.StartTracerCalls[0].Name)

	assert.Equal(t, 1, len(mockStack.MockTracer.AttributeCalls))
	assert.Equal(t, map[string]string{"key1": "value1"}, mockStack.MockTracer.AttributeCalls[0].Attrs)

	assert.Equal(t, 1, len(mockStack.MockTracer.StatusCalls))
	assert.Equal(t, codes.Ok, mockStack.MockTracer.StatusCalls[0].Code)
	assert.Equal(t, "ok", mockStack.MockTracer.StatusCalls[0].Message)

	assert.Equal(t, 1, len(mockStack.MockTracer.ErrorCalls))
	assert.Equal(t, 1, len(mockStack.MockTracer.StopSpanCalls))

	// Use the metrics service
	metricsService.IncrementCounter(ctx, "test-counter", 42, map[string]string{"key2": "value2"})
	metricsService.RecordHistogram(ctx, "test-histogram", 99.9, map[string]string{"key3": "value3"})

	// Verify metrics calls
	assert.Equal(t, 1, len(mockStack.MockMetrics.CounterCalls))
	assert.Equal(t, "test-counter", mockStack.MockMetrics.CounterCalls[0].Name)
	assert.Equal(t, int64(42), mockStack.MockMetrics.CounterCalls[0].Value)
	assert.Equal(t, map[string]string{"key2": "value2"}, mockStack.MockMetrics.CounterCalls[0].Attrs)

	assert.Equal(t, 1, len(mockStack.MockMetrics.HistogramCalls))
	assert.Equal(t, "test-histogram", mockStack.MockMetrics.HistogramCalls[0].Name)
	assert.Equal(t, 99.9, mockStack.MockMetrics.HistogramCalls[0].Value)
	assert.Equal(t, map[string]string{"key3": "value3"}, mockStack.MockMetrics.HistogramCalls[0].Attrs)

	// Use the logger
	loggerService.Info(ctx, "info message")
	loggerService.Error(ctx, "error message")
	loggerService.Debug(ctx, "debug message")
	loggerService.Warn(ctx, "warn message")

	// Verify logger calls
	assert.Equal(t, 1, len(mockStack.MockLogger.InfoMessages))
	assert.Equal(t, "info message", mockStack.MockLogger.InfoMessages[0].Args[0])

	assert.Equal(t, 1, len(mockStack.MockLogger.ErrorMessages))
	assert.Equal(t, "error message", mockStack.MockLogger.ErrorMessages[0].Args[0])

	assert.Equal(t, 1, len(mockStack.MockLogger.DebugMessages))
	assert.Equal(t, "debug message", mockStack.MockLogger.DebugMessages[0].Args[0])

	assert.Equal(t, 1, len(mockStack.MockLogger.WarnMessages))
	assert.Equal(t, "warn message", mockStack.MockLogger.WarnMessages[0].Args[0])
}
