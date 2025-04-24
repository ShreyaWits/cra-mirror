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
	Repo           repositories.IConfigRepo
	WebhookService IWebhookService
}

type IConfigService interface {
	AdminService(dto *dtos.AdminDto, secret string) (*dtos.ResponseAdminDto, error)
	StoreConfigService(env string, service string, req map[string]interface{}) (interface{}, error)
	GetConfigService(service string, env string) (interface{}, error)
	GetConfigValueService(serviceName string, env string, key string) (interface{}, error)
	GetConfigMetadataService(serviceName string, env string) (interface{}, error)
}

func NewConfigService(repo repositories.IConfigRepo, webHook IWebhookService) IConfigService {
	return &ConfigService{Repo: repo, WebhookService: webHook}
}

func (s *ConfigService) AdminService(dto *dtos.AdminDto, secret string) (*dtos.ResponseAdminDto, error) {
	// Create access token (1 hour expiry)
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"admin": true,
		"exp":   time.Now().Add(time.Hour * 1).Unix(),
	})

	accessTokenString, err := accessToken.SignedString([]byte(secret))
	if err != nil {
		return nil, err
	}
	
	refreshSecret := secret // fallback to same secret if not set

	// Create refresh token (7 days expiry)
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"admin": true,
		"exp":   time.Now().Add(time.Hour * 24 * 7).Unix(),
	})

	refreshTokenString, err := refreshToken.SignedString([]byte(refreshSecret))
	if err != nil {
		return nil, err
	}

	// Return both tokens
	return &dtos.ResponseAdminDto{
		Success:      true,
		Token:        accessTokenString,
		RefreshToken: refreshTokenString,
	}, nil
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

	// Define configUpdate with an appropriate value
	// Notify the webhook asynchronously
	for _, hook := range hooks {
		//TODO: ERROR HANDLING
		go s.WebhookService.NotifyWebhook(hook, req)
	}

	return response, nil

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
