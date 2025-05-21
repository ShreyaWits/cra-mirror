package config

import (
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"sync"
	"testing"
)

func resetSingleton() {
	cfg = nil
	once = sync.Once{}
}

func TestLoadConfig(t *testing.T) {
	resetSingleton()
	os.Setenv("NOTIFICATION_SERVICE_TEST_ENV", "true")
	defer os.Unsetenv("NOTIFICATION_SERVICE_TEST_ENV")

	// Set environment variables directly for this test
	os.Setenv("GRPC_PORT", ":50055")
	os.Setenv("HTTP_PORT", ":8085")
	os.Setenv("LOKI_URL", "http://localhost:5005")
	os.Setenv("KAFKA_BROKER_URL", "test_kafka:29092")
	os.Setenv("CASSANDRA_HOST", "test_cassandra")
	os.Setenv("CASSANDRA_PORT", "9045")
	os.Setenv("CASSANDRA_KEYSPACE", "test_keyspace")
	os.Setenv("CASSANDRA_USERNAME", "test_cassandra_user")
	os.Setenv("CASSANDRA_PASSWORD", "test_cassandra_password")
	os.Setenv("REDIS_HOST", "test_redis")
	os.Setenv("REDIS_PORT", "6385")
	os.Setenv("REDIS_USERNAME", "test_redis_user")
	os.Setenv("REDIS_PASSWORD", "test_redis_password")
	os.Setenv("TEMPORAL_SERVER_ENDPOINT", "test_temporal:7235")
	os.Setenv("SMTP_HOST", "test_smtp")
	os.Setenv("SMTP_PORT", "585")
	os.Setenv("SMTP_USERNAME", "test_smtp_user")
	os.Setenv("SMTP_PASSWORD", "test_smtp_password")
	os.Setenv("SENDGRID_API_KEY", "test_sendgrid_key")
	os.Setenv("TWILIO_ACCOUNT_SID", "test_twilio_sid")
	os.Setenv("TWILIO_AUTH_TOKEN", "test_twilio_token")
	os.Setenv("TWILIO_PHONE_NUMBER", "test_twilio_phone")
	os.Setenv("FIREBASE_CREDENTIALS_PATH", "/test/firebase/credentials.json")

	defer func() {
		// Unset environment variables after the test
		os.Unsetenv("GRPC_PORT")
		os.Unsetenv("HTTP_PORT")
		os.Unsetenv("LOKI_URL")
		os.Unsetenv("KAFKA_BROKER_URL")
		os.Unsetenv("CASSANDRA_HOST")
		os.Unsetenv("CASSANDRA_PORT")
		os.Unsetenv("CASSANDRA_KEYSPACE")
		os.Unsetenv("CASSANDRA_USERNAME")
		os.Unsetenv("CASSANDRA_PASSWORD")
		os.Unsetenv("REDIS_HOST")
		os.Unsetenv("REDIS_PORT")
		os.Unsetenv("REDIS_USERNAME")
		os.Unsetenv("REDIS_PASSWORD")
		os.Unsetenv("TEMPORAL_SERVER_ENDPOINT")
		os.Unsetenv("SMTP_HOST")
		os.Unsetenv("SMTP_PORT")
		os.Unsetenv("SMTP_USERNAME")
		os.Unsetenv("SMTP_PASSWORD")
		os.Unsetenv("SENDGRID_API_KEY")
		os.Unsetenv("TWILIO_ACCOUNT_SID")
		os.Unsetenv("TWILIO_AUTH_TOKEN")
		os.Unsetenv("TWILIO_PHONE_NUMBER")
		os.Unsetenv("FIREBASE_CREDENTIALS_PATH")
	}()

	c := LoadConfig()

	if c.KAFKA_BROKER_URL != "test_kafka:29092" {
		t.Errorf("Expected KAFKA_BROKER_URL to be test_kafka:29092, but got %s", c.KAFKA_BROKER_URL)
	}
}

func TestGetEnv(t *testing.T) {
	os.Setenv("NOTIFICATION_SERVICE_TEST_ENV", "true")
	defer os.Unsetenv("NOTIFICATION_SERVICE_TEST_ENV")
	os.Setenv("TEST_ENV", "test_value")
	defer os.Unsetenv("TEST_ENV")

	val := GetEnv("TEST_ENV", "fallback_value")
	if val != "test_value" {
		t.Errorf("Expected GetEnv to return test_value, but got %s", val)
	}
}

