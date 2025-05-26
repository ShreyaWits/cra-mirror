package services

import (
	"context"
	"encoding/json"
	"fmt"
	"thirdparty_service/internal/config"
	"thirdparty_service/internal/modules/config/dto"
	cacheclient "thirdparty_service/internal/modules/execute/clients/cache_client"
	"time"

	"github.com/google/uuid"
)

type CacheManagerService struct {
	cacheclient cacheclient.RedisClient // This will be initialized later
}

// Remove the local interface definition

func NewCacheManager(cacheClient cacheclient.RedisClient) (*CacheManagerService, error) {
	if cacheClient == nil {
		return nil, fmt.Errorf("cacheClient cannot be zero in NewCacheManager")
	}

	return &CacheManagerService{
		cacheclient :cacheClient,
	}, nil
}

func (s *CacheManagerService) SetDataToCache(ctx context.Context, key string, cfg *dto.ConfigResponse) error {
	if s.cacheclient == nil {
		return fmt.Errorf("cache client not initialized")
	}
	// Add debug logs to show cache operation details
	// s.obs.LoggerService.Debug(ctx, "Setting data to cache ", "namespace ", config.AppConfig.ServiceName, "key", key)

	data, err := json.Marshal(cfg)
	if err != nil {
		// s.obs.LoggerService.Error(ctx, "Failed to marshal config ", "error", err)
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	trackingID := uuid.NewString()
	ttl := time.Duration(240) * time.Hour
	// s.obs.LoggerService.Debug(ctx, "Cache parameters ", "ttl_hours ", s.env.MESSAGING_SERVICE_REDIS_TTL, "tracking_id", trackingID)

	return s.cacheclient.SetCache(ctx, config.AppConfig.ServiceName, key, string(data), ttl, trackingID)
}

func (s *CacheManagerService) GetDataToCache(ctx context.Context, key string) (*dto.ConfigResponse, error) {
	if s.cacheclient == nil {
		return nil, fmt.Errorf("cache client not initialized")
	}
	// Add debug logs to show cache lookup details
	// s.obs.LoggerService.Debug(ctx, "Getting data from cache", "namespace", config.AppConfig.ServiceName, "key", key)

	trackingID := uuid.NewString()
	// s.obs.LoggerService.Debug(ctx, "Cache lookup with tracking_id", "tracking_id", trackingID)

	val, found, err := s.cacheclient.GetCache(ctx, config.AppConfig.ServiceName, key, trackingID)

	if err != nil {
		// s.obs.LoggerService.Error(ctx, "Failed to get cache", "error", err, "namespace", config.AppConfig.ServiceName, "key", key)
		return nil, fmt.Errorf("failed to get cache: %w", err)
	}
	if !found {
		// s.obs.LoggerService.Debug(ctx, "Config not found in cache", "namespace", config.AppConfig.ServiceName, "key", key)
		return nil, fmt.Errorf("config not found in cache")
	}

	// s.obs.LoggerService.Debug(ctx, "Successfully retrieved data from cache", "data_length", len(val))

	var cfg dto.ConfigResponse
	if err := json.Unmarshal([]byte(val), &cfg); err != nil {
		// s.obs.LoggerService.Error(ctx, "Failed to unmarshal cached config", "error", err)
		return nil, fmt.Errorf("failed to unmarshal cached config: %w", err)
	}

	return &cfg, nil
}
