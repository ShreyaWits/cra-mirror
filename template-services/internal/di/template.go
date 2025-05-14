package di

import (
	configs "template-services/internal/configs"
	"template-services/internal/pkg/cache"
	"template-services/internal/pkg/db"
	"template-services/internal/template/handler"
	"template-services/internal/template/repository"
	"template-services/internal/template/service"
)

type Container struct {
	TemplateRepo        repository.TemplateRepository
	TemplateService     *service.TemplateServiceImpl
	TemplateHandler     *handler.TemplateHandler
	TemplateGRPCHandler *handler.TemplateGRPCHandler
	YugabyteDB          db.DBModeler
	RedisCache          *cache.RedisCache
}

func NewContainer() (*Container, error) {
	container := &Container{}

	// Load config
	configEnv, err := configs.LoadConfig()
	if err != nil {
		return nil, err
	}

	// Initialize database
	yugabyteDB := db.NewYugabyteDB(
		configEnv.YugabyteDBHost,
		configEnv.YugabyteDBPort,
		configEnv.YugabyteDBUser,
		configEnv.YugabyteDBPassword,
		configEnv.YugabyteDBName,
	)
	if yugabyteDB == nil {
		return nil, err
	}
	container.YugabyteDB = yugabyteDB

	// Initialize Redis cache
	redisCache := cache.NewRedisCache(
		configEnv.RedisHost,
		configEnv.RedisPort,
		configEnv.RedisPassword,
	)
	container.RedisCache = redisCache

	// Initialize repository
	templateRepo := repository.NewTemplateRepository(yugabyteDB)
	container.TemplateRepo = templateRepo

	// Initialize service
	templateService := service.NewTemplateService(templateRepo, redisCache)
	container.TemplateService = templateService

	// Initialize handlers
	templateHandler := handler.NewTemplateHandler(templateService)
	container.TemplateHandler = templateHandler

	templateGRPCHandler := handler.NewTemplateGRPCHandler(templateService)
	container.TemplateGRPCHandler = templateGRPCHandler

	return container, nil
}

func InitHandlers(container *Container) (*handler.TemplateHandler, *handler.TemplateGRPCHandler, error) {
	return container.TemplateHandler, container.TemplateGRPCHandler, nil
}
