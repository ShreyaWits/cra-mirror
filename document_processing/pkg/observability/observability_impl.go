package observability

import (
	// "Document-Processing/internal/config"
	"document_processing/internal/constants"
	"document_processing/pkg/logger"
	metrics "document_processing/pkg/matrics"
	"document_processing/pkg/tracer"
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
