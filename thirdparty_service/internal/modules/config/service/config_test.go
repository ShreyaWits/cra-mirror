package service

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"thirdparty_service/internal/modules/config/dto"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
)

// Mock HTTP client and server setup
func setupMockServer(handler http.HandlerFunc) *httptest.Server {
	server := httptest.NewServer(handler)
	return server
}

// Helper function to create a ConfigServiceImpl with a mock HTTP client
func newTestConfigService(serverURL string) *ConfigServiceImpl {
	return &ConfigServiceImpl{
		Environment:           "test",
		ServiceName:           "test-service",
		ConfigServiceURL:      serverURL,
		ConfigServiceUsername: "testuser",
		ConfigServicePassword: "testpassword",
		HttpClient:            http.DefaultClient, // Use default client for httptest
	}
}

func TestNewConfigService(t *testing.T) {
	env := "dev"
	serviceName := "my-app"
	url := "http://localhost:8080"
	username := "admin"
	password := "password"

	cfgService := NewConfigService(env, serviceName, url, username, password)

	assert.NotNil(t, cfgService)

	impl, ok := cfgService.(*ConfigServiceImpl)
	assert.True(t, ok)
	assert.Equal(t, env, impl.Environment)
	assert.Equal(t, serviceName, impl.ServiceName)
	assert.Equal(t, url, impl.ConfigServiceURL)
	assert.Equal(t, username, impl.ConfigServiceUsername)
	assert.Equal(t, password, impl.ConfigServicePassword)
	assert.NotNil(t, impl.HttpClient)
}

func TestLoginToConfigService_Success(t *testing.T) {
	server := setupMockServer(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/admin/login", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		bodyBytes, _ := io.ReadAll(r.Body)
		var payload map[string]string
		json.Unmarshal(bodyBytes, &payload)
		assert.Equal(t, "testuser", payload["username"])
		assert.Equal(t, "testpassword", payload["password"])

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"token":   "mock-token",
			"success": true,
			"message": "Login successful",
		})
	})
	defer server.Close()

	cfgService := newTestConfigService(server.URL)

	token, err := cfgService.LoginToConfigService()

	assert.NoError(t, err)
	assert.Equal(t, "mock-token", token)
}

func TestLoginToConfigService_Failure(t *testing.T) {
	server := setupMockServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Invalid credentials",
		})
	})
	defer server.Close()

	cfgService := newTestConfigService(server.URL)

	token, err := cfgService.LoginToConfigService()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "login failed: status code 401")
	assert.Empty(t, token)
}

func TestFetchDynamicConfig_Success(t *testing.T) {
	server := setupMockServer(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/config/test/test-service", r.URL.Path)
		assert.Equal(t, "Bearer mock-token", r.Header.Get("Authorization"))

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(dto.ConfigResponse{
			HTTPListenAddress: "localhost",
			HTTPListenPort:    8080,
			GRPCListenAddress: "localhost",
			GRPCListenPort:    50051,
			DatabaseHost:      "db.example.com",
			DatabasePort:      5432,
			DatabaseUser:      "user",
			DatabasePassword:  "password",
			DatabaseName:      "mydb",
			RedisHost:         "redis.example.com",
			RedisPort:         6379,
			TwilioAccountSID:  "ACxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
			TwilioAuthToken:   "your_auth_token",
			TwilioFormNumber:  "+15017122661",
			SendGridApiKey:    "SG.xxxxxxxxxxxxxxxxxxxxxxxxxxxx",
			SendGridFromEmail: "test@example.com",
			SendGridFromName:  "Test User",
			SendWhatsAppMessageSID:        "ACxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
			SendWhatsAppMessageToken:      "your_auth_token",
			SendWhatsAppMessageFromNumber: "+15017122661",
			PushNotificationAccountCreds: "creds",
			PushNotificationProjectID:    "project-id",
		})
	})
	defer server.Close()

	cfgService := newTestConfigService(server.URL)

	config, err := cfgService.FetchDynamicConfig("mock-token")

	assert.NoError(t, err)
	assert.NotNil(t, config)
	assert.Equal(t, "localhost", config.HTTPListenAddress)
	assert.Equal(t, 8080, config.HTTPListenPort)
	// Add assertions for other fields as needed
}

