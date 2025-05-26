package config

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

// Config represents the application configuration
type Config struct {
	// // Server settings
	// ServerPort string `json:"server_port" validate:"required"`
	// GrpcPort   string `json:"grpc_port" validate:"required"`

	// // User Service
	UserServiceURL string `json:"USER_SERVICE_URL" validate:"required"`

	// Vault settings
	VaultAddr  string `json:"HASHICORP_VAULT_ADDR" validate:"required"`
	VaultToken string `json:"HASHICORP_VAULT_TOKEN" validate:"required"`
	VaultPath  string `json:"HASHICORP_VAULT_PATH" validate:"required"`

	// Observability settings
	OtelCollectorGrpcEndpoint string `json:"OTEL_COLLECTOR_GRPC_ENDPOINT" validate:"required"`
}

// EnvConfig represents the environment configuration
type EnvConfig struct {
	// Server settings
	ServerPort string `env:"ENCRYPTION_SERVICE_REST_PORT" json:"server_port" validate:"required"`
	GrpcPort   string `env:"ENCRYPTION_SERVICE_GRPC_PORT" json:"grpc_port" validate:"required"`

	// Environment information
	Environment string `env:"ENVIRONMENT" json:"environment" validate:"required"`

	// Config service information
	ConfigServiceUrl   string `env:"CONFIG_SERVICE_URL" json:"config_service_url" validate:"required"`
	ConfigServiceToken string `env:"ENCRYPTION_CONFIG_SERVICE_TOKEN" json:"config_service_token" validate:"required"`

	// User Service
	// UserServiceURL string `env:"USER_SERVICE_URL" json:"USER_SERVICE_URL" validate:"required"`

	// Cache service
	CacheUrl string `env:"CACHING_SERVICE_GRPC_URL" json:"cache_url" validate:"required"`
	CacheTTL int    `env:"ENCRYPTION_SERVICE_REDIS_TTL" json:"cache_ttl" validate:"required"`

	// Vault settings (may be moved to config service)
	// VaultAddr        string `env:"HASHICORP_VAULT_ADDR" json:"HASHICORP_VAULT_ADDR" validate:"required"`
	// VaultToken       string `env:"HASHICORP_VAULT_TOKEN" json:"HASHICORP_VAULT_TOKEN" validate:"required"`
	// VaultPath        string `env:"HASHICORP_VAULT_PATH" json:"HASHICORP_VAULT_PATH" validate:"required"`
	// OtelCollectorGrpcEndpoint string `env:"OTLEL_COLLECTOR_GRPC_ENDPOINT" json:"OTEL_COLLECTOR_GRPC_ENDPOINT" validate:"required"`
}

var (
	config          *Config    = &Config{}
	envConfig       *EnvConfig = &EnvConfig{}
	validate        *validator.Validate
	SERVICE_NAME    string = "encryption_service"
	SERVICE_VERSION string = "1.0.0"
)

func init() {
	// Initialize validator
	validate = validator.New()

	// Register field name function for better error messages
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return fld.Name
		}
		return name
	})
}

func SetConfig(cfg *Config) error {
	fmt.Println("Setting config:", cfg)
	// Validate configuration before setting it
	if err := validate.Struct(cfg); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	config = cfg
	return nil
}

func GetConfig() *Config {
	return config
}

func GetMockConfig() *Config {
	cfg := &Config{
		// ServerPort:       "8082",
		// GrpcPort:         "50051",
		// UserServiceURL:   "http://host.docker.internal:8080",
		VaultAddr:                 "https://host.docker.internal:8200",
		VaultToken:                "mock-token",
		VaultPath:                 "transit",
		OtelCollectorGrpcEndpoint: "host.docker.internal:4317",
	}

	// Validate mock config to ensure it meets requirements
	if err := validate.Struct(cfg); err != nil {
		panic(fmt.Sprintf("invalid mock configuration: %v", err))
	}

	return cfg
}
