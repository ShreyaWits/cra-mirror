package di

import (
	"nps-config-service/internal/configs"
	handler "nps-config-service/internal/modules/config-manager/apis/handlers"
	"nps-config-service/internal/modules/config-manager/repositories"
	"nps-config-service/internal/modules/config-manager/services"
	etcdDB "nps-config-service/pkg/etcd"
	workflows "nps-config-service/pkg/temporal"
)

func InitHandlers(container *Container) (*handler.AdminHandler, *handler.ConfigHandler, *handler.WebhookHandler, error) {
	adminHandler := handler.NewAdminHandler(container.AdminService)
	configHandler := handler.NewConfigHandler(container.ConfigService)
	webhookHandler := handler.NewWebhookHandler(container.WebhookService)
	return adminHandler, configHandler, webhookHandler, nil
}

type Container struct {
	ConfigRepo     repositories.IConfigRepo
	AdminService   *services.AdminService
	WebhookService *services.WebhookService
	ConfigService  *services.ConfigService
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
	configRepo := repositories.NewConfigRepository(etcdClient)
	container.ConfigRepo = configRepo

	// Service layer
	adminService := services.NewAdminService(configRepo)
	container.AdminService = adminService.(*services.AdminService)

	webhookService := services.NewWebhookService(configRepo)
	container.WebhookService = webhookService.(*services.WebhookService)

	// defer c.Close()
	configService := services.NewConfigService(configRepo, webhookService, *temporalClient)
	container.ConfigService = configService.(*services.ConfigService)

	return container, nil
}
