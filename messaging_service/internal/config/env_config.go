package config

import (
	"fmt"
	"os"
	"strings" // Need to import strings for splitting

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application
type Config struct {
	// Server settings
	GrpcPort string

	// Kafka settings for client connection
	KafkaBrokers          []string // Corrected name for clarity and consistency
	KafkaAutoCreateTopics string   // Keeping as string to match env, though bool might be better
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	// Load .env file if it exists
	// godotenv.Load() will not return an error if the .env file is not found,
	// which is useful for environments where env vars are set directly.
	godotenv.Load()

	config := &Config{
		GrpcPort: getEnvString("GRPC_PORT", "50051"),
		// Load and split the KAFKA_BROKERS string into a slice
		KafkaBrokers:          getEnvStringSlice("KAFKA_BROKERS", "localhost:9092"), // Default should be a single broker or match your .env
		KafkaAutoCreateTopics: getEnvString("KAFKA_AUTO_CREATE_TOPICS_ENABLE", "true"),
	}

	// --- Validation ---

	if config.GrpcPort == "" {
		return nil, fmt.Errorf("GRPC_PORT environment variable is required")
	}

	if len(config.KafkaBrokers) == 0 {
		// Check if the original environment variable was empty after splitting
		kafkaBrokersEnv := os.Getenv("KAFKA_BROKERS")
		if kafkaBrokersEnv == "" {
			return nil, fmt.Errorf("KAFKA_BROKERS environment variable is required")
		}
		// If env var was not empty but splitting resulted in empty slice, something is wrong.
		// This might happen if the env var is just whitespace or delimiters.
		return nil, fmt.Errorf("KAFKA_BROKERS environment variable contains no valid broker addresses")
	}

	// Removed validation for Zookeeper and Kafka broker-specific settings
	// as these are not needed by the client application.

	return config, nil
}

// getEnvString gets a string environment variable or returns a default value
func getEnvString(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// getEnvStringSlice gets a string environment variable, splits it by comma,
// and returns a slice of strings. Returns default if the env var is empty.
func getEnvStringSlice(key, defaultValue string) []string {
	value := os.Getenv(key)
	if value == "" {
		value = defaultValue
	}
	// Handle case where default is also empty or just whitespace
	if value == "" {
		return []string{}
	}
	// Split by comma and trim whitespace from each part
	parts := strings.Split(value, ",")
	var result []string
	for _, part := range parts {
		trimmedPart := strings.TrimSpace(part)
		if trimmedPart != "" {
			result = append(result, trimmedPart)
		}
	}
	return result
}

// Removed the old getEnv function as more specific helpers are used.
