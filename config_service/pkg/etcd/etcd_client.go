package etcdDB

import (
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

//	func InitEtcdDB() {
//		var err error
//		Client, err = clientv3.New(clientv3.Config{
//			Endpoints:   []string{"localhost:2379"},
//			DialTimeout: 5 * time.Second,
//		})
//		if err != nil {
//			log.Fatalf("Failed to connect to etcd: %v", err)
//		}
//	}

func InitEtcdDB(endpoint string) (*clientv3.Client, error) {
	return clientv3.New(clientv3.Config{
		Endpoints:   []string{endpoint},
		DialTimeout: 5 * time.Second,
	})
}
