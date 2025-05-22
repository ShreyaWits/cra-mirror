package config

import (
	"fmt"
	"messaging_service/internal/modules/message_broker/models"
	"os"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
)

type Config = models.MessaggingConfigResponse
type Env = models.EnvConfig

var (
	config          *Config = &Config{}
	envConfig       *Env    = &Env{}
	validate        *validator.Validate
	SERVICE_NAME    string = "messaging_service"
	SERVICE_VERSION string = "1.0.0"
)

func GetKafkaRetentionMs() int {
	return config.KafkaRetentionMs
}

func init() {
	// Initialize validator
	validate = validator.New()

	// Register field name function for better error messages
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return fld.Name
		}
		return name
	})
}

func SetConfig(cfg *Config) error {
	// Validate configuration before setting it
	if err := validate.Struct(cfg); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	config = cfg
	return nil
}

func GetConfig() *Config {
	return config
}

func GetMockConfig() *Config {
	cfg := &Config{
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

	// Validate mock config to ensure it meets requirements
	if err := validate.Struct(cfg); err != nil {
		panic(fmt.Sprintf("invalid mock configuration: %v", err))
	}

	return cfg
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Env, error) {
	// Load .env file if it exists
	err := godotenv.Load()
	if err != nil {
		return nil, fmt.Errorf("error loading .env file: %w", err)
	}

	// Load environment variables into struct
	envConfig = &models.EnvConfig{
		ConfigServiceUrl:            getEnvString("CONFIG_SERVICE_URL", ""),
		ConfigServiceToken:          getEnvString("MESSAGING_CONFIG_SERVICE_TOKEN", ""),
		Environment:                 getEnvString("ENVIRONMENT", ""),
		HttpPort:                    getEnvString("REST_PORT", ""),
		GrpcPort:                    getEnvString("GRPC_PORT", ""),
		CacheUrl:                    getEnvString("CACHING_SERVICE_GRPC_URL", ""),
		ObservabilityUrl:            getEnvString("OTLEL_COLLECTOR_GRPC_ENDPOINT", ""),
		MESSAGING_SERVICE_REDIS_TTL: getEnvInt("MESSAGING_SERVICE_REDIS_TTL", 240), // Default 24 hours
	}

	// Validate using the validator
	if err := validate.Struct(envConfig); err != nil {
		validationErrors, ok := err.(validator.ValidationErrors)
		if ok {
			// Format validation errors more nicely
			var errorMessages []string
			for _, e := range validationErrors {
				errorMessages = append(errorMessages, fmt.Sprintf("Field '%s' failed validation: %s", e.Field(), e.Tag()))
			}
			return nil, fmt.Errorf("environment validation failed: %s", strings.Join(errorMessages, "; "))
		}
		return nil, fmt.Errorf("environment validation failed: %w", err)
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

// getEnvInt gets an integer environment variable or returns a default value
func getEnvInt(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	var result int
	_, err := fmt.Sscanf(value, "%d", &result)
	if err != nil {
		return defaultValue
	}

	return result
}

// getEnvBool gets a boolean environment variable or returns a default value
func getEnvBool(key string, defaultValue bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	var result bool
	if strings.ToLower(value) == "true" || value == "1" {
		result = true
	} else if strings.ToLower(value) == "false" || value == "0" {
		result = false
	} else {
		return defaultValue
	}

	return result
}

// getEnvFloat gets a float environment variable or returns a default value
func getEnvFloat(key string, defaultValue float64) float64 {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	var result float64
	_, err := fmt.Sscanf(value, "%f", &result)
	if err != nil {
		return defaultValue
	}

	return result
}
