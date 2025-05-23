package observability

import (
	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

// "messaging_service/pkg/logger"
// metrics "messaging_service/pkg/matrics"
// "messaging_service/pkg/tracer"

// ObservabilityStack provides a unified interface for tracing, metrics, and logging
type ObservabilityStack struct {
	TracerService  trace.Tracer
	MetricsService metric.Meter
	Logger         *Logger // Using our custom Logger wrapper
}

// NewObservabilityStack creates a new ObservabilityStack instance
func NewObservabilityStack(serviceName string) *ObservabilityStack {
	tracer := otel.Tracer(serviceName)
	meter := otel.Meter(serviceName)
	baseLogger := otelslog.NewLogger(serviceName)
	logger := NewLogger(baseLogger)

	// You might want to configure the logger further,
	// for example, setting a specific output or format.
	// For now, it defaults to the global logger.

	return &ObservabilityStack{
		TracerService:  tracer,
		MetricsService: meter,
		Logger:         logger,
	}
}
