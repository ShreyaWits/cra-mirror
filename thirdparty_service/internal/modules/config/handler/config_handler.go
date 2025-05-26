package handler

import (
	"context"
	"thirdparty_service/internal/modules/config/dto"
	configservice "thirdparty_service/internal/modules/config/service"
	"thirdparty_service/internal/modules/execute/services/cache"
)

type ConfigHandler struct {
	configService configservice.ConfigService
	cacheservice  cache.CacheManager
}

func NewConfigHandler(service configservice.ConfigService, cacheservice cache.CacheManager) *ConfigHandler {
	return &ConfigHandler{
		configService: service,
		cacheservice:  cacheservice,
	}
}

func (c *ConfigHandler) GetDynamicConfig() (*dto.ConfigResponse, error) {
	token, err := c.configService.LoginToConfigService()
	if err != nil {
		return nil, err
	}

	config, err := c.configService.FetchDynamicConfig(token)

	if err != nil {
		return nil, err
	}

	if err := c.configService.ValidateConfig(*config); err != nil {
		return nil, err
	}

	return config, nil
}

func (c *ConfigHandler) GetCacheConfig() (*dto.ConfigResponse, error) {
	ctx := context.Background()
	return c.cacheservice.GetDataToCache(ctx, "config")
}
