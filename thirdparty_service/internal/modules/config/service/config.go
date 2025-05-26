package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"thirdparty_service/internal/modules/config/dto"

	"github.com/go-playground/validator/v10"
)

// ConfigService defines the interface for the configuration service.
// ConfigService defines the interface for the configuration service.
type ConfigService interface {
	LoginToConfigService() (string, error)
	FetchDynamicConfig(token string) (*dto.ConfigResponse, error)
	ValidateConfig(config dto.ConfigResponse) error
}

// ConfigServiceImpl is the concrete implementation of the ConfigService interface.
type ConfigServiceImpl struct {
	Environment           string
	ServiceName           string
	ConfigServiceURL      string
	ConfigServiceUsername string
	ConfigServicePassword string
	HttpClient            *http.Client
}

// Ensure the concrete ConfigServiceImpl struct implements the ConfigService interface
var _ ConfigService = (*ConfigServiceImpl)(nil)

func NewConfigService(Environment, ServiceName, ConfigServiceURL, ConfigServiceUsername, ConfigServicePassword string) ConfigService {
	return &ConfigServiceImpl{
		Environment:           Environment,
		ServiceName:           ServiceName,
		ConfigServiceURL:      ConfigServiceURL,
		ConfigServiceUsername: ConfigServiceUsername,
		ConfigServicePassword: ConfigServicePassword,
		HttpClient:            &http.Client{},
	}
}

func (cs *ConfigServiceImpl) LoginToConfigService() (string, error) {
	loginPayload := map[string]string{
		"username": cs.ConfigServiceUsername,
		"password": cs.ConfigServicePassword,
	}
	body, _ := json.Marshal(loginPayload)

	req, err := http.NewRequest(http.MethodPost, cs.ConfigServiceURL+"/admin/login", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("creating login request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := cs.HttpClient.Do(req)

	if err != nil {
		return "", fmt.Errorf("sending login request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return "", fmt.Errorf("login failed: status code %d", resp.StatusCode)
	}

	var result struct {
		Token   string `json:"token"`
		Success bool   `json:"success"`
		Message string `json:"message"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decoding login response: %w", err)
	}

	return result.Token, nil
}

func (cs *ConfigServiceImpl) FetchDynamicConfig(token string) (*dto.ConfigResponse, error) {
	url := fmt.Sprintf("%s/config/%s/%s", cs.ConfigServiceURL, cs.Environment, cs.ServiceName)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating config request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := cs.HttpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("sending config request: %w", err)
	}
	defer resp.Body.Close()

	log.Println("res", resp)

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, fmt.Errorf("config fetch failed: status code %d", resp.StatusCode)
	}

	var result struct {
		Token   string             `json:"token"`
		Success bool               `json:"success"`
		Message string             `json:"message"`
		Data    dto.ConfigResponse `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding config response: %w", err)
	}

	return &result.Data, nil

}

func (cs *ConfigServiceImpl) ValidateConfig(cfg dto.ConfigResponse) error {

	validator := validator.New(validator.WithRequiredStructEnabled())

	if err := validator.Struct(cfg); err != nil {
		return err
	}

	return nil
}
