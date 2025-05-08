package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Exported variables
var (
	REDIS_URL      string
	REDIS_PASSWORD string
	REDIS_DB       string
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
}

func getEnv(key, fallback string) string {
	exists := os.Getenv(key)
	if exists == "" {
		return fallback
	}
	return exists
}
