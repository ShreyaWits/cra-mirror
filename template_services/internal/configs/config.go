package configEnv

import (
	"os"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application
type Config struct {
	// Server settings
	ServerPort string
	GRPCPort   string
	// Database settings
	RedisHost     string
	RedisPort     string
	RedisUser     string
	RedisPassword string
	JWTSecret     string
	// YugabyteDB settings
	YugabyteDBHost     string
	YugabyteDBPort     string
	YugabyteDBUser     string
	YugabyteDBPassword string
	YugabyteDBName     string
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	// Load .env file if it exists
	godotenv.Load()

	// Default values
	config := &Config{
		ServerPort:         getEnv("PORT", "8080"),
		GRPCPort:           getEnv("GRPC_PORT", "50051"),
		RedisHost:          getEnv("REDIS_HOST", "localhost"),
		RedisPort:          getEnv("REDIS_PORT", "6379"),
		RedisUser:          getEnv("REDIS_USER", "templateuser"),
		RedisPassword:      getEnv("REDIS_PASSWORD", "templatepassword"),
		JWTSecret:          getEnv("JWT_SECRET", "myTemplateSecureKey1234567890@GoLan"),
		YugabyteDBHost:     getEnv("YUGABYTEDB_HOST", "yugabyte"),
		YugabyteDBPort:     getEnv("YUGABYTEDB_PORT", "5433"),
		YugabyteDBUser:     getEnv("YUGABYTEDB_USER", "yugabyte"),
		YugabyteDBPassword: getEnv("YUGABYTEDB_PASSWORD", "yugabyte"),
		YugabyteDBName:     getEnv("YUGABYTEDB_NAME", "yugabyte"),
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
