package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Exported variables
var (
	REDIS_URL                    string
	REDIS_PASSWORD               string
	REDIS_USERNAME               string
	REDIS_DB                     string
	CACHING_SERVICE_REST_PORT    string
	CACHING_SERVICE_GRPC_PORT    string
	OTEL_COLLECTOR_GRPC_ENDPOINT string
	ENVIRONMENT                  string
)

// LoadEnv reads from .env and sets global config variables
func LoadEnv() {
	if os.Getenv("IS_DOCKER") != "true" {
        if err := godotenv.Load(); err != nil {
            log.Printf("Warning: No .env file found. Proceeding without it. Error: %v AND IS_DOCKER not true", err)
        } else {
            log.Println("Loaded .env file")
        }
    }

	REDIS_HOST := getEnv("REDIS_HOST", "")
	REDIS_PORT := getEnv("REDIS_PORT", "6379")

	REDIS_URL = fmt.Sprintf("%s:%s", REDIS_HOST, REDIS_PORT)
	REDIS_PASSWORD = getEnv("REDIS_PASSWORD", "")
	REDIS_USERNAME = getEnv("REDIS_USERNAME", "")
	REDIS_DB = getEnv("REDIS_DB", "0")
	CACHING_SERVICE_REST_PORT = getEnv("CACHING_SERVICE_REST_PORT", ":8080")
	CACHING_SERVICE_GRPC_PORT = getEnv("CACHING_SERVICE_GRPC_PORT", ":50051")
	OTEL_COLLECTOR_GRPC_ENDPOINT = getEnv("OTEL_COLLECTOR_GRPC_ENDPOINT", "")
	ENVIRONMENT = getEnv("ENVIRONMENT", "development")
}

func getEnv(key, fallback string) string {
	exists := os.Getenv(key)
	if exists == "" {
		return fallback
	}
	return exists
}
