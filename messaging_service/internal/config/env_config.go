package config

import (
	"fmt"
	"messaging_service/internal/modules/message_broker/models"
	"os"

	"github.com/joho/godotenv"
)

type Config = models.MessaggingConfigResponse
type Env = models.EnvConfig

var config *Config = &Config{}

var envConfig *models.EnvConfig = &models.EnvConfig{}

func SetConfig(cfg *Config) {
	config = cfg
}

func GetConfig() *Config {
	return config
}

func GetMockConfig() *Config {
	return &Config{

		KafkaBrokers:          []string{"localhost:9092"},
		KafkaAutoCreateTopics: "true",

		KafkaNumPartitions:     3,
		KafkaReplicationFactor: 3,

		KafkaBatchSize:             100,
		KafkaBatchBytes:            1048576,
		KafkaBatchTimeoutMs:        500,
		KafkaCompressionCodec:      "snappy",
		KafkaMaxAttempts:           3,
		KafkaRetryBackoffMs:        100,
		KafkaReadTimeoutMs:         5000,
		KafkaWriteTimeoutMs:        5000,
		KafkaRequireActiveListener: true,

		KafkaRetentionMs: 3000,

		KafkaConsumerMaxWaitMs:        5000,
		KafkaConsumerCommitIntervalMs: 5000,
		KafkaConsumerSessionTimeoutMs: 30000,
		KafkaConsumerHeartbeatMs:      1000,
		KafkaConsumerMaxPollRecords:   1000,
		KafkaConsumerAutoOffsetReset:  "earliest",
		KafkaEnableAutoCommit:         false,
		KafkaIsolationLevel:           "read_committed",
	}
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Env, error) {
	// Load .env file if it exists
	godotenv.Load()

	envConfig = &models.EnvConfig{
		// // Server settings
		ConfigServiceUrl:   getEnvString("CONFIG_SERVICE_URL", ""),
		ConfigServiceToken: getEnvString("CONFIG_SERVICE_TOKEN", ""),
		Environment:        getEnvString("ENVIRONMENT", ""),
		ServiceName:        getEnvString("SERVICE_NAME", ""),
		HttpPort:           getEnvString("HTTP_PORT", ""),
		GrpcPort:           getEnvString("GRPC_PORT", ""),
		CacheUrl:           getEnvString("CACHE_URL", ""),
	}

	// --- Validation ---

	if envConfig.ConfigServiceUrl == "" {
		return nil, fmt.Errorf("CONFIG_SERVICE_URL environment variable is required")
	}

	if envConfig.ConfigServiceToken == "" {
		return nil, fmt.Errorf("CONFIG_SERVICE_TOKEN environment variable is required")
	}
	if envConfig.Environment == "" {
		return nil, fmt.Errorf("ENVIRONMENT environment variable is required")
	}
	if envConfig.ServiceName == "" {
		return nil, fmt.Errorf("SERVICE_NAME environment variable is required")
	}

	if envConfig.HttpPort == "" {
		return nil, fmt.Errorf("HttpPort environment variable is required")
	}

	if envConfig.GrpcPort == "" {
		return nil, fmt.Errorf("GrpcPort environment variable is required")
	}

	if envConfig.CacheUrl == "" {
		return nil, fmt.Errorf("CacheUrl environment variable is required")
	}

	return envConfig, nil
}

// getEnvString gets a string environment variable or returns a default value
func getEnvString(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
