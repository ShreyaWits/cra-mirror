package repositories

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"nps-config-service/internal/modules/config-manager/models"
	etcdDB "nps-config-service/pkg/etcd"

	"strconv"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

type ConfigRepository struct {
	EtcdClient etcdDB.EtcdClientImpl // ✅ Injected interface
}

func NewConfigRepository(etcdClient etcdDB.EtcdClientImpl) *ConfigRepository {
	return &ConfigRepository{EtcdClient: etcdClient}
}

func (r *ConfigRepository) StoreConfig(serviceName, environment string, configData map[string]interface{}) (interface{}, error) {

	// Store each config field as a separate key
	baseKey := fmt.Sprintf("%s/%s", serviceName, environment)

	log.Printf("Storing config with base key: %s", baseKey)

	// Store config data fields
	for key, value := range configData {
		configKey := fmt.Sprintf("%s/%s", baseKey, key)
		valueStr := fmt.Sprintf("%v", value)
		log.Printf("Storing key: %s, value: %s", configKey, valueStr)

		err := r.EtcdClient.PutKey(configKey, valueStr)
		if err != nil {
			log.Printf("Error storing key %s: %v", configKey, err)
			return nil, fmt.Errorf("failed to store config field %s: %v", key, err)
		}
	}

	// Store metadata
	now := time.Now()
	metadata := map[string]string{
		"created_at": now.Format(time.RFC3339Nano),
		"updated_at": now.Format(time.RFC3339Nano),
	}

	for key, value := range metadata {
		metadataKey := fmt.Sprintf("%s/%s", baseKey, key)
		log.Printf("Storing metadata key: %s, value: %s", metadataKey, value)

		err := r.EtcdClient.PutKey(metadataKey, value)
		if err != nil {
			log.Printf("Error storing metadata key %s: %v", metadataKey, err)
			return nil, fmt.Errorf("failed to store metadata field %s: %v", key, err)
		}
	}

	log.Printf("Successfully stored all config and metadata for %s", baseKey)
	return configData, nil
}

// GetConfig retrieves a configuration from etcd
func (r *ConfigRepository) GetConfig(serviceName, environment string) (map[string]interface{}, error) {
	// app.InitEtcdDB()
	// defer app.Client.Close()

	baseKey := fmt.Sprintf("%s/%s", environment, serviceName)
	log.Printf("Getting all config for base key: %s", baseKey)

	// Get all keys under the base key
	keys, err := r.EtcdClient.GetAllKeys(baseKey)
	if err != nil {
		log.Printf("Error getting all keys: %v", err)
		return nil, fmt.Errorf("failed to get config keys: %v", err)
	}

	// Create result map
	result := make(map[string]interface{})

	// Process each key-value pair
	for key, value := range keys {
		// Skip metadata fields
		if key == "created_at" || key == "updated_at" {
			continue
		}
		result[key] = value
		log.Printf("Retrieved key: %s, value: %v", key, value)
	}

	log.Printf("Successfully retrieved all config for %s", baseKey)
	return result, nil
}

func (r *ConfigRepository) GetAllKeys(prefix string) (map[string]string, error) {
	return r.EtcdClient.GetAllKeys(prefix)
}

// GetConfigMetadata retrieves metadata for a configuration
func (r *ConfigRepository) GetConfigMetadata(serviceName, environment string) (*models.ConfigMetadata, error) {

	key := fmt.Sprintf("/metadata/%s/%s", environment, serviceName)
	res, err := r.EtcdClient.GetKey(key)
	if err != nil {
		return nil, fmt.Errorf("failed to get metadata: %v", err)
	}

	var metadata models.ConfigMetadata
	err = json.Unmarshal([]byte(res), &metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %v", err)
	}

	return &metadata, nil
}

// GetConfigValue retrieves a specific config value from etcd
func (r *ConfigRepository) GetConfigValue(serviceName, environment, key string) (interface{}, error) {
	// app.InitEtcdDB()
	// defer app.Client.Close()

	// Construct the full key
	fullKey := fmt.Sprintf("%s/%s/%s", environment, serviceName, key)
	log.Printf("Getting config value for key: %s", fullKey)

	// Get the value
	value, err := r.EtcdClient.GetKey(fullKey)
	if err != nil {
		log.Printf("Error getting key %s: %v", fullKey, err)
		return nil, fmt.Errorf("failed to get config value: %v", err)
	}

	// Try to convert the value to appropriate type
	var result interface{}
	if value == "true" || value == "false" {
		result = value == "true"
	} else if intValue, err := strconv.Atoi(value); err == nil {
		result = intValue
	} else if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
		result = floatValue
	} else {
		result = value
	}

	log.Printf("Successfully retrieved value for key %s: %v", fullKey, result)
	return result, nil
}

// Test function to demonstrate usage
func (r *ConfigRepository) StoreAndRetrieveConfig() {
	// Example config data
	configData := map[string]interface{}{
		"db_url":      "postgres://user:pass@localhost:5432/db",
		"retry_count": 3,
		"timeout":     5000,
	}

	// Store config
	_, err := r.StoreConfig("user-service", "prod", configData)
	if err != nil {
		log.Printf("Failed to store config: %v", err)
		return
	}

	// Retrieve config
	config, err := r.GetConfig("user-service", "prod")
	if err != nil {
		log.Printf("Failed to get config: %v", err)
		return
	}

	fmt.Printf("Retrieved config: %+v\n", config)

	// Retrieve metadata
	metadata, err := r.GetConfigMetadata("user-service", "prod")
	if err != nil {
		log.Printf("Failed to get metadata: %v", err)
		return
	}

	fmt.Printf("Retrieved metadata: %+v\n", metadata)
}

// webhook service
func (r *ConfigRepository) Set(ctx context.Context, key string, data string, ttl time.Duration) error {
	// app.InitEtcdDB()
	// defer app.Client.Close()
	if ttl > 0 {
		// Create a lease
		leaseResp, err := r.EtcdClient.Client.Grant(ctx, int64(ttl.Seconds()))
		if err != nil {
			log.Printf("Failed to create lease: %v", err)
			return fmt.Errorf("lease creation failed: %w", err)
		}

		_, err = r.EtcdClient.Client.Put(ctx, key, data, clientv3.WithLease(leaseResp.ID))
		if err != nil {
			log.Printf("Error storing key with lease %s: %v", key, err)
			return fmt.Errorf("failed to store key %s: %w", key, err)
		}
	} else {
		_, err := r.EtcdClient.Client.Put(ctx, key, data)
		if err != nil {
			log.Printf("Error storing key %s: %v", key, err)
			return fmt.Errorf("failed to store key %s: %w", key, err)
		}
	}

	return nil
}

func (r *ConfigRepository) Get(ctx context.Context, key string) (string, error) {

	resp, err := r.EtcdClient.Client.Get(ctx, key)
	if err != nil {
		log.Printf("Failed to get key %s: %v", key, err)
		return "", fmt.Errorf("failed to retrieve key %s: %w", key, err)
	}

	if len(resp.Kvs) == 0 {
		return "", fmt.Errorf("key %s not found", key)
	}

	return string(resp.Kvs[0].Value), nil
}
func (r *ConfigRepository) Delete(ctx context.Context, key string) error {
	_, err := r.EtcdClient.Client.Delete(ctx, key)
	if err != nil {
		log.Printf("Failed to delete key %s: %v", key, err)
		return fmt.Errorf("failed to delete key %s: %w", key, err)
	}
	return nil
}
