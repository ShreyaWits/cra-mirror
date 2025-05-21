package observability

import (
	"messaging_service/internal/config"
	"messaging_service/pkg/logger"
	metrics "messaging_service/pkg/matrics"
	"messaging_service/pkg/tracer"
)

type ObservabilityStack struct {
	TracerService  tracer.TracerService
	MetricsService metrics.MetricsService
	LoggerService  logger.Logger // Logger will be initialized once and passed reference
}

func NewObservabilityStack(env *config.Env) *ObservabilityStack {
	return &ObservabilityStack{
		TracerService:  tracer.NewTracer(env.ServiceName, true),
		MetricsService: metrics.NewMetricsService(env.ServiceName, true),
		LoggerService:  logger.NewLogger(env.ServiceName, true),
	}
}
