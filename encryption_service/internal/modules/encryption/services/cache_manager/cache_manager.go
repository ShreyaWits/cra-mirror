package cachemanager

import (
	"context"
	"encoding/json"
	"encryption_microservice/internal/config"
	cacheclient "encryption_microservice/internal/modules/encryption/client/cache_client"
	"encryption_microservice/pkg/observability"
	"fmt"
	"time"
)

const (
	configNamespace = "encryption-service"
)

// CacheManager manages cache operations for the encryption service
type CacheManager struct {
	cacheClient cacheclient.RedisClient
	env         *config.EnvConfig
	obs         *observability.ObservabilityStack
}

// NewCacheManager creates a new instance of CacheManager
func NewCacheManager(
	cacheClient cacheclient.RedisClient,
	env *config.EnvConfig,
	obs *observability.ObservabilityStack,
) (*CacheManager, error) {
	if cacheClient == nil {
		return nil, fmt.Errorf("cache client cannot be nil")
	}

	if env == nil {
		return nil, fmt.Errorf("environment config cannot be nil")
	}

	if obs == nil {
		return nil, fmt.Errorf("observability stack cannot be nil")
	}

	return &CacheManager{
		cacheClient: cacheClient,
		env:         env,
		obs:         obs,
	}, nil
}

// SetDataToCache serializes and stores data in the cache
func (cm *CacheManager) SetDataToCache(ctx context.Context, key string, value interface{}) error {
	if cm.cacheClient == nil {
		return fmt.Errorf("cache client is not initialized")
	}

	// Serialize the data to JSON
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	// Store in cache
	ttl := time.Duration(cm.env.CacheTTL) * time.Minute
	err = cm.cacheClient.SetCache(ctx, configNamespace, key, string(data), ttl, "")
	if err != nil {
		return fmt.Errorf("failed to set data in cache: %w", err)
	}

	cm.obs.LoggerService.Info(ctx, "Successfully cached data", map[string]interface{}{
		"namespace": configNamespace,
		"key":       key,
		"ttl":       ttl.String(),
	})

	return nil
}

// GetDataFromCache retrieves and deserializes data from the cache
func (cm *CacheManager) GetDataFromCache(ctx context.Context, key string, target interface{}) (bool, error) {
	if cm.cacheClient == nil {
		return false, fmt.Errorf("cache client is not initialized")
	}

	// Retrieve from cache
	data, found, err := cm.cacheClient.GetCache(ctx, configNamespace, key, "")
	if err != nil {
		return false, fmt.Errorf("failed to get data from cache: %w", err)
	}

	if !found {
		return false, nil
	}

	// Deserialize the data
	if err := json.Unmarshal([]byte(data), target); err != nil {
		return true, fmt.Errorf("failed to unmarshal value: %w", err)
	}

	cm.obs.LoggerService.Info(ctx, "Successfully retrieved cached data", map[string]interface{}{
		"namespace": configNamespace,
		"key":       key,
	})

	return true, nil
}
