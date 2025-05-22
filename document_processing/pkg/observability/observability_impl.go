package observability

import (
	// "Document-Processing/internal/config"
	"Document-Processing/internal/constants"
	"Document-Processing/pkg/logger"
	metrics "Document-Processing/pkg/matrics"
	"Document-Processing/pkg/tracer"
)

type ObservabilityStack struct {
	TracerService  tracer.TracerService
	MetricsService metrics.MetricsService
	LoggerService  logger.Logger // Logger will be initialized once and passed reference
}

func NewObservabilityStack() *ObservabilityStack {
	return &ObservabilityStack{
		TracerService:  tracer.NewTracer(constants.SERVICE_NAME, true),
		MetricsService: metrics.NewMetricsService(constants.SERVICE_NAME, true),
		LoggerService:  logger.NewLogger(constants.SERVICE_NAME, true),
	}
}
