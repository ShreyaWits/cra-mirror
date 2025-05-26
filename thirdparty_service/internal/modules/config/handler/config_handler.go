package handler

import (
	"thirdparty_service/internal/modules/config/dto"
	configservice "thirdparty_service/internal/modules/config/service"
)

type ConfigHandler struct {
	configService configservice.ConfigService
}

func NewConfigHandler(service configservice.ConfigService) *ConfigHandler { // Modified signature
	return &ConfigHandler{
		configService: service,
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
