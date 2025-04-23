package services

import (
	"context"
	"encoding/json"
	"fmt"
	"nps-config-service/internal/modules/config-manager/apis/dtos"
	"nps-config-service/internal/modules/config-manager/repositories"
	"time"

	"github.com/golang-jwt/jwt"
)

type ConfigService struct {
	Repo *repositories.ConfigRepository
}

func NewConfigService(repo *repositories.ConfigRepository) *ConfigService {
	return &ConfigService{Repo: repo}
}

func (s *ConfigService) AdminService(dto *dtos.AdminDto, secret string) (*dtos.ResponseAdminDto, error) {

	// Create JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"admin": true,
		"exp":   time.Now().Add(time.Hour * 24).Unix(), // Token expires in 1 day
	})

	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return nil, err
	}

	// Return the token in response DTO
	return &dtos.ResponseAdminDto{
		Token: tokenString,
	}, nil
}


func (s *ConfigService) StoreConfigService(env string, service string, req map[string]interface{}) (interface{}, error) {

	_, err := s.Repo.StoreConfig(env, service, req)

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

	// Define configUpdate with an appropriate value
	// Notify the webhook asynchronously
	for _, hook := range hooks {
		//TODO: ERROR HANDLING
		go s.NotifyWebhook(hook, req)
	}

	return map[string]interface{}{
		"status":  200,
		"message": "Config updated and webhook notification sent",
	}, nil

}

func (s *ConfigService) GetConfigService(service string, env string) (interface{}, error) {

	response, err := s.Repo.GetConfig(service, env)

	if err != nil {
		return nil, err
	}

	return response, nil
}

func (s *ConfigService) GetConfigValueService(serviceName string, env string, key string) (interface{}, error) {
	response, err := s.Repo.GetConfigValue(serviceName, env, key)

	if err != nil {
		return nil, err
	}

	return response, nil
}

func (s *ConfigService) GetConfigMetadataService(serviceName string, env string) (interface{}, error) {
	response, err := s.Repo.GetConfigMetadata(serviceName, env)

	if err != nil {
		return nil, err
	}

	return response, nil
}
