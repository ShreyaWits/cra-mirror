package di

import (
	"fmt"
	configEnv "template-services/internal/configs"
	configs "template-services/internal/configs"
	"template-services/internal/constants"
	"template-services/internal/template/handler"
	"template-services/internal/template/repository"
	"template-services/internal/template/service"
	cacheclient "template-services/pkg/client/cache_client"
	"template-services/pkg/db"
	"template-services/pkg/logger"
	"template-services/pkg/metrics"
	"template-services/pkg/observability"
	"template-services/pkg/tracer"
)

type Container struct {
	TemplateRepo        repository.TemplateRepository
	TemplateService     *service.TemplateServiceImpl
	TemplateHandler     *handler.TemplateHandler
	TemplateGRPCHandler *handler.TemplateGRPCHandler
	YugabyteDB          db.DBModeler
	Observability       *observability.ObservabilityStack
	ConfigService       *service.ConfigServiceImpl
	RedisCache          *cacheclient.RedisClientStruct
}

var GlobalContainer *Container

func InitCacheConfig() {
	// Initialize Redis cache
	redisCache, err := cacheclient.NewRedisClient(configEnv.ImmutableConfigs.CacheSrvAddr)
	if err != nil {
		panic("failed to initialize Redis cache")
	}

	GlobalContainer = &Container{}
	GlobalContainer.RedisCache = redisCache

	obs := &observability.ObservabilityStack{
		TracerService:  tracer.NewTracer(constants.ServiceName, true),
		MetricsService: metrics.NewMetricsService(constants.ServiceName, true),
		LoggerService:  logger.NewLogger(constants.ServiceName, true),
	}
	GlobalContainer.ConfigService = service.NewConfigService(redisCache, obs)
}

func NewContainer() (*Container, error) {
	cacheService, err := cacheclient.NewRedisClient(configs.Configs.CacheUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to create Redis client: %w", err)
	}

	obs := &observability.ObservabilityStack{
		TracerService:  tracer.NewTracer(constants.ServiceName, true),
		MetricsService: metrics.NewMetricsService(constants.ServiceName, true),
		LoggerService:  logger.NewLogger(constants.ServiceName, true),
	}

	// Initialize database
	yugabyteDB := db.NewYugabyteDB(
		configs.Configs.YugabyteDBHost,
		configs.Configs.YugabyteDBPort,
		configs.Configs.YugabyteDBUser,
		configs.Configs.YugabyteDBPassword,
		configs.Configs.YugabyteDBName,
	)
	if yugabyteDB == nil {
		return nil, err
	}
	GlobalContainer.YugabyteDB = yugabyteDB

	// Initialize repository
	templateRepo := repository.NewTemplateRepository(yugabyteDB, obs)
	GlobalContainer.TemplateRepo = templateRepo

	// Initialize service
	templateService := service.NewTemplateService(templateRepo, cacheService, obs, configs.Configs)
	GlobalContainer.TemplateService = templateService

	// Initialize handlers
	templateHandler := handler.NewTemplateHandler(templateService, obs, GlobalContainer.ConfigService)
	GlobalContainer.TemplateHandler = templateHandler

	templateGRPCHandler := handler.NewTemplateGRPCHandler(templateService, obs)
	GlobalContainer.TemplateGRPCHandler = templateGRPCHandler

	return GlobalContainer, nil
}

func InitHandlers(container *Container) (*handler.TemplateHandler, *handler.TemplateGRPCHandler, error) {
	return container.TemplateHandler, container.TemplateGRPCHandler, nil
}
