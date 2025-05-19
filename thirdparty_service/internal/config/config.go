package config

import (
	"errors"
	"fmt"
	"log"
	"os"
	"thirdparty_service/internal/modules/config/dto"
	"thirdparty_service/internal/modules/config/handler"
	"thirdparty_service/internal/modules/config/service"
	"thirdparty_service/internal/utils"

	"github.com/joho/godotenv"
)

var AppConfig = new(appConfig)

type appConfig struct {
	dynamicConfig
	StaticConfig
}

type dynamicConfig = *dto.ConfigResponse

type StaticConfig struct {
	Environment           string `validate:"required" env:"ENVIRONMENT"`
	ServiceName           string `validate:"required" env:"SERVICE_NAME"`
	ConfigServiceURL      string `validate:"required" env:"CONFIG_SERVICE_URL"`
	ConfigServiceUsername string `validate:"required" env:"CONFIG_SERVICE_USERNAME"`
	ConfigServicePassword string `validate:"required" env:"CONFIG_SERVICE_PASSWORD"`
	OtelCollectorURL      string `validate:"required" env:"OTEL_COLLECTOR_URL"`
}

// LoadEnv reads from .env and sets global config variables
func LoadEnv() error {
	err := godotenv.Load()
	if err != nil {
		log.Println(".env file not found, falling back to system env")
	}

	AppConfig.StaticConfig = StaticConfig{
		Environment:           getEnv("ENVIRONMENT", "development"),
		ServiceName:           getEnv("SERVICE_NAME", "thirdparty-service"),
		ConfigServiceURL:      getEnv("CONFIG_SERVICE_URL", "http://localhost:4001/api/v1"),
		ConfigServiceUsername: getEnv("CONFIG_SERVICE_USERNAME", "admin1"),
		ConfigServicePassword: getEnv("CONFIG_SERVICE_PASSWORD", "Test@1234"),
		OtelCollectorURL:      getEnv("OTEL_COLLECTOR_URL", "http://localhost:4317"),
	}

	validateErr := utils.Validate(AppConfig.StaticConfig)

	if validateErr != nil || len(validateErr) > 0 {
		return errors.New("invalid config")
	}
	// fetch config from config service
	configService := service.NewConfigService(AppConfig.Environment, AppConfig.ServiceName, AppConfig.ConfigServiceURL, AppConfig.ConfigServiceUsername, AppConfig.ConfigServicePassword)
	configHandler := handler.NewConfigHandler(configService)
	dConfig, err := configHandler.GetConfig()
	if err != nil {
		return err
	}
	AppConfig.dynamicConfig = dConfig

	return nil

}

func getEnv(key, fallback string) string {
	exists := os.Getenv(key)
	if exists == "" {
		return fallback
	}
	return exists
}

func (c *appConfig) SetEnv(payload *dto.ConfigResponse) string {
	c.dynamicConfig = payload
	return c.Environment
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
