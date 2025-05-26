package observability

import (
	configEnv "template-services/internal/configs"
	"template-services/internal/constants"
	"template-services/pkg/logger"
	metrics "template-services/pkg/metrics"
	"template-services/pkg/tracer"
)

type ObservabilityStack struct {
	TracerService  tracer.TracerService
	MetricsService metrics.MetricsService
	LoggerService  logger.Logger // Logger will be initialized once and passed reference
}

func NewObservabilityStack(env *configEnv.Config) (*ObservabilityStack, error) {

	return &ObservabilityStack{
		TracerService:  tracer.NewTracer(constants.ServiceName, true),
		MetricsService: metrics.NewMetricsService(constants.ServiceName, true),
		LoggerService:  logger.NewLogger(constants.ServiceName, true),
	}, nil
}
