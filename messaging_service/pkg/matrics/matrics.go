package metrics

import "context"

type Metrics interface {
	// Tracer returns the tracer service for the audit service.
	IncrementCounter(ctx context.Context, name string, value int64, attrs map[string]string)
	RecordHistogram(ctx context.Context, name string, value float64, attrs map[string]string)
}