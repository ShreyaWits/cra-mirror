
package metrics

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

// MetricsService defines the contract for metrics collection.
type MetricsService interface {
	IncrementCounter(ctx context.Context, name string, value int64, attrs map[string]string)
	RecordHistogram(ctx context.Context, name string, value float64, attrs map[string]string)
}

// MetricsServiceStruct is a concrete implementation of MetricsService.
type MetricsServiceStruct struct {
	meter        metric.Meter
	enableMetric bool
}

// NewMetricsService initializes the MetricsService with a meter.
func NewMetricsService(pkgName string, enableMetric bool) *MetricsServiceStruct {
	if enableMetric {
		meter := otel.Meter(pkgName)
		return &MetricsServiceStruct{
			meter:        meter,
			enableMetric: true,
		}
	}
	return &MetricsServiceStruct{
		enableMetric: false,
	}
}

// IncrementCounter increments a counter metric.
func (m *MetricsServiceStruct) IncrementCounter(ctx context.Context, name string, value int64, attrs map[string]string) {
	if !m.enableMetric {
		return
	}

	counter, err := m.meter.Int64Counter(name)
	if err != nil {
		return // log this in a real-world app
	}

	counter.Add(ctx, value, metric.WithAttributes(convertAttributes(attrs)...))
}

// RecordHistogram records a value in a histogram.
func (m *MetricsServiceStruct) RecordHistogram(ctx context.Context, name string, value float64, attrs map[string]string) {
	if !m.enableMetric {
		return
	}

	histogram, err := m.meter.Float64Histogram(name)
	if err != nil {
		return // log this in a real-world app
	}

	histogram.Record(ctx, value, metric.WithAttributes(convertAttributes(attrs)...))
}

// convertAttributes converts a map[string]string to []attribute.KeyValue
func convertAttributes(attrs map[string]string) []attribute.KeyValue {
	var result []attribute.KeyValue
	for k, v := range attrs {
		result = append(result, attribute.String(k, v))
	}
	return result
}
