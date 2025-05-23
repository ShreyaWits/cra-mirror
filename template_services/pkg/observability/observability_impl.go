package observability

import (
	cacheclient "template-services/pkg/client/cache_client"
	"template-services/pkg/logger"
	metrics "template-services/pkg/metrics"
	"template-services/pkg/tracer"
)

type ObservabilityStack struct {
	TracerService  tracer.TracerService
	MetricsService metrics.MetricsService
	LoggerService  logger.Logger // Logger will be initialized once and passed reference
	CacheClient    cacheclient.RedisClient
}
