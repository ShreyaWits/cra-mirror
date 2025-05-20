package observability

import (
	"log/slog"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

// "messaging_service/pkg/logger"
// metrics "messaging_service/pkg/matrics"
// "messaging_service/pkg/tracer"
type ObservabilityStack struct {
	TracerService  trace.Tracer
	MetricsService metric.Meter
	Logger         *slog.Logger // Using the standard log.Logger for simplicity with otelslog
}

func NewObservabilityStack(serviceName string) *ObservabilityStack {
	tracer := otel.Tracer(serviceName)
	meter := otel.Meter(serviceName)
	logger := otelslog.NewLogger(serviceName)

	// You might want to configure the logger further,
	// for example, setting a specific output or format.
	// For now, it defaults to the global logger.

	return &ObservabilityStack{
		TracerService:  tracer,
		MetricsService: meter,
		Logger:         logger, // Access the underlying log.Logger
	}
}
