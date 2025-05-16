package service

import (
	"context"
	"encoding/json"
	"fmt"
	"messaging_service/internal/config"
	cacheclient "messaging_service/internal/modules/message_broker/client/cache_client"
	client "messaging_service/internal/modules/message_broker/client/config_client"
	"time"

	"github.com/google/uuid"
)

type ConfigManagerService struct {
	configClient client.ConfigClient
	cacheclient  cacheclient.RedisClient
	env          config.Env
}

func NewConfigManager(configClient client.ConfigClient, cacheclient cacheclient.RedisClient) *ConfigManagerService {
	return &ConfigManagerService{
		configClient: configClient,
		cacheclient:  cacheclient,
	}
}
func (s *ConfigManagerService) GetFromApiConfiguration() (*config.Config, error) {

	var data *config.Config
	var err error
	var cacheErr error
	data, err = s.configClient.FetchConfig(context.Background())
	if err != nil {
		data, cacheErr = s.GetDataToCache(context.Background(), s.env.ServiceName)
		if cacheErr != nil {
			return nil, err
		}
		return data, nil
	}
	config.SetConfig(data)
	err = s.SetDataToCache(context.Background(), s.env.ServiceName, data)
	if err != nil {
		fmt.Println("error saving data in cache service", err)
	}
	return data, nil
}

func (s *ConfigManagerService) SetDataToCache(ctx context.Context, key string, cfg *config.Config) error {
	data, err := json.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	return s.cacheclient.SetCache(ctx, s.env.ServiceName, key, string(data), time.Duration(s.env.CACHE_TTL)*time.Hour, uuid.NewString())
}

func (s *ConfigManagerService) GetDataToCache(ctx context.Context, key string) (*config.Config, error) {

	val, found, err := s.cacheclient.GetCache(ctx, s.env.ServiceName, key, uuid.NewString())

	if err != nil {
		return nil, fmt.Errorf("failed to get cache: %w", err)
	}
	if !found {
		return nil, fmt.Errorf("config not found in cache")
	}

	var cfg config.Config
	if err := json.Unmarshal([]byte(val), &cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cached config: %w", err)
	}

	return &cfg, nil
}
