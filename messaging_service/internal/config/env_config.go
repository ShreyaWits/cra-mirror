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

	GrpcPort string

	// Zookeeper settings
	ZookeeperVersion    string
	ZookeeperClientPort string
	ZookeeperTickTime   string

	// Kafka settings
	KafkaVersion             string
	KafkaPort                string
	KafkaBrokerID            string
	KafkaZookeeperConnect    string
	KafkaAdvertisedListeners string
	KafkaReplicationFactor   string
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		return nil, fmt.Errorf("error loading .env file: %v", err)
	}

	// Default values
	config := &Config{
		ServerPort:               getEnv("PORT", "8082"),
		GrpcPort:                 getEnv("GRPC_PORT", "50051"),
		ZookeeperVersion:         getEnv("ZOOKEEPER_VERSION", "7.5.0"),
		ZookeeperClientPort:      getEnv("ZOOKEEPER_CLIENT_PORT", "2181"),
		ZookeeperTickTime:        getEnv("ZOOKEEPER_TICK_TIME", "2000"),
		KafkaVersion:             getEnv("KAFKA_VERSION", "7.5.0"),
		KafkaPort:                getEnv("KAFKA_PORT", "9092"),
		KafkaBrokerID:            getEnv("KAFKA_BROKER_ID", "1"),
		KafkaZookeeperConnect:    getEnv("KAFKA_ZOOKEEPER_CONNECT", "zookeeper:2181"),
		KafkaAdvertisedListeners: getEnv("KAFKA_ADVERTISED_LISTENERS", "PLAINTEXT://localhost:9092"),
		KafkaReplicationFactor:   getEnv("KAFKA_REPLICATION_FACTOR", "1"),
	}

	if config.GrpcPort == "" {
		return nil, fmt.Errorf("GRPC_PORT environment variable is required")
	}

	// Validate required Kafka configuration
	if config.KafkaVersion == "" {
		return nil, fmt.Errorf("KAFKA_VERSION environment variable is required")
	}
	if config.KafkaPort == "" {
		return nil, fmt.Errorf("KAFKA_PORT environment variable is required")
	}
	if config.KafkaBrokerID == "" {
		return nil, fmt.Errorf("KAFKA_BROKER_ID environment variable is required")
	}
	if config.KafkaZookeeperConnect == "" {
		return nil, fmt.Errorf("KAFKA_ZOOKEEPER_CONNECT environment variable is required")
	}
	if config.KafkaAdvertisedListeners == "" {
		return nil, fmt.Errorf("KAFKA_ADVERTISED_LISTENERS environment variable is required")
	}
	if config.KafkaReplicationFactor == "" {
		return nil, fmt.Errorf("KAFKA_REPLICATION_FACTOR environment variable is required")
	}

	// Validate required Zookeeper configuration
	if config.ZookeeperVersion == "" {
		return nil, fmt.Errorf("ZOOKEEPER_VERSION environment variable is required")
	}
	if config.ZookeeperClientPort == "" {
		return nil, fmt.Errorf("ZOOKEEPER_CLIENT_PORT environment variable is required")
	}
	if config.ZookeeperTickTime == "" {
		return nil, fmt.Errorf("ZOOKEEPER_TICK_TIME environment variable is required")
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
