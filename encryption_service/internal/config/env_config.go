package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application
type Config struct {
	// Server settings
	ServerPort string

	// User Service
	UserServiceURL string

	// Vault settings
	VaultAddr  string
	VaultToken string
	VaultPath  string
	GrpcPort   string
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	// Load .env file if it exists
	godotenv.Load()

	// Default values
	config := &Config{
		ServerPort: getEnv("REST_PORT", "8082"),

		UserServiceURL: getEnv("USER_SERVICE_URL", "http://localhost:8080"),

		// Load Vault configuration
		VaultAddr:  getEnv("HASHICORP_VAULT_ADDR", "https://localhost:8200"),
		VaultToken: getEnv("HASHICORP_VAULT_TOKEN", "root"),
		VaultPath:  getEnv("HASHICORP_VAULT_PATH", "transit"),
		GrpcPort:   getEnv("GRPC_PORT", "50051"),
	}

	// Validate required Vault configuration
	if config.VaultAddr == "" {
		return nil, fmt.Errorf("HASHICORP_VAULT_ADDR environment variable is required")
	}
	if config.VaultToken == "" {
		return nil, fmt.Errorf("HASHICORP_VAULT_TOKEN environment variable is required")
	}
	if config.VaultPath == "" {
		return nil, fmt.Errorf("HASHICORP_VAULT_PATH environment variable is required")
	}
	if config.GrpcPort == "" {
		return nil, fmt.Errorf("GRPC_PORT environment variable is required")
	}

	return config, nil
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
