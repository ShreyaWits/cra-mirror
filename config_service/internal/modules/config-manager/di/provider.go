package di

import (
	"nps-config-service/internal/configs"
	handler "nps-config-service/internal/modules/config-manager/apis/handlers"
	"nps-config-service/internal/modules/config-manager/repositories"
	"nps-config-service/internal/modules/config-manager/services"
	etcdDB "nps-config-service/pkg/etcd"
	"sync"
)

type SharedDependencies struct {
	Repo           repositories.IConfigRepo
	WebhookService services.IWebhookService
}

var (
	sharedDeps     *SharedDependencies
	sharedDepsOnce sync.Once
	sharedDepsErr  error
)

func InitSharedDependencies() (*SharedDependencies, error) {
	sharedDepsOnce.Do(func() {
		client, err := etcdDB.InitEtcdDB(configs.AppConfig.EtcdEndpoint)
		if err != nil {
			sharedDepsErr = err
			return
		}

		etcdClient := etcdDB.NewEtcdClientImpl(client)
		repo := repositories.NewConfigRepository(*etcdClient)
		webhookService := services.NewWebhookService(repo)

		sharedDeps = &SharedDependencies{
			Repo:           repo,
			WebhookService: webhookService,
		}
	})

	return sharedDeps, sharedDepsErr
}
func InitConfigHandler() (*handler.ConfigHandler, error) {

	// trigger ETCD
	client, err := etcdDB.InitEtcdDB(configs.AppConfig.EtcdEndpoint)
	if err != nil {
		return nil, err
	}

	// Create etcd client
	etcdClient := etcdDB.NewEtcdClientImpl(client)

	// Inject into repository
	repo := repositories.NewConfigRepository(*etcdClient)

	// Inject into service
	serviceWebhook := services.NewWebhookService(repo)
	service := services.NewConfigService(repo, serviceWebhook)

	return handler.NewConfigHandler(service), nil
}
func InitWebhookHandler() (*handler.WebhookHandler, error) {
	// trigger ETCD
	client, err := etcdDB.InitEtcdDB(configs.AppConfig.EtcdEndpoint)
	if err != nil {
		return nil, err
	}

	// Create etcd client
	etcdClient := etcdDB.NewEtcdClientImpl(client)

	// Inject into repository
	repo := repositories.NewConfigRepository(*etcdClient)

	// Inject into service
	service := services.NewWebhookService(repo)

	return handler.NewWebhookHandler(service), nil
}
