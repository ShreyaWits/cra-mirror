package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Helper to reset env vars for test isolation
func unsetEnvVars(keys ...string) {
	for _, k := range keys {
		os.Unsetenv(k)
	}
}

func Test_getEnv_ReturnsEnvVar(t *testing.T) {
	os.Setenv("TEST_KEY", "test_value")
	defer os.Unsetenv("TEST_KEY")

	val := getEnv("TEST_KEY", "fallback")
	assert.Equal(t, "test_value", val)
}

func Test_getEnv_ReturnsFallback(t *testing.T) {
	os.Unsetenv("TEST_KEY")
	val := getEnv("TEST_KEY", "fallback")
	assert.Equal(t, "fallback", val)
}

func TestLoadEnv_WithEnvVars(t *testing.T) {
	os.Setenv("REDIS_URL", "redis-host:1234")
	os.Setenv("REDIS_PASSWORD", "pass")
	os.Setenv("REDIS_DB", "5")
	os.Setenv("PORT", ":9999")
	os.Setenv("GRPC_PORT", ":8888")
	defer unsetEnvVars("REDIS_URL", "REDIS_PASSWORD", "REDIS_DB", "PORT", "GRPC_PORT")

	LoadEnv()

	assert.Equal(t, "redis-host:1234", REDIS_URL)
	assert.Equal(t, "pass", REDIS_PASSWORD)
	assert.Equal(t, "5", REDIS_DB)
	assert.Equal(t, ":9999", PORT)
	assert.Equal(t, ":8888", GRPC_PORT)
}

func TestLoadEnv_UsesFallbacks(t *testing.T) {
	unsetEnvVars("REDIS_URL", "REDIS_PASSWORD", "REDIS_DB", "PORT", "GRPC_PORT")

	LoadEnv()

	assert.Equal(t, "localhost:6379", REDIS_URL)
	assert.Equal(t, "", REDIS_PASSWORD)
	assert.Equal(t, "0", REDIS_DB)
	assert.Equal(t, ":8080", PORT)
	assert.Equal(t, ":50051", GRPC_PORT)
}

// Optionally, test that godotenv.Load is called (integration test style)
// For pure unit test, you could mock godotenv.Load, but that's not typical in Go
func TestLoadEnv_GodotenvLoadDoesNotPanic(t *testing.T) {
	// This just ensures godotenv.Load is called and doesn't panic
	// (You could use a custom .env file for more advanced integration tests)
	assert.NotPanics(t, func() { LoadEnv() })
}