func TestFetchDynamicConfig_Failure(t *testing.T) {
	server := setupMockServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "Internal server error",
		})
	})
	defer server.Close()

	cfgService := newTestConfigService(server.URL)

	config, err := cfgService.FetchDynamicConfig("mock-token")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "config fetch failed: status code 500")
	assert.Nil(t, config)
}

func TestValidateConfig_Success(t *testing.T) {
	cfgService := &ConfigServiceImpl{} // Validator doesn't need other fields

	validConfig := dto.ConfigResponse{
		HTTPListenAddress: "localhost",
		HTTPListenPort:    8080,
		GRPCListenAddress: "localhost",
		GRPCListenPort:    50051,
		DatabaseHost:      "db.example.com",
		DatabasePort:      5432,
		DatabaseUser:      "user",
		DatabasePassword:  "password",
		DatabaseName:      "mydb",
		RedisHost:         "redis.example.com",
		RedisPort:         6379,
		TwilioAccountSID:  "ACxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
		TwilioAuthToken:   "your_auth_token",
		TwilioFormNumber:  "+15017122661",
		SendGridApiKey:    "SG.xxxxxxxxxxxxxxxxxxxxxxxxxxxx",
		SendGridFromEmail: "test@example.com",
		SendGridFromName:  "Test User",
		SendWhatsAppMessageSID:        "ACxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
		SendWhatsAppMessageToken:      "your_auth_token",
		SendWhatsAppMessageFromNumber: "+15017122661",
		PushNotificationAccountCreds: "creds",
		PushNotificationProjectID:    "project-id",
	}

	err := cfgService.ValidateConfig(validConfig)

	assert.NoError(t, err)
}

func TestValidateConfig_Failure(t *testing.T) {
	cfgService := &ConfigServiceImpl{} // Validator doesn't need other fields

	// Test case with missing required fields
	invalidConfig := dto.ConfigResponse{
		// Missing HTTPListenAddress, GRPCListenAddress, DatabaseHost, etc.
		HTTPListenPort: 8080, // Provide one field to show it's not just an empty struct
	}

	err := cfgService.ValidateConfig(invalidConfig)

	assert.Error(t, err)
	assert.IsType(t, validator.ValidationErrors{}, err)

	// You can add more specific test cases for individual missing fields if needed
	// For example:
	// invalidConfigMissingDBHost := dto.ConfigResponse{
	// 	HTTPListenAddress: "localhost",
	// 	HTTPListenPort:    8080,
	// 	GRPCListenAddress: "localhost",
	// 	GRPCListenPort:    50051,
	// 	// DatabaseHost is missing
	// 	DatabasePort:      5432,
	// 	DatabaseUser:      "user",
	// 	DatabasePassword:  "password",
	// 	DatabaseName:      "mydb",
	// 	RedisHost:         "redis.example.com",
	// 	RedisPort:         6379,
	// 	TwilioAccountSID:  "ACxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
	// 	TwilioAuthToken:   "your_auth_token",
	// 	TwilioFormNumber:  "+15017122661",
	// 	SendGridApiKey:    "SG.xxxxxxxxxxxxxxxxxxxxxxxxxxxx",
	// 	SendGridFromEmail: "test@example.com",
	// 	SendGridFromName:  "Test User",
	// 	SendWhatsAppMessageSID:        "ACxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
	// 	SendWhatsAppMessageToken:      "your_auth_token",
	// 	SendWhatsAppMessageFromNumber: "+15017122661",
	// 	PushNotificationAccountCreds: "creds",
	// 	PushNotificationProjectID:    "project-id",
	// }
	// err = cfgService.ValidateConfig(invalidConfigMissingDBHost)
	// assert.Error(t, err)
	// validationErrors, ok := err.(validator.ValidationErrors)
	// assert.True(t, ok)
	// assert.Len(t, validationErrors, 1) // Expecting one error for the missing field
	// assert.Equal(t, "required", validationErrors[0].Tag())
	// assert.Equal(t, "DatabaseHost", validationErrors[0].Field())
}
