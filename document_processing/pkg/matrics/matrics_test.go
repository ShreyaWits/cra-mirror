package metrics_test

import (
	"context"
	metrics "Document-Processing/pkg/matrics"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMockMetricsService(t *testing.T) {
	mockMetrics := metrics.NewMockMetricsService()
	ctx := context.Background()

	// Test IncrementCounter
	attrs := map[string]string{
		"service": "test-service",
		"event":   "test-event",
	}
	mockMetrics.IncrementCounter(ctx, "test_counter", 5, attrs)

	assert.Equal(t, 1, len(mockMetrics.CounterCalls))
	assert.Equal(t, "test_counter", mockMetrics.CounterCalls[0].Name)
	assert.Equal(t, int64(5), mockMetrics.CounterCalls[0].Value)
	assert.Equal(t, attrs, mockMetrics.CounterCalls[0].Attrs)
	assert.Equal(t, ctx, mockMetrics.CounterCalls[0].Ctx)

	// Test RecordHistogram
	mockMetrics.RecordHistogram(ctx, "test_histogram", 123.45, attrs)

	assert.Equal(t, 1, len(mockMetrics.HistogramCalls))
	assert.Equal(t, "test_histogram", mockMetrics.HistogramCalls[0].Name)
	assert.Equal(t, 123.45, mockMetrics.HistogramCalls[0].Value)
	assert.Equal(t, attrs, mockMetrics.HistogramCalls[0].Attrs)
	assert.Equal(t, ctx, mockMetrics.HistogramCalls[0].Ctx)

	// Test multiple calls
	mockMetrics.IncrementCounter(ctx, "another_counter", 10, attrs)
	assert.Equal(t, 2, len(mockMetrics.CounterCalls))

	// Verify that attributes are copied and not shared
	mockMetrics.RecordHistogram(ctx, "another_histogram", 67.89, attrs)
	assert.Equal(t, 2, len(mockMetrics.HistogramCalls))

	// Modify the original attrs
	attrs["new_key"] = "new_value"

	// The recorded calls should not be affected
	assert.NotContains(t, mockMetrics.CounterCalls[0].Attrs, "new_key")
	assert.NotContains(t, mockMetrics.HistogramCalls[0].Attrs, "new_key")
}

func TestMetricsService(t *testing.T) {
	// Test with metrics enabled
	enabledMetrics := metrics.NewMetricsService("test-package", true)
	assert.NotNil(t, enabledMetrics)

	// These calls should not panic even though we can't easily verify their behavior
	ctx := context.Background()
	attrs := map[string]string{"test": "value"}

	assert.NotPanics(t, func() {
		enabledMetrics.IncrementCounter(ctx, "test_counter", 1, attrs)
		enabledMetrics.RecordHistogram(ctx, "test_histogram", 42.0, attrs)
	})

	// Test with metrics disabled
	disabledMetrics := metrics.NewMetricsService("test-package", false)
	assert.NotNil(t, disabledMetrics)

	assert.NotPanics(t, func() {
		disabledMetrics.IncrementCounter(ctx, "test_counter", 1, attrs)
		disabledMetrics.RecordHistogram(ctx, "test_histogram", 42.0, attrs)
	})
}

func TestConvertAttributes(t *testing.T) {
	// Since convertAttributes is an internal function, we can't test it directly.
	// Instead, we'll test it indirectly by verifying the behavior of the functions that use it.

	mockMetrics := metrics.NewMockMetricsService()
	ctx := context.Background()

	// Empty attributes
	emptyAttrs := map[string]string{}
	mockMetrics.IncrementCounter(ctx, "empty_attrs", 1, emptyAttrs)
	assert.Equal(t, emptyAttrs, mockMetrics.CounterCalls[0].Attrs)

	// Nil attributes (should be handled without panicking)
	assert.NotPanics(t, func() {
		mockMetrics.IncrementCounter(ctx, "nil_attrs", 1, nil)
	})

	// Multiple attributes
	multiAttrs := map[string]string{
		"key1": "value1",
		"key2": "value2",
		"key3": "value3",
	}
	mockMetrics.IncrementCounter(ctx, "multi_attrs", 1, multiAttrs)
	assert.Equal(t, 3, len(mockMetrics.CounterCalls[2].Attrs))
	assert.Equal(t, multiAttrs, mockMetrics.CounterCalls[2].Attrs)
}
