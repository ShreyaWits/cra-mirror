package config_test

import (
	"Document-Processing/internal/config"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestLoadConfigFromAPI(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request headers
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("Expected Authorization header 'Bearer test-token', got '%s'", r.Header.Get("Authorization"))
		}

		// Return mock response
		response := config.ConfigResponse{
			StatusCode: 200,
			Message:    "Success",
			Data: map[string]string{
				"MINIO_ENDPOINT":    "localhost:9000",
				"MINIO_ACCESS_KEY":  "test-access-key",
				"MINIO_SECRET_KEY":  "test-secret-key",
				"MINIO_BUCKET_NAME": "test-bucket",
				"GEMINI_API_KEY":    "test-gemini-key",
				"SERVER_PORT":       "8080",
				"YUGABYTE_HOST":     "localhost",
				"YUGABYTE_USER":     "test-user",
				"YUGABYTE_PASSWORD": "test-password",
				"YUGABYTE_NAME":     "test-db",
				"YUGABYTE_PORT":     "5433",
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Set environment variables for the test
	os.Setenv("CONFIG_API", server.URL)
	os.Setenv("ENVIRONMENT", "test")
	defer os.Unsetenv("CONFIG_API")
	defer os.Unsetenv("ENVIRONMENT")

	// Test successful case
	data, err := config.LoadConfigFromAPI("test-token")
	if err != nil {
		t.Errorf("LoadConfigFromAPI failed: %v", err)
	}

	// Verify the returned data
	expectedKeys := []string{
		"MINIO_ENDPOINT",
		"MINIO_ACCESS_KEY",
		"MINIO_SECRET_KEY",
		"MINIO_BUCKET_NAME",
		"GEMINI_API_KEY",
		"SERVER_PORT",
		"YUGABYTE_HOST",
		"YUGABYTE_USER",
		"YUGABYTE_PASSWORD",
		"YUGABYTE_NAME",
		"YUGABYTE_PORT",
	}

	for _, key := range expectedKeys {
		if _, exists := data[key]; !exists {
			t.Errorf("Expected key %s not found in response", key)
		}
	}
}

func TestNewConfig(t *testing.T) {
	tests := []struct {
		name    string
		data    map[string]string
		wantErr bool
	}{
		{
			name: "Valid config",
			data: map[string]string{
				"MINIO_ENDPOINT":    "localhost:9000",
				"MINIO_ACCESS_KEY":  "test-access-key",
				"MINIO_SECRET_KEY":  "test-secret-key",
				"MINIO_BUCKET_NAME": "test-bucket",
				"GEMINI_API_KEY":    "test-gemini-key",
				"SERVER_PORT":       "8080",
				"YUGABYTE_HOST":     "localhost",
				"YUGABYTE_USER":     "test-user",
				"YUGABYTE_PASSWORD": "test-password",
				"YUGABYTE_NAME":     "test-db",
				"YUGABYTE_PORT":     "5433",
			},
			wantErr: false,
		},
		{
			name: "Missing required key",
			data: map[string]string{
				"MINIO_ENDPOINT":    "localhost:9000",
				"MINIO_ACCESS_KEY":  "test-access-key",
				"MINIO_SECRET_KEY":  "test-secret-key",
				"MINIO_BUCKET_NAME": "test-bucket",
				"GEMINI_API_KEY":    "test-gemini-key",
				"SERVER_PORT":       "8080",
				"YUGABYTE_HOST":     "localhost",
				"YUGABYTE_USER":     "test-user",
				"YUGABYTE_PASSWORD": "test-password",
				"YUGABYTE_NAME":     "test-db",
				// Missing YUGABYTE_PORT
			},
			wantErr: true,
		},
		{
			name: "Invalid server port",
			data: map[string]string{
				"MINIO_ENDPOINT":    "localhost:9000",
				"MINIO_ACCESS_KEY":  "test-access-key",
				"MINIO_SECRET_KEY":  "test-secret-key",
				"MINIO_BUCKET_NAME": "test-bucket",
				"GEMINI_API_KEY":    "test-gemini-key",
				"SERVER_PORT":       "invalid",
				"YUGABYTE_HOST":     "localhost",
				"YUGABYTE_USER":     "test-user",
				"YUGABYTE_PASSWORD": "test-password",
				"YUGABYTE_NAME":     "test-db",
				"YUGABYTE_PORT":     "5433",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := config.NewConfig(tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && cfg == nil {
				t.Error("NewConfig() returned nil config when no error was expected")
			}
		})
	}
}
