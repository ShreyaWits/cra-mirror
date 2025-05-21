package configEnv

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application
type Config struct {
	// Server settings
	ServerPort        string
	OtpExpiryDuration string

	// Database settings
	REDIS_HOST     string
	REDIS_PORT     string
	REDIS_PASSWORD string
	REDIS_USERNAME string

	JWTSecret        string
	RedirectionURL   string
	Kafka_Broker_Url string
	KafkaProducer    string
	AppEnv           string

	// Cassandra settings
	CASSANDRA_HOST     string
	CASSANDRA_KEYSPACE string
	CASSANDRA_USERNAME string
	CASSANDRA_PASSWORD string
	CASSANDRA_PORT     string
	GRPCPort           string
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
		ServerPort:         getEnv("PORT", "9090"),
		GRPCPort:           getEnv("GRPC_PORT", "50051"),
		REDIS_HOST:         getEnv("REDIS_HOST", "localhost"),
		REDIS_PORT:         getEnv("REDIS_PORT", "6379"),
		REDIS_PASSWORD:     getEnv("REDIS_PASSWORD", ""),
		REDIS_USERNAME:     getEnv("REDIS_USERNAME", ""),
		OtpExpiryDuration:  getEnv("REDIS_TTL", "5m"),
		KafkaProducer:      getEnv("PROTECTED_LINK_SERVICE_KAFKA_PRODUCER_TOPIC", "send_notification"),
		AppEnv:             getEnv("APP_ENV", "local"),
		Kafka_Broker_Url:   getEnv("KAFKA_BROKER_URL", "localhost:9092"),
		JWTSecret:          getEnv("JWT_SECRET", "mySuperSecureKey1234567890@GoLan"),
		RedirectionURL:     getEnv("REDIRECTION_URL", "http://localhost:8080"),
		CASSANDRA_HOST:     getEnv("CASSANDRA_HOST", "cassandra"),
		CASSANDRA_KEYSPACE: getEnv("CASSANDRA_KEYSPACE", "protectedlink"),
		CASSANDRA_USERNAME: getEnv("CASSANDRA_USERNAME", "cassandra"),
		CASSANDRA_PASSWORD: getEnv("CASSANDRA_PASSWORD", "cassandra"),
		CASSANDRA_PORT:     getEnv("CASSANDRA_PORT", "9042"),
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
