package metrics

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

// MetricsServiceStruct is a concrete implementation of MetricsService.
type MetricsServiceStruct struct {
	meter        metric.Meter
	enableMetric bool
	serviceName  string
}

// NewMetricsService initializes the MetricsService with a meter.
func NewMetricsService(serviceName string, enableMetric bool) MetricsService {
	if enableMetric {
		meter := otel.Meter(serviceName)
		return &MetricsServiceStruct{
			meter:        meter,
			enableMetric: true,
			serviceName:  serviceName,
		}
	}
	return &MetricsServiceStruct{
		enableMetric: false,
		serviceName:  serviceName,
	}
}

// IncrementCounter increments a counter metric.
func (m *MetricsServiceStruct) IncrementCounter(ctx context.Context, name string, value int64, attrs map[string]string) {
	if !m.enableMetric {
		return
	}

	// Create a qualified name with service prefix
	metricName := fmt.Sprintf("%s.%s", m.serviceName, name)

	// Get or create counter
	counter, err := m.meter.Int64Counter(metricName)
	if err != nil {
		// In a production environment, this would be logged
		// We can't use the logger here to avoid circular dependencies
		fmt.Printf("Error creating counter metric %s: %v\n", metricName, err)
		return
	}

	// Add service name to attributes if not already present
	attributeMap := map[string]string{
		"service": m.serviceName,
	}

	// Copy user attributes to prevent nil map panics
	if attrs != nil {
		for k, v := range attrs {
			attributeMap[k] = v
		}
	}

	// Add the counter value
	counter.Add(ctx, value, metric.WithAttributes(convertAttributes(attributeMap)...))
}

// RecordHistogram records a value in a histogram.
func (m *MetricsServiceStruct) RecordHistogram(ctx context.Context, name string, value float64, attrs map[string]string) {
	if !m.enableMetric {
		return
	}

	// Create a qualified name with service prefix
	metricName := fmt.Sprintf("%s.%s", m.serviceName, name)

	// Get or create histogram
	histogram, err := m.meter.Float64Histogram(metricName)
	if err != nil {
		// In a production environment, this would be logged
		// We can't use the logger here to avoid circular dependencies
		fmt.Printf("Error creating histogram metric %s: %v\n", metricName, err)
		return
	}

	// Add service name to attributes if not already present
	attributeMap := map[string]string{
		"service": m.serviceName,
	}

	// Copy user attributes to prevent nil map panics
	if attrs != nil {
		for k, v := range attrs {
			attributeMap[k] = v
		}
	}

	// Record the histogram value
	histogram.Record(ctx, value, metric.WithAttributes(convertAttributes(attributeMap)...))
}

// convertAttributes converts a map[string]string to []attribute.KeyValue
func convertAttributes(attrs map[string]string) []attribute.KeyValue {
	var result []attribute.KeyValue

	// Protect against nil map
	if attrs == nil {
		return result
	}

	// Convert each attribute
	for k, v := range attrs {
		result = append(result, attribute.String(k, v))
	}
	return result
}
