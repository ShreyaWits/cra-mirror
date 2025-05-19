package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
)

var AppConfig = new(appConfig)

type appConfig struct {
	dynamicConfig
	staticConfig
}

type dynamicConfig struct {
	HTTPListenAddress string `validate:"required" json:"HTTP_LISTEN_ADDRESS"`
	HTTPListenPort    int    `validate:"required,min=1,max=65535" json:"HTTP_LISTEN_PORT"`

	GRPCListenAddress string `validate:"required" json:"GRPC_LISTEN_ADDRESS"`
	GRPCListenPort    int    `validate:"required,min=1,max=65535" json:"GRPC_LISTEN_PORT"`

	DatabaseHost     string `validate:"required" json:"DATABASE_HOST"`
	DatabasePort     int    `validate:"required,min=1,max=65535" json:"DATABASE_PORT"`
	DatabaseUser     string `validate:"required" json:"DATABASE_USER"`
	DatabasePassword string `validate:"required" json:"DATABASE_PASSWORD"`
	DatabaseName     string `validate:"required" json:"DATABASE_NAME"`

	RedisHost string `validate:"required" json:"REDIS_HOST"`
	RedisPort int    `validate:"required,min=1,max=65535" json:"REDIS_PORT"`

	TwilioAccountSID string `validate:"required" json:"TWILIO_ACCOUNT_SID"`
	TwilioAuthToken  string `validate:"required" json:"TWILIO_AUTH_TOKEN"`
	TwilioFormNumber string `validate:"required" json:"TWILIO_FORM_NUMBER"`

	SendGridApiKey    string `validate:"required" json:"SENDGRID_API_KEY"`
	SendGridFromEmail string `validate:"required" json:"SENDGRID_FROM_EMAIL"`
	SendGridFromName  string `validate:"required" json:"SENDGRID_FROM_NAME"`

	SendWhatsAppMessageSID 	string `validate:"required" json:"WHATSAPP_TWILIO_ACCOUNT_SID"`
	SendWhatsAppMessageToken  string `validate:"required" json:"WHATSAPP_TWILIO_AUTH_TOKEN"`
	SendWhatsAppMessageFromNumber  string `validate:"required" json:"WHATSAPP_FROM_NUMBER"`
}

type staticConfig struct {
	Environment           string `validate:"required" env:"ENVIRONMENT"`
	ServiceName           string `validate:"required" env:"SERVICE_NAME"`
	ConfigServiceURL      string `validate:"required" env:"CONFIG_SERVICE_URL"`
	ConfigServiceUsername string `validate:"required" env:"CONFIG_SERVICE_USERNAME"`
	ConfigServicePassword string `validate:"required" env:"CONFIG_SERVICE_PASSWORD"`
	OTEL_COLLECTOR_URL    string `validate:"required" env:"OTEL_COLLECTOR_URL"`
	SERVICE_NAME          string `validate:"required" env:"SERVICE_NAME"`
	DEPLOYMENT_ENV        string `validate:"required" env:"DEPLOYMENT_ENV"`
}

// Initialization
func init() {
	if err := loadStaticConfig(); err != nil {
		log.Fatalf("error loading env variables: %v", err)
	}

	if _, err := FetchDynamicConfig(); err != nil {
		log.Fatalf("failed to load application config from server: %v\n\n", err)
	}
}

// Load static config from env/.env
func loadStaticConfig() error {
	if os.Getenv("IS_DOCKER") != "true" {
		if err := godotenv.Load(); err != nil {
			return fmt.Errorf("loading .env file: %w", err)
		}
	}

	AppConfig.staticConfig = staticConfig{
		Environment:           os.Getenv("ENVIRONMENT"),
		ServiceName:           os.Getenv("SERVICE_NAME"),
		ConfigServiceURL:      os.Getenv("CONFIG_SERVICE_URL"),
		ConfigServiceUsername: os.Getenv("CONFIG_SERVICE_USERNAME"),
		ConfigServicePassword: os.Getenv("CONFIG_SERVICE_PASSWORD"),
	}

	return validateStruct(AppConfig.staticConfig)
}

// Fetch and apply dynamic config from config server
func FetchDynamicConfig() (dynamicConfig, error) {
	token, err := loginToConfigService()
	if err != nil {
		return dynamicConfig{}, err
	}

	configData, err := retrieveConfigData(token)
	if err != nil {
		return dynamicConfig{}, err
	}

	AppConfig.dynamicConfig = configData
	return configData, nil
}

// Authenticate to config server and get token
func loginToConfigService() (string, error) {
	loginPayload := map[string]string{
		"username": AppConfig.ConfigServiceUsername,
		"password": AppConfig.ConfigServicePassword,
	}
	body, _ := json.Marshal(loginPayload)

	req, err := http.NewRequest(http.MethodPost, AppConfig.ConfigServiceURL+"/admin/login", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("creating login request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient().Do(req)
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

// Retrieve dynamic config using auth token
func retrieveConfigData(token string) (dynamicConfig, error) {
	url := fmt.Sprintf("%s/config/%s/%s", AppConfig.ConfigServiceURL, AppConfig.Environment, AppConfig.ServiceName)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return dynamicConfig{}, fmt.Errorf("creating config request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := httpClient().Do(req)
	if err != nil {
		return dynamicConfig{}, fmt.Errorf("sending config request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return dynamicConfig{}, fmt.Errorf("config fetch failed: status code %d", resp.StatusCode)
	}

	var result struct {
		Data dynamicConfig `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return dynamicConfig{}, fmt.Errorf("decoding config response: %w", err)
	}

	if err := validateStruct(result.Data); err != nil {
		return dynamicConfig{}, err
	}

	return result.Data, nil
}

// Shared HTTP client with timeout
func httpClient() *http.Client {
	return &http.Client{Timeout: 5 * time.Second}
}

// Reusable validator
func validateStruct(s interface{}) error {
	validate := validator.New()
	if err := validate.Struct(s); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}
	return nil
}

func (c *appConfig) GetDSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		c.DatabaseHost, c.DatabasePort, c.DatabaseUser, c.DatabasePassword, c.DatabaseName)
}

func (c *appConfig) GetHTTPListenAddress() string {
	return fmt.Sprintf("%s:%d", c.HTTPListenAddress, c.HTTPListenPort)
}

func (c *appConfig) GetGRPCListenAddress() string {
	return fmt.Sprintf("%s:%d", c.GRPCListenAddress, c.GRPCListenPort)
}

func (c *appConfig) GetRedisAddress() string {
	return fmt.Sprintf("%s:%d", c.RedisHost, c.RedisPort)
}
