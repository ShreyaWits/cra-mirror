package tracer_test

import (
	"context"
	"errors"
	"testing"

	"Document-Processing/pkg/tracer"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/codes"
)

func TestMockTracerService(t *testing.T) {
	mockTracer := tracer.NewMockTracerService()
	ctx := context.Background()

	// Test StartTracer
	newCtx, span := mockTracer.StartTracer(ctx, "test-span")
	assert.Equal(t, ctx, newCtx) // Mock just returns the original context
	assert.Equal(t, 1, len(mockTracer.StartTracerCalls))
	assert.Equal(t, "test-span", mockTracer.StartTracerCalls[0].Name)

	// Test StopSpan
	mockTracer.StopSpan(span)
	assert.Equal(t, 1, len(mockTracer.StopSpanCalls))
	assert.Equal(t, span, mockTracer.StopSpanCalls[0].Span)

	// Test SetAttributes
	attrs := map[string]string{
		"key1": "value1",
		"key2": "value2",
	}
	mockTracer.SetAttributes(span, attrs)
	assert.Equal(t, 1, len(mockTracer.AttributeCalls))
	assert.Equal(t, span, mockTracer.AttributeCalls[0].Span)
	assert.Equal(t, attrs, mockTracer.AttributeCalls[0].Attrs)

	// Test SetStatus
	mockTracer.SetStatus(span, codes.Error, "test error message")
	assert.Equal(t, 1, len(mockTracer.StatusCalls))
	assert.Equal(t, span, mockTracer.StatusCalls[0].Span)
	assert.Equal(t, codes.Error, mockTracer.StatusCalls[0].Code)
	assert.Equal(t, "test error message", mockTracer.StatusCalls[0].Message)

	// Test RecordError
	testErr := errors.New("test error")
	mockTracer.RecordError(span, testErr)
	assert.Equal(t, 1, len(mockTracer.ErrorCalls))
	assert.Equal(t, span, mockTracer.ErrorCalls[0].Span)
	assert.Equal(t, testErr, mockTracer.ErrorCalls[0].Err)
}

func TestNewTracer(t *testing.T) {
	// Test with tracing enabled
	enabledTracer := tracer.NewTracer("test-package", true)
	assert.NotNil(t, enabledTracer)

	// Check that the tracer is functional with tracing enabled
	ctx := context.Background()
	_, span := enabledTracer.StartTracer(ctx, "test-span")
	assert.NotPanics(t, func() {
		enabledTracer.SetAttributes(span, map[string]string{"key": "value"})
		enabledTracer.SetStatus(span, codes.Ok, "success")
		enabledTracer.RecordError(span, errors.New("test error"))
		enabledTracer.StopSpan(span)
	})

	// Test with tracing disabled
	disabledTracer := tracer.NewTracer("test-package", false)
	assert.NotNil(t, disabledTracer)

	// Check that the tracer is functional with tracing disabled
	ctx2 := context.Background()
	_, span2 := disabledTracer.StartTracer(ctx2, "test-span-disabled")
	assert.NotPanics(t, func() {
		disabledTracer.SetAttributes(span2, map[string]string{"key": "value"})
		disabledTracer.SetStatus(span2, codes.Ok, "success")
		disabledTracer.RecordError(span2, errors.New("test error"))
		disabledTracer.StopSpan(span2)
	})
}

// Test handling of nil values
func TestTracerWithNilValues(t *testing.T) {
	tracer := tracer.NewTracer("test-package", true)
	assert.NotNil(t, tracer)

	// These should not panic with nil values
	assert.NotPanics(t, func() {
		tracer.StopSpan(nil)
		tracer.SetAttributes(nil, map[string]string{"key": "value"})
		tracer.SetStatus(nil, codes.Ok, "success")
		tracer.RecordError(nil, errors.New("test error"))
		tracer.RecordError(nil, nil) // Both nil
	})
}
