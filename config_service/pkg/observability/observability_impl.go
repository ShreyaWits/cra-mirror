package observability

import (
	"nps-config-service/pkg/observability/logger"
	"nps-config-service/pkg/observability/metrics"
	"nps-config-service/pkg/observability/tracer"

	"go.opentelemetry.io/contrib/bridges/otelslog"
)

// "messaging_service/pkg/logger"
// metrics "messaging_service/pkg/matrics"
// "messaging_service/pkg/tracer"

// ObservabilityStack provides a unified interface for tracing, metrics, and logging
type ObservabilityStack struct {
	TracerService  tracer.TracerService
	MetricsService metrics.MetricsService
	Logger         logger.Logger // Using our custom Logger wrapper
}

// NewObservabilityStack creates a new ObservabilityStack instance
func NewObservabilityStack(serviceName string) *ObservabilityStack {
	tracer := tracer.NewTracer(serviceName, true)
	meter := metrics.NewMetricsService(serviceName, true)
	baseLogger := otelslog.NewLogger(serviceName)
	logger := logger.NewLogger(baseLogger)

	// You might want to configure the logger further,
	// for example, setting a specific output or format.
	// For now, it defaults to the global logger.

	return &ObservabilityStack{
		TracerService:  tracer,
		MetricsService: meter,
		Logger:         logger,
	}
}
