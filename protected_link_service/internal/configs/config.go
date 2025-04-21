package configEnv

import (
	"os"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application
type Config struct {
	// Server settings
	ServerPort string

	// Database settings
	DBHost         string
	DBPort         string
	DBUser         string
	DBPassword     string
	DBName         string
	DBSSLMode      string
	JWTSecret      string
	RedirectionURL string
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	// Load .env file if it exists
	godotenv.Load()

	// Default values
	config := &Config{
		ServerPort: getEnv("PORT", "9090"),

		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "sarbjeet"),
		DBPassword: getEnv("DB_PASSWORD", "sarb"),
		DBName:     getEnv("DB_NAME", "nps"),
		DBSSLMode:  getEnv("DB_SSL_MODE", "require"),
		JWTSecret:  getEnv("JWT_SECRET", "secret"),
		RedirectionURL: getEnv("REDIRECTION_URL", "http://localhost:9090"),
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
