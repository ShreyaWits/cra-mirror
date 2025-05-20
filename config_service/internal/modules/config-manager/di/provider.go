package di

import (
	"nps-config-service/internal/configs"
	handler "nps-config-service/internal/modules/config-manager/apis/handlers"
	"nps-config-service/internal/modules/config-manager/repositories"
	"nps-config-service/internal/modules/config-manager/services"
	etcdDB "nps-config-service/pkg/etcd"
	"nps-config-service/pkg/observability"
	workflows "nps-config-service/pkg/temporal"
)

func InitHandlers(container *Container) (*handler.AdminHandler, *handler.ConfigHandler, *handler.WebhookHandler, error) {
	adminHandler := handler.NewAdminHandler(container.AdminService, container.ObservabilityStack)
	configHandler := handler.NewConfigHandler(container.ConfigService, container.ObservabilityStack)
	webhookHandler := handler.NewWebhookHandler(container.WebhookService, configHandler.ObservabilityStack)
	return adminHandler, configHandler, webhookHandler, nil
}

type Container struct {
	ConfigRepo     repositories.IConfigRepo
	AdminService   *services.AdminService
	WebhookService *services.WebhookService
	ConfigService  *services.ConfigService
	ObservabilityStack *observability.ObservabilityStack
	
}

func NewContainer() (*Container, error) {
	container := &Container{}

	// Load config
	client, err := etcdDB.InitEtcdDB(configs.AppConfig.EtcdEndpoint)
	if err != nil {
		return nil, err
	}

	etcdClient := etcdDB.NewEtcdClientImpl(client)
	temporalUrl := configs.AppConfig.TemporalEndpoint
	temporalClient, err := workflows.InitTemporal(temporalUrl)
	if err != nil {
		return nil, err
	}
	// Repository layer
	observabilityStack := observability.NewObservabilityStack("config-service")
	container.ObservabilityStack = observabilityStack
	configRepo := repositories.NewConfigRepository(etcdClient, observabilityStack)
	container.ConfigRepo = configRepo

	// Service layer
	adminService := services.NewAdminService(configRepo, observabilityStack)
	container.AdminService = adminService.(*services.AdminService)

	webhookService := services.NewWebhookService(configRepo, observabilityStack)
	container.WebhookService = webhookService.(*services.WebhookService)

	// defer c.Close()
	configService := services.NewConfigService(configRepo, webhookService, *temporalClient, observabilityStack)
	container.ConfigService = configService.(*services.ConfigService)

	return container, nil
}
