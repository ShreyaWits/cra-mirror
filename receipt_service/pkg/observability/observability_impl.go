package observability

import (
	"nps-reciept-service/pkg/logger"
	"nps-reciept-service/pkg/metrics"
	"nps-reciept-service/pkg/tracer"
)

type ObservabilityStack struct {
	TracerService  tracer.TracerService
	MetricsService metrics.MetricsService
	LoggerService  logger.Logger // Logger will be initialized once and passed reference
}
