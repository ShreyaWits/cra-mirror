package services

import (
	"context"
	"encoding/json"
	"fmt"
	"nps-config-service/internal/modules/config-manager/apis/dtos"
	"nps-config-service/internal/modules/config-manager/repositories"
)

type ConfigService struct {
	Repo           repositories.IConfigRepo
	WebhookService IWebhookService
}

type IConfigService interface {
	StoreConfigService(env string, service string, req map[string]interface{}) (interface{}, error)
	GetConfigService(service string, env string) (interface{}, error)
	GetConfigValueService(serviceName string, env string, key string) (interface{}, error)
	GetConfigMetadataService(serviceName string, env string) (interface{}, error)
}

func NewConfigService(repo repositories.IConfigRepo, webHook IWebhookService) IConfigService {
	return &ConfigService{Repo: repo, WebhookService: webHook}
}

func (s *ConfigService) StoreConfigService(env string, service string, req map[string]interface{}) (interface{}, error) {

	response, err := s.Repo.StoreConfig(env, service, req)

	if err != nil {
		fmt.Printf("failed to store %s: %v", env, err)
		return nil, fmt.Errorf("no webhook registered for %s", env)
	}
	key := fmt.Sprintf("/webhooks/%s/%s", env, service)
	ctx := context.Background()
	webHook, err := s.Repo.Get(ctx, key)
	if err != nil {
		fmt.Printf("No webhook found for %s: %v", key, err)
		return nil, fmt.Errorf("no webhook registered for %s", key)
	}

	var hooks []dtos.RegisterWebhookRequest
	if err := json.Unmarshal([]byte(webHook), &hooks); err != nil {
		fmt.Printf("Failed to parse webhook data for %s: %v", key, err)
		return nil, fmt.Errorf("invalid webhook data stored for %s", key)
	}

	for _, hook := range hooks {
		//TODO: ERROR HANDLING
		go s.WebhookService.NotifyWebhook(hook, req)
	}

	return response, nil

}

func (s *ConfigService) GetConfigService(service string, env string) (any, error) {

	response, err := s.Repo.GetConfig(service, env)

	if err != nil {
		return nil, err
	}
	if response == nil {
		return nil, nil // Explicitly return nil if response is nil
	}

	return response, nil
}

func (s *ConfigService) GetConfigValueService(serviceName string, env string, key string) (any, error) {
	response, err := s.Repo.GetConfigValue(serviceName, env, key)

	if err != nil {
		return nil, err
	}
	if response == nil {
		return nil, nil // Explicitly return nil if response is nil
	}

	return response, nil
}

func (s *ConfigService) GetConfigMetadataService(serviceName string, env string) (any, error) {
	response, err := s.Repo.GetConfigMetadata(serviceName, env)

	if err != nil {
		return nil, err
	}

	if response == nil {
		return nil, nil // Explicitly return nil if response is nil
	}
	return response, nil
}
