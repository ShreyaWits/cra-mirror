package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// LoadEnv loads environment variables from the .env file
func LoadEnv() {
	if err := godotenv.Load(".env"); err != nil {
		log.Println(err)
		log.Println("❌ Error loading .env file")
	}
	log.Println("✅ Environment variables loaded successfully")
}

// GetEnv retrieves an environment variable or returns a default value
func GetEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
