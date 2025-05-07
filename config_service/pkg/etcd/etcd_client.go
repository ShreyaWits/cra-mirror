package etcdDB

import (
	"log"

	clientv3 "go.etcd.io/etcd/client/v3"
)

func InitEtcdDB(endpoint string) (*clientv3.Client, error) {
	client, err := clientv3.New(clientv3.Config{Endpoints: []string{endpoint}})
	if err != nil {
		log.Printf("Failed to initialize Etcd client: %v", err)
		return nil, err
	}
	log.Printf("Etcd client initialized successfully at endpoint: %v", endpoint)
	return client, nil
}
// docker exec etcd etcdctl --endpoints=localhost:2379 del "" --prefix