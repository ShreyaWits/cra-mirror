package handler

import (
	"thirdparty_service/internal/modules/config/dto"
	"thirdparty_service/internal/modules/config/service"
)

type ConfigHandler struct {
	service service.ConfigService
}

func NewConfigHandler(service service.ConfigService) *ConfigHandler {
	return &ConfigHandler{service: service}
}

func (c *ConfigHandler) GetConfig() (*dto.ConfigResponse, error) {
	token, err := c.service.LoginToConfigService()
	if err != nil {
		return nil, err
	}

	config, err := c.service.FetchDynamicConfig(token)

	if err != nil {
		return nil, err
	}
	
	if err := c.service.ValidateConfig(*config); err != nil {
		return nil, err
	}

	return config, nil
}
