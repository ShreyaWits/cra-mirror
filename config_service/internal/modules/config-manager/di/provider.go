package di

import (
	"nps-config-service/internal/configs"
	handler "nps-config-service/internal/modules/config-manager/apis/handlers"
	"nps-config-service/internal/modules/config-manager/repositories"
	"nps-config-service/internal/modules/config-manager/services"
	etcdDB "nps-config-service/pkg/etcd"
)

func InitHandlers(container *Container) (*handler.ConfigHandler, *handler.WebhookHandler, error) {
	configHandler := handler.NewConfigHandler(container.ConfigService)
	webhookHandler := handler.NewWebhookHandler(container.WebhookService)
	return configHandler, webhookHandler, nil
}

type Container struct {
	ConfigRepo     repositories.IConfigRepo
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

	// Repository layer
	configRepo := repositories.NewConfigRepository(*etcdClient)
	container.ConfigRepo = configRepo

	// Service layer
	webhookService := services.NewWebhookService(configRepo)
	container.WebhookService = webhookService.(*services.WebhookService)

	configService := services.NewConfigService(configRepo, webhookService)
	container.ConfigService = configService.(*services.ConfigService)

	return container, nil
}