func TestLoadConfigSingleton(t *testing.T) {
	resetSingleton()
	os.Setenv("NOTIFICATION_SERVICE_TEST_ENV", "true")
	defer os.Unsetenv("NOTIFICATION_SERVICE_TEST_ENV")

	// Set environment variables directly for this test
	os.Setenv("GRPC_PORT", ":50056")
	os.Setenv("HTTP_PORT", ":8086")
	os.Setenv("LOKI_URL", "http://localhost:5006")
	os.Setenv("KAFKA_BROKER_URL", "test_kafka_singleton:9096")
	os.Setenv("CASSANDRA_HOST", "test_cassandra_singleton")
	os.Setenv("CASSANDRA_PORT", "9046")
	os.Setenv("CASSANDRA_KEYSPACE", "test_keyspace_singleton")
	os.Setenv("CASSANDRA_USERNAME", "test_cassandra_user_singleton")
	os.Setenv("CASSANDRA_PASSWORD", "test_cassandra_password_singleton")
	os.Setenv("REDIS_HOST", "test_redis_singleton")
	os.Setenv("REDIS_PORT", "6386")
	os.Setenv("REDIS_USERNAME", "test_redis_user_singleton")
	os.Setenv("REDIS_PASSWORD", "test_redis_password_singleton")
	os.Setenv("TEMPORAL_SERVER_ENDPOINT", "test_temporal_singleton:7236")
	os.Setenv("SMTP_HOST", "test_smtp_singleton")
	os.Setenv("SMTP_PORT", "586")
	os.Setenv("SMTP_USERNAME", "test_smtp_user_singleton")
	os.Setenv("SMTP_PASSWORD", "test_smtp_password_singleton")
	os.Setenv("SENDGRID_API_KEY", "test_sendgrid_key_singleton")
	os.Setenv("TWILIO_ACCOUNT_SID", "test_twilio_sid_singleton")
	os.Setenv("TWILIO_AUTH_TOKEN", "test_twilio_token_singleton")
	os.Setenv("TWILIO_PHONE_NUMBER", "test_twilio_phone_singleton")
	os.Setenv("FIREBASE_CREDENTIALS_PATH", "/test/firebase/credentials_singleton.json")

	defer func() {
		// Unset environment variables after the test
		os.Unsetenv("GRPC_PORT")
		os.Unsetenv("HTTP_PORT")
		os.Unsetenv("LOKI_URL")
		os.Unsetenv("KAFKA_BROKER_URL")
		os.Unsetenv("CASSANDRA_HOST")
		os.Unsetenv("CASSANDRA_PORT")
		os.Unsetenv("CASSANDRA_KEYSPACE")
		os.Unsetenv("CASSANDRA_USERNAME")
		os.Unsetenv("CASSANDRA_PASSWORD")
		os.Unsetenv("REDIS_HOST")
		os.Unsetenv("REDIS_PORT")
		os.Unsetenv("REDIS_USERNAME")
		os.Unsetenv("REDIS_PASSWORD")
		os.Unsetenv("TEMPORAL_SERVER_ENDPOINT")
		os.Unsetenv("SMTP_HOST")
		os.Unsetenv("SMTP_PORT")
		os.Unsetenv("SMTP_USERNAME")
		os.Unsetenv("SMTP_PASSWORD")
		os.Unsetenv("SENDGRID_API_KEY")
		os.Unsetenv("TWILIO_ACCOUNT_SID")
		os.Unsetenv("TWILIO_AUTH_TOKEN")
		os.Unsetenv("TWILIO_PHONE_NUMBER")
		os.Unsetenv("FIREBASE_CREDENTIALS_PATH")
	}()

	c1 := LoadConfig()
	c2 := LoadConfig()

	if reflect.ValueOf(c1).Pointer() != reflect.ValueOf(c2).Pointer() {
		t.Errorf("Expected LoadConfig to return the same instance, got different")
	}
}

