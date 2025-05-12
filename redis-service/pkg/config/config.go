package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Exported variables
var (
	REDIS_URL          string
	REDIS_PASSWORD     string
	REDIS_DB           string
	PORT               string
	GRPC_PORT          string
	OTEL_COLLECTOR_URL string
	SERVICE_NAME       string
	DEPLOYMENT_ENV     string
)

// LoadEnv reads from .env and sets global config variables
func LoadEnv() {
	err := godotenv.Load()
	if err != nil {
		log.Println(".env file not found, falling back to system env")
	}

	REDIS_URL = getEnv("REDIS_URL", "localhost:6379")
	REDIS_PASSWORD = getEnv("REDIS_PASSWORD", "")
	REDIS_DB = getEnv("REDIS_DB", "0")
	PORT = getEnv("PORT", ":8080")
	GRPC_PORT = getEnv("GRPC_PORT", ":50051")
	OTEL_COLLECTOR_URL = getEnv("OTEL_COLLECTOR_URL", "")
	SERVICE_NAME = getEnv("SERVICE_NAME", "config-service")
	DEPLOYMENT_ENV = getEnv("DEPLOYMENT_ENV", "development")
}

func getEnv(key, fallback string) string {
	exists := os.Getenv(key)
	if exists == "" {
		return fallback
	}
	return exists
}
