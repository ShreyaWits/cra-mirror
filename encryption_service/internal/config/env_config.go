package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
)

// Config holds all configuration for the application

// LoadConfig loads configuration from environment variables
func LoadConfig() (*EnvConfig, error) {
	// Load .env file if it exists
	if getEnvString("IS_DOCKER", "false") == "false" {
		err := godotenv.Load()
		if err != nil {
			return nil, fmt.Errorf("failed to load .env file: %w", err)
		}
	}

	// Load environment variables into struct
	envConfig = &EnvConfig{
		ServerPort:         getEnvString("ENCRYPTION_SERVICE_REST_PORT", "8002"),
		GrpcPort:           getEnvString("ENCRYPTION_SERVICE_GRPC_PORT", "50502"),
		Environment:        getEnvString("ENVIRONMENT", "development"),
		ConfigServiceUrl:   getEnvString("CONFIG_SERVICE_URL", ""),
		ConfigServiceToken: getEnvString("ENCRYPTION_CONFIG_SERVICE_TOKEN", ""),
		// UserServiceURL:     getEnvString("USER_SERVICE_URL", "http://host.docker.internal:8080"),
		CacheUrl: getEnvString("CACHING_SERVICE_GRPC_URL", ""),
		CacheTTL: getEnvInt("ENCRYPTION_SERVICE_REDIS_TTL", 240), // Default 4 hours in minutes
		// VaultAddr:          getEnvString("HASHICORP_VAULT_ADDR", "https://host.docker.internal:8200"),
		// VaultToken:         getEnvString("HASHICORP_VAULT_TOKEN", "root"),
		// VaultPath:          getEnvString("HASHICORP_VAULT_PATH", "transit"),
		// OtelCollectorGrpcEndpoint:   getEnvString("OTLEL_COLLECTOR_GRPC_ENDPOINT", ""),
	}

	// Validate using the validator
	if err := validate.Struct(envConfig); err != nil {
		validationErrors, ok := err.(validator.ValidationErrors)
		if ok {
			// Format validation errors more nicely
			var errorMessages []string
			for _, e := range validationErrors {
				errorMessages = append(errorMessages, fmt.Sprintf("Field '%s' failed validation: %s", e.Field(), e.Tag()))
			}
			return nil, fmt.Errorf("environment validation failed: %s", strings.Join(errorMessages, "; "))
		}
		return nil, fmt.Errorf("environment validation failed: %w", err)
	}

	return envConfig, nil
}

// getEnvString gets a string environment variable or returns a default value
func getEnvString(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// getEnvInt gets an integer environment variable or returns a default value
func getEnvInt(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	var result int
	_, err := fmt.Sscanf(value, "%d", &result)
	if err != nil {
		return defaultValue
	}

	return result
}

// getEnvBool gets a boolean environment variable or returns a default value
func getEnvBool(key string, defaultValue bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	var result bool
	if strings.ToLower(value) == "true" || value == "1" {
		result = true
	} else if strings.ToLower(value) == "false" || value == "0" {
		result = false
	} else {
		return defaultValue
	}

	return result
}

// getEnvFloat gets a float environment variable or returns a default value
func getEnvFloat(key string, defaultValue float64) float64 {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	var result float64
	_, err := fmt.Sscanf(value, "%f", &result)
	if err != nil {
		return defaultValue
	}

	return result
}
