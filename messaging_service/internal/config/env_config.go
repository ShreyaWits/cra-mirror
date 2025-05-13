package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application
type Config struct {
	// Server settings
	GrpcPort string

	// Kafka settings for client connection
	KafkaBrokers          []string
	KafkaAutoCreateTopics string

	// Kafka topic configuration
	KafkaNumPartitions     int
	KafkaReplicationFactor int

	// Kafka producer configuration
	KafkaBatchSize        int
	KafkaBatchBytes       int64
	KafkaBatchTimeoutMs   int
	KafkaCompressionCodec string
	KafkaMaxAttempts      int
	KafkaRetryBackoffMs   int
	KafkaReadTimeoutMs    int
	KafkaWriteTimeoutMs   int

	// Kafka topic settings
	KafkaRetentionMs int

	// Kafka consumer configuration
	KafkaConsumerMaxWaitMs        int
	KafkaConsumerCommitIntervalMs int
	KafkaConsumerSessionTimeoutMs int
	KafkaConsumerHeartbeatMs      int
	KafkaConsumerMaxPollRecords   int
	KafkaConsumerAutoOffsetReset  string
	KafkaEnableAutoCommit         bool
	KafkaIsolationLevel           string
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	// Load .env file if it exists
	godotenv.Load()

	config := &Config{
		// Server settings
		GrpcPort: getEnvString("GRPC_PORT", "50051"),

		// Kafka connection settings
		KafkaBrokers:          getEnvStringSlice("KAFKA_BROKERS", "localhost:9092"),
		KafkaAutoCreateTopics: getEnvString("KAFKA_AUTO_CREATE_TOPICS_ENABLE", "true"),

		// Kafka topic configuration
		KafkaNumPartitions:     getEnvInt("KAFKA_NUM_PARTITIONS", 3),
		KafkaReplicationFactor: getEnvInt("KAFKA_REPLICATION_FACTOR", 3),

		// Kafka producer configuration
		KafkaBatchSize:        getEnvInt("KAFKA_BATCH_SIZE", 100),
		KafkaBatchBytes:       getEnvInt64("KAFKA_BATCH_BYTES", 1048576),
		KafkaBatchTimeoutMs:   getEnvInt("KAFKA_BATCH_TIMEOUT_MS", 500),
		KafkaCompressionCodec: getEnvString("KAFKA_COMPRESSION_CODEC", "snappy"),
		KafkaMaxAttempts:      getEnvInt("KAFKA_MAX_ATTEMPTS", 3),
		KafkaRetryBackoffMs:   getEnvInt("KAFKA_RETRY_BACKOFF_MS", 100),
		KafkaReadTimeoutMs:    getEnvInt("KAFKA_READ_TIMEOUT_MS", 5000),
		KafkaWriteTimeoutMs:   getEnvInt("KAFKA_WRITE_TIMEOUT_MS", 5000),

		// Kafka topic settings
		KafkaRetentionMs: getEnvInt("KAFKA_RETENTION_MS", 3000),

		// Kafka consumer configuration
		KafkaConsumerMaxWaitMs:        getEnvInt("KAFKA_CONSUMER_MAX_WAIT_MS", 5000),
		KafkaConsumerCommitIntervalMs: getEnvInt("KAFKA_CONSUMER_COMMIT_INTERVAL_MS", 5000),
		KafkaConsumerSessionTimeoutMs: getEnvInt("KAFKA_CONSUMER_SESSION_TIMEOUT_MS", 30000),
		KafkaConsumerHeartbeatMs:      getEnvInt("KAFKA_CONSUMER_HEARTBEAT_MS", 1000),
		KafkaConsumerMaxPollRecords:   getEnvInt("KAFKA_CONSUMER_MAX_POLL_RECORDS", 1000),
		KafkaConsumerAutoOffsetReset:  getEnvString("KAFKA_CONSUMER_AUTO_OFFSET_RESET", "earliest"),
		KafkaEnableAutoCommit:         getEnvBool("KAFKA_ENABLE_AUTO_COMMIT", false),
		KafkaIsolationLevel:           getEnvString("KAFKA_ISOLATION_LEVEL", "read_committed"),
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

// getEnvInt gets an integer environment variable or returns a default value
func getEnvInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
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

// getEnvInt64 gets an int64 environment variable or returns a default value
func getEnvInt64(key string, defaultValue int64) int64 {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.ParseInt(valueStr, 10, 64)
	if err != nil {
		return defaultValue
	}

	return value
}

// getEnvBool gets a boolean environment variable or returns a default value
func getEnvBool(key string, defaultValue bool) bool {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.ParseBool(valueStr)
	if err != nil {
		return defaultValue
	}

	return value
}

// Removed the old getEnv function as more specific helpers are used.