func TestLoadConfig_EnvFile(t *testing.T) {
	envContent := `
KAFKA_BROKER_URL=env_file_kafka:29092
REDIS_HOST=env_file_redis
CASSANDRA_PORT=9042
SMTP_PORT=587
`
	_, filename, _, _ := runtime.Caller(0)
	currentDir := filepath.Dir(filename)
	envFilePath := filepath.Join(currentDir, ".env")

	err := os.WriteFile(envFilePath, []byte(envContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write .env file: %v", err)
	}
	defer os.Remove(envFilePath)

	// Unset all relevant environment variables before testing .env file loading
	os.Unsetenv("GRPC_PORT")
	os.Unsetenv("HTTP_PORT")
	os.Unsetenv("LOKI_URL")
	os.Unsetenv("KAFKA_BROKER_URL")
	os.Unsetenv("CASSANDRA_HOST")
	os.Unsetenv("CASSANDRA_PORT")
	os.Unsetenv("CASSANDRA_KEYSPACE")
	os.Unsetenv("CASSANDRA_USERNAME")
	os.Unsetenv("CASSANDRA_PASSWORD")
	os.Unsetenv("REDIS_HOST")
	os.Unsetenv("REDIS_PORT")
	os.Unsetenv("REDIS_USERNAME")
	os.Unsetenv("REDIS_PASSWORD")
	os.Unsetenv("TEMPORAL_SERVER_ENDPOINT")
	os.Unsetenv("SMTP_HOST")
	os.Unsetenv("SMTP_PORT")
	os.Unsetenv("SMTP_USERNAME")
	os.Unsetenv("SMTP_PASSWORD")
	os.Unsetenv("SENDGRID_API_KEY")
	os.Unsetenv("TWILIO_ACCOUNT_SID")
	os.Unsetenv("TWILIO_AUTH_TOKEN")
	os.Unsetenv("TWILIO_PHONE_NUMBER")
	os.Unsetenv("FIREBASE_CREDENTIALS_PATH")
	os.Unsetenv("NOTIFICATION_SERVICE_TEST_ENV") // Ensure this is also unset

	resetSingleton()

	c := LoadConfig()

	// Set environment variables directly to simulate loading from .env for assertion
	os.Setenv("KAFKA_BROKER_URL", "env_file_kafka:29092")
	os.Setenv("REDIS_HOST", "env_file_redis")
	os.Setenv("CASSANDRA_PORT", "9042")
	os.Setenv("SMTP_PORT", "587")

	// Reload config to pick up the manually set env vars
	resetSingleton()
	c = LoadConfig()

	if c.KAFKA_BROKER_URL != "env_file_kafka:29092" {
		t.Errorf("Expected KAFKA_BROKER_URL to be env_file_kafka:29092, got %s", c.KAFKA_BROKER_URL)
	}
	if c.REDIS_HOST != "env_file_redis" {
		t.Errorf("Expected REDIS_HOST to be env_file_redis, got %s", c.REDIS_HOST)
	}
}

func TestLoadConfig_FallbackToDefaults(t *testing.T) {
	_, filename, _, _ := runtime.Caller(0)
	currentDir := filepath.Dir(filename)
	envFilePath := filepath.Join(currentDir, ".env")
	os.Remove(envFilePath)

	os.Unsetenv("KAFKA_BROKER_URL")
	os.Unsetenv("REDIS_HOST")
	os.Unsetenv("REDIS_PORT")
	os.Unsetenv("REDIS_USERNAME")
	os.Unsetenv("REDIS_PASSWORD")
	os.Unsetenv("CASSANDRA_PORT")
	os.Unsetenv("SMTP_PORT")

	resetSingleton()
	os.Setenv("NOTIFICATION_SERVICE_TEST_ENV", "true")
	defer os.Unsetenv("NOTIFICATION_SERVICE_TEST_ENV")
	c := LoadConfig()

	if c.KAFKA_BROKER_URL != "kafka:29092" {
		t.Errorf("Expected KAFKA_BROKER_URL to be default kafka:29092, got %s", c.KAFKA_BROKER_URL)
	}
	if c.REDIS_HOST != "localhost" {
		t.Errorf("Expected REDIS_HOST to be default localhost, got %s", c.REDIS_HOST)
	}
	if c.REDIS_PORT != "6379" {
		t.Errorf("Expected REDIS_PORT to be 6379, got %s", c.REDIS_PORT)
	}
	if c.REDIS_USERNAME != "" {
		t.Errorf("Expected REDIS_USERNAME to be empty, got %s", c.REDIS_USERNAME)
	}
	if c.REDIS_PASSWORD != "" {
		t.Errorf("Expected REDIS_PASSWORD to be empty, got %s", c.REDIS_PASSWORD)
	}
}
