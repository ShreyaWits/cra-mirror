package etcdDB

import (
	"context"
	"fmt"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

type EtcdClientImpl struct {
	Client *clientv3.Client
}

func NewEtcdClientImpl(client *clientv3.Client) *EtcdClientImpl {
	return &EtcdClientImpl{Client: client}
}

// PutKey stores a value with a key
func (e *EtcdClientImpl) PutKey(key, value string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := e.Client.Put(ctx, key, value)
	return err
}

// GetKey retrieves the value for a given key
func (e *EtcdClientImpl) GetKey(key string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	resp, err := e.Client.Get(ctx, key)
	if err != nil {
		return "", err
	}

	if len(resp.Kvs) > 0 {
		return string(resp.Kvs[0].Value), nil
	}
	return "", fmt.Errorf("key not found")
}

// DeleteKey removes a key
func (e *EtcdClientImpl) DeleteKey(key string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := e.Client.Delete(ctx, key)
	return err
}

// WatchKey listens for changes
func (e *EtcdClientImpl) WatchKey(key string) {
	rch := e.Client.Watch(context.Background(), key)
	for wresp := range rch {
		for _, ev := range wresp.Events {
			fmt.Printf("🔄 %s %q: %q\n", ev.Type, ev.Kv.Key, ev.Kv.Value)
		}
	}
}

// GetAllKeys gets all keys under a prefix
func (e *EtcdClientImpl) GetAllKeys(prefix string) (map[string]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := e.Client.Get(ctx, prefix, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}

	result := make(map[string]string)
	for _, kv := range resp.Kvs {
		// Extract the key without the prefix
		key := string(kv.Key[len(prefix)+1:])
		result[key] = string(kv.Value)
	}

	return result, nil
}
