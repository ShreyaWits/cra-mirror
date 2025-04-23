package di

import (
	"nps-config-service/internal/configs"
	handler "nps-config-service/internal/modules/config-manager/apis/handlers"
	"nps-config-service/internal/modules/config-manager/repositories"
	"nps-config-service/internal/modules/config-manager/services"
	etcdDB "nps-config-service/pkg/etcd"
)

func InitHandler() (*handler.Handler, error) {

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
	service := services.NewConfigService(repo)

	return handler.NewHandler(service), nil
}
