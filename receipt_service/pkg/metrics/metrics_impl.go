package metrics

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric" // import SDK to access Shutdown
)

// MetricsService defines the contract for metrics collection.
type MetricsService interface {
	IncrementCounter(ctx context.Context, name string, value int64, attrs map[string]string)
	RecordHistogram(ctx context.Context, name string, value float64, attrs map[string]string)
	Shutdown(ctx context.Context) error
}

// MetricsServiceStruct is a concrete implementation of MetricsService.
type MetricsServiceStruct struct {
	meterProvider *sdkmetric.MeterProvider
	meter         metric.Meter
	enableMetric  bool
}

// NewMetricsService initializes the MetricsService with a meter.
func NewMetricsService(pkgName string, enableMetric bool) MetricsService {
	if enableMetric {
		provider := sdkmetric.NewMeterProvider()
		meter := provider.Meter(pkgName)
		return &MetricsServiceStruct{
			meter:         meter,
			meterProvider: provider,
			enableMetric:  true,
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
		return // consider logging in production
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
		return // consider logging in production
	}

	histogram.Record(ctx, value, metric.WithAttributes(convertAttributes(attrs)...))
}

// Shutdown gracefully shuts down the MeterProvider.
func (m *MetricsServiceStruct) Shutdown(ctx context.Context) error {
	if m.enableMetric && m.meterProvider != nil {
		return m.meterProvider.Shutdown(ctx)
	}
	return nil
}

// convertAttributes converts a map[string]string to []attribute.KeyValue
func convertAttributes(attrs map[string]string) []attribute.KeyValue {
	var result []attribute.KeyValue
	for k, v := range attrs {
		result = append(result, attribute.String(k, v))
	}
	return result
}
