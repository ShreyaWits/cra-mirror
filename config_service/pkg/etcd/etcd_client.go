package etcdDB

import (
	"log"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

var Client *clientv3.Client

func InitEtcdDB() {
	var err error
	Client, err = clientv3.New(clientv3.Config{
		Endpoints:   []string{"localhost:2379"},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		log.Fatalf("Failed to connect to etcd: %v", err)
	}
}
