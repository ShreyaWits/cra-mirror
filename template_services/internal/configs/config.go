package configEnv

import (
	"log"
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
	err := godotenv.Load()
	if err != nil {
		log.Println(".env file not found, falling back to system env")
	}

	// Default values
	config := &Config{
		ServerPort:         getEnv("REST_PORT", "8080"),
		GRPCPort:           getEnv("GRPC_PORT", "50051"),
		RedisHost:          getEnv("REDIS_HOST", "localhost"),
		RedisPort:          getEnv("REDIS_PORT", "6379"),
		RedisUser:          getEnv("REDIS_USER", "templateuser"),
		RedisPassword:      getEnv("REDIS_PASSWORD", "templatepassword"),
		JWTSecret:          getEnv("JWT_SECRET", "myTemplateSecureKey1234567890@GoLan"),
		YugabyteDBHost:     getEnv("YUGABYTE_DATABASE_HOST", "yugabyte"),
		YugabyteDBPort:     getEnv("YUGABYTE_DATABASE_PORT", "5433"),
		YugabyteDBUser:     getEnv("YUGABYTE_DATABASE_USER", "yugabyte"),
		YugabyteDBPassword: getEnv("YUGABYTE_DATABASE_PASSWORD", "yugabyte"),
		YugabyteDBName:     getEnv("TEMPLATE_SERVICE_YUGABYTE_DATABASE_NAME", "yugabyte"),
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
