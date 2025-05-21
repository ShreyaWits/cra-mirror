package service

import (
	"context"
	"encoding/json"
	"fmt"
	"messaging_service/internal/config"
	cacheclient "messaging_service/internal/modules/message_broker/client/cache_client"
	client "messaging_service/internal/modules/message_broker/client/config_client"
	"messaging_service/pkg/observability"
	"time"

	"github.com/google/uuid"
)

type ConfigManagerService struct {
	configClient client.ConfigClient
	cacheclient  cacheclient.RedisClient
	env          *config.Env
	obs          *observability.ObservabilityStack
}

func NewConfigManager(configClient client.ConfigClient, cacheclient cacheclient.RedisClient, obs *observability.ObservabilityStack, env *config.Env) (*ConfigManagerService, error) {
	// Validate input arguments
	if configClient == nil {
		return nil, fmt.Errorf("configClient cannot be nil in NewConfigManager")
	}
	if cacheclient == nil {
		return nil, fmt.Errorf("cacheClient cannot be nil in NewConfigManager")
	}
	if obs == nil {
		return nil, fmt.Errorf("observability stack cannot be nil in NewConfigManager")
	}

	return &ConfigManagerService{
		configClient: configClient,
		cacheclient:  cacheclient,
		obs:          obs,
		env:          env,
	}, nil
}

// SetEnvironment sets the environment configuration
func (s *ConfigManagerService) SetEnvironment(env *config.Env) {
	s.env = env
}

func (s *ConfigManagerService) GetFromApiConfiguration(ctx context.Context) (*config.Config, error) {
	var data *config.Config
	var err error
	var cacheErr error

	// Try to fetch config from API
	s.obs.LoggerService.Debug(ctx, "Attempting to fetch configuration from API")
	data, err = s.configClient.FetchConfig(ctx)
	if err != nil {
		s.obs.LoggerService.Error(ctx, "API fetch failed, falling back to cache", "error", err)
		// If API fetch fails, try to get from cache
		data, cacheErr = s.GetDataToCache(ctx, "config")
		if cacheErr != nil {
			s.obs.LoggerService.Error(ctx, "Failed to retrieve from cache after API failure", "error", cacheErr)
			return nil, cacheErr
		}
		s.obs.LoggerService.Debug(ctx, "Successfully retrieved config from cache after API failure")
	} else {
		s.obs.LoggerService.Debug(ctx, "Successfully fetched config from API")
	}

	// Validate and set the global config
	if err := config.SetConfig(data); err != nil {
		s.obs.LoggerService.Error(ctx, "Error setting config", "error", err)
		return data, nil
	}
	s.obs.LoggerService.Debug(ctx, "Successfully set global config")

	// Store in cache
	s.obs.LoggerService.Debug(ctx, "Storing config in cache", "key", config.SERVICE_NAME)
	err = s.SetDataToCache(ctx, "config", data)
	if err != nil {
		s.obs.LoggerService.Error(ctx, "Error saving data in cache service ", "key", config.SERVICE_NAME, "error", err)
	} else {
		s.obs.LoggerService.Debug(ctx, "Successfully stored config in cache", "key", config.SERVICE_NAME)
	}

	return data, nil
}

func (s *ConfigManagerService) SetDataToCache(ctx context.Context, key string, cfg *config.Config) error {
	// Add debug logs to show cache operation details
	s.obs.LoggerService.Debug(ctx, "Setting data to cache ", "namespace ", config.SERVICE_NAME, "key", key)

	data, err := json.Marshal(cfg)
	if err != nil {
		s.obs.LoggerService.Error(ctx, "Failed to marshal config ", "error", err)
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	trackingID := uuid.NewString()
	ttl := time.Duration(s.env.MESSAGING_SERVICE_REDIS_TTL) * time.Hour
	s.obs.LoggerService.Debug(ctx, "Cache parameters ", "ttl_hours ", s.env.MESSAGING_SERVICE_REDIS_TTL, "tracking_id", trackingID)

	return s.cacheclient.SetCache(ctx, config.SERVICE_NAME, key, string(data), ttl, trackingID)
}

func (s *ConfigManagerService) GetDataToCache(ctx context.Context, key string) (*config.Config, error) {
	// Add debug logs to show cache lookup details
	s.obs.LoggerService.Debug(ctx, "Getting data from cache", "namespace", config.SERVICE_NAME, "key", key)

	trackingID := uuid.NewString()
	s.obs.LoggerService.Debug(ctx, "Cache lookup with tracking_id", "tracking_id", trackingID)

	val, found, err := s.cacheclient.GetCache(ctx, config.SERVICE_NAME, key, trackingID)

	if err != nil {
		s.obs.LoggerService.Error(ctx, "Failed to get cache", "error", err, "namespace", config.SERVICE_NAME, "key", key)
		return nil, fmt.Errorf("failed to get cache: %w", err)
	}
	if !found {
		s.obs.LoggerService.Debug(ctx, "Config not found in cache", "namespace", config.SERVICE_NAME, "key", key)
		return nil, fmt.Errorf("config not found in cache")
	}

	s.obs.LoggerService.Debug(ctx, "Successfully retrieved data from cache", "data_length", len(val))

	var cfg config.Config
	if err := json.Unmarshal([]byte(val), &cfg); err != nil {
		s.obs.LoggerService.Error(ctx, "Failed to unmarshal cached config", "error", err)
		return nil, fmt.Errorf("failed to unmarshal cached config: %w", err)
	}

	return &cfg, nil
}
