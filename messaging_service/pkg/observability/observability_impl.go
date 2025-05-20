package observability

import (
	"messaging_service/pkg/logger"
	metrics "messaging_service/pkg/matrics"
	"messaging_service/pkg/tracer"
)

type ObservabilityStack struct {
	TracerService  tracer.TracerService
	MetricsService metrics.MetricsService
	LoggerService  logger.Logger // Logger will be initialized once and passed reference
}
