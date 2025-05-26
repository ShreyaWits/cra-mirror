package metrics

import "context"

// MetricsService defines the contract for metrics collection
type MetricsService interface {
	// IncrementCounter increments a counter metric with the given name, value, and attributes
	IncrementCounter(ctx context.Context, name string, value int64, attrs map[string]string)

	// RecordHistogram records a value in a histogram with the given name and attributes
	RecordHistogram(ctx context.Context, name string, value float64, attrs map[string]string)
}
