package configservice

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

// Helper function to create a ConfigServiceImpl with a custom HTTP client
func newTestConfigService(serverURL string, client *http.Client) *ConfigServiceImpl {
	if client == nil {
		client = http.DefaultClient // Use default client if none is provided
	}
	return &ConfigServiceImpl{
		Environment:           "test",
		ServiceName:           "test-service",
		ConfigServiceURL:      serverURL,
		ConfigServiceUsername: "testuser",
		ConfigServicePassword: "testpassword",
		HttpClient:            client,
	}
}

func TestNewConfigService(t *testing.T) {
	env := "dev"
	serviceName := "my-app"
	url := "http://localhost:8080"
	username := "admin"
	password := "password"

	cfgService := NewConfigService(env, serviceName, url, username, password,)

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

	cfgService := newTestConfigService(server.URL, nil)

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

	cfgService := newTestConfigService(server.URL, nil)

	token, err := cfgService.LoginToConfigService()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "login failed: status code 401")
	assert.Empty(t, token)
}

func TestLoginToConfigService_HttpClientError(t *testing.T) {
	// Create a mock HTTP client that returns an error
	mockClient := &http.Client{
		Transport: &mockTransport{
			Err: assert.AnError, // Simulate a network error
		},
	}

	cfgService := newTestConfigService("http://localhost:8080", mockClient)

	token, err := cfgService.LoginToConfigService()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "sending login request:")
	assert.Empty(t, token)
}

func TestLoginToConfigService_NewRequestError(t *testing.T) {
	cfgService := newTestConfigService("invalid url", nil) // Invalid URL to cause NewRequest error

	token, err := cfgService.LoginToConfigService()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "sending login request:")
	assert.Empty(t, token)
}

func TestLoginToConfigService_InvalidJsonResponse(t *testing.T) {
	server := setupMockServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		// Return invalid JSON
		w.Write([]byte("invalid json"))
	})
	defer server.Close()

	cfgService := newTestConfigService(server.URL, nil)

	token, err := cfgService.LoginToConfigService()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "decoding login response:")
	assert.Empty(t, token)
}

// mockTransport is a mock http.RoundTripper that returns a predefined error
type mockTransport struct {
	Err error
}

func (m *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return nil, m.Err
}

func TestFetchDynamicConfig_Success(t *testing.T) {
	server := setupMockServer(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/config/test/test-service", r.URL.Path)
		assert.Equal(t, "Bearer mock-token", r.Header.Get("Authorization"))

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"token":   "mock-token",
			"success": true,
			"message": "Config fetched successfully",
			"data": dto.ConfigResponse{
				HTTPListenAddress: "localhost",
				HTTPListenPort:    8080,
				GRPCListenAddress: "localhost",
				GRPCListenPort:    50051,
				DatabaseHost:      "db.example.com",
				DatabasePort:      5432,
				DatabaseUser:      "user",
				DatabasePassword:  "password",
				DatabaseName:      "mydb",
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
			},
		})
	})
	defer server.Close()

	cfgService := newTestConfigService(server.URL, nil)

	config, err := cfgService.FetchDynamicConfig("mock-token")

	assert.NoError(t, err)
	assert.NotNil(t, config)
	assert.Equal(t, "localhost", config.HTTPListenAddress)
	assert.Equal(t, 8080, config.HTTPListenPort)
	// Add assertions for other fields as needed
}

func TestFetchDynamicConfig_NewRequestError(t *testing.T) {
	cfgService := newTestConfigService("invalid url", nil) // Invalid URL to cause NewRequest error

	config, err := cfgService.FetchDynamicConfig("mock-token")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "sending config request:")
	assert.Nil(t, config)
}

func TestFetchDynamicConfig_Failure(t *testing.T) {
	server := setupMockServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "Internal server error",
		})
	})
	defer server.Close()

	cfgService := newTestConfigService(server.URL, nil)

	config, err := cfgService.FetchDynamicConfig("mock-token")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "config fetch failed: status code 500")
	assert.Nil(t, config)
}

func TestFetchDynamicConfig_HttpClientError(t *testing.T) {
	// Create a mock HTTP client that returns an error
	mockClient := &http.Client{
		Transport: &mockTransport{
			Err: assert.AnError, // Simulate a network error
		},
	}

	cfgService := newTestConfigService("http://localhost:8080", mockClient)

	config, err := cfgService.FetchDynamicConfig("mock-token")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "sending config request:")
	assert.Nil(t, config)
}

func TestFetchDynamicConfig_InvalidJsonResponse(t *testing.T) {
	server := setupMockServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		// Return invalid JSON
		w.Write([]byte("invalid json"))
	})
	defer server.Close()

	cfgService := newTestConfigService(server.URL, nil)

	config, err := cfgService.FetchDynamicConfig("mock-token")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "decoding config response:")
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

	testCases := []struct {
		name          string
		config        dto.ConfigResponse
		expectedField string
	}{
		{
			name: "Missing HTTPListenAddress",
			config: dto.ConfigResponse{
				HTTPListenPort:    8080,
				GRPCListenAddress: "localhost",
				GRPCListenPort:    50051,
				DatabaseHost:      "db.example.com",
				DatabasePort:      5432,
				DatabaseUser:      "user",
				DatabasePassword:  "password",
				DatabaseName:      "mydb",
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
			},
			expectedField: "HTTPListenAddress",
		},
		{
			name: "Missing HTTPListenPort",
			config: dto.ConfigResponse{
				HTTPListenAddress: "localhost",
				GRPCListenAddress: "localhost",
				GRPCListenPort:    50051,
				DatabaseHost:      "db.example.com",
				DatabasePort:      5432,
				DatabaseUser:      "user",
				DatabasePassword:  "password",
				DatabaseName:      "mydb",
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
			},
			expectedField: "HTTPListenPort",
		},
		{
			name: "Missing GRPCListenAddress",
			config: dto.ConfigResponse{
				HTTPListenAddress: "localhost",
				HTTPListenPort:    8080,
				GRPCListenPort:    50051,
				DatabaseHost:      "db.example.com",
				DatabasePort:      5432,
				DatabaseUser:      "user",
				DatabasePassword:  "password",
				DatabaseName:      "mydb",
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
			},
			expectedField: "GRPCListenAddress",
		},
		{
			name: "Missing GRPCListenPort",
			config: dto.ConfigResponse{
				HTTPListenAddress: "localhost",
				HTTPListenPort:    8080,
				GRPCListenAddress: "localhost",
				DatabaseHost:      "db.example.com",
				DatabasePort:      5432,
				DatabaseUser:      "user",
				DatabasePassword:  "password",
				DatabaseName:      "mydb",
				
				
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
			},
			expectedField: "GRPCListenPort",
		},
		{
			name: "Missing DatabaseHost",
			config: dto.ConfigResponse{
				HTTPListenAddress: "localhost",
				HTTPListenPort:    8080,
				GRPCListenAddress: "localhost",
				GRPCListenPort:    50051,
				DatabasePort:      5432,
				DatabaseUser:      "user",
				DatabasePassword:  "password",
				DatabaseName:      "mydb",
				
				
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
			},
			expectedField: "DatabaseHost",
		},
		{
			name: "Missing DatabasePort",
			config: dto.ConfigResponse{
				HTTPListenAddress: "localhost",
				HTTPListenPort:    8080,
				GRPCListenAddress: "localhost",
				GRPCListenPort:    50051,
				DatabaseHost:      "db.example.com",
				DatabaseUser:      "user",
				DatabasePassword:  "password",
				DatabaseName:      "mydb",
				
				
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
			},
			expectedField: "DatabasePort",
		},
		{
			name: "Missing DatabaseUser",
			config: dto.ConfigResponse{
				HTTPListenAddress: "localhost",
				HTTPListenPort:    8080,
				GRPCListenAddress: "localhost",
				GRPCListenPort:    50051,
				DatabaseHost:      "db.example.com",
				DatabasePort:      5432,
				DatabasePassword:  "password",
				DatabaseName:      "mydb",
				
				
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
			},
			expectedField: "DatabaseUser",
		},
		{
			name: "Missing DatabasePassword",
			config: dto.ConfigResponse{
				HTTPListenAddress: "localhost",
				HTTPListenPort:    8080,
				GRPCListenAddress: "localhost",
				GRPCListenPort:    50051,
				DatabaseHost:      "db.example.com",
				DatabasePort:      5432,
				DatabaseUser:      "user",
				DatabaseName:      "mydb",
				
				
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
			},
			expectedField: "DatabasePassword",
		},
		{
			name: "Missing DatabaseName",
			config: dto.ConfigResponse{
				HTTPListenAddress: "localhost",
				HTTPListenPort:    8080,
				GRPCListenAddress: "localhost",
				GRPCListenPort:    50051,
				DatabaseHost:      "db.example.com",
				DatabasePort:      5432,
				DatabaseUser:      "user",
				DatabasePassword:  "password",
				
				
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
			},
			expectedField: "DatabaseName",
		},
		{
			name: "Missing RedisHost",
			config: dto.ConfigResponse{
				HTTPListenAddress: "localhost",
				HTTPListenPort:    8080,
				GRPCListenAddress: "localhost",
				GRPCListenPort:    50051,
				DatabaseHost:      "db.example.com",
				DatabasePort:      5432,
				DatabaseUser:      "user",
				DatabasePassword:  "password",
				DatabaseName:      "mydb",
				
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
			},
			expectedField: "RedisHost",
		},
		{
			name: "Missing RedisPort",
			config: dto.ConfigResponse{
				HTTPListenAddress: "localhost",
				HTTPListenPort:    8080,
				GRPCListenAddress: "localhost",
				GRPCListenPort:    50051,
				DatabaseHost:      "db.example.com",
				DatabasePort:      5432,
				DatabaseUser:      "user",
				DatabasePassword:  "password",
				DatabaseName:      "mydb",
				
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
			},
			expectedField: "RedisPort",
		},
		{
			name: "Missing TwilioAccountSID",
			config: dto.ConfigResponse{
				HTTPListenAddress: "localhost",
				HTTPListenPort:    8080,
				GRPCListenAddress: "localhost",
				GRPCListenPort:    50051,
				DatabaseHost:      "db.example.com",
				DatabasePort:      5432,
				DatabaseUser:      "user",
				DatabasePassword:  "password",
				DatabaseName:      "mydb",
				
				
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
			},
			expectedField: "TwilioAccountSID",
		},
		{
			name: "Missing TwilioAuthToken",
			config: dto.ConfigResponse{
				HTTPListenAddress: "localhost",
				HTTPListenPort:    8080,
				GRPCListenAddress: "localhost",
				GRPCListenPort:    50051,
				DatabaseHost:      "db.example.com",
				DatabasePort:      5432,
				DatabaseUser:      "user",
				DatabasePassword:  "password",
				DatabaseName:      "mydb",
				
				
				TwilioAccountSID:  "ACxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
				TwilioFormNumber:  "+15017122661",
				SendGridApiKey:    "SG.xxxxxxxxxxxxxxxxxxxxxxxxxxxx",
				SendGridFromEmail: "test@example.com",
				SendGridFromName:  "Test User",
				SendWhatsAppMessageSID:        "ACxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
				SendWhatsAppMessageToken:      "your_auth_token",
				SendWhatsAppMessageFromNumber: "+15017122661",
				PushNotificationAccountCreds: "creds",
				PushNotificationProjectID:    "project-id",
			},
			expectedField: "TwilioAuthToken",
		},
		{
			name: "Missing TwilioFormNumber",
			config: dto.ConfigResponse{
				HTTPListenAddress: "localhost",
				HTTPListenPort:    8080,
				GRPCListenAddress: "localhost",
				GRPCListenPort:    50051,
				DatabaseHost:      "db.example.com",
				DatabasePort:      5432,
				DatabaseUser:      "user",
				DatabasePassword:  "password",
				DatabaseName:      "mydb",
				
				
				TwilioAccountSID:  "ACxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
				TwilioAuthToken:   "your_auth_token",
				SendGridApiKey:    "SG.xxxxxxxxxxxxxxxxxxxxxxxxxxxx",
				SendGridFromEmail: "test@example.com",
				SendGridFromName:  "Test User",
				SendWhatsAppMessageSID:        "ACxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
				SendWhatsAppMessageToken:      "your_auth_token",
				SendWhatsAppMessageFromNumber: "+15017122661",
				PushNotificationAccountCreds: "creds",
				PushNotificationProjectID:    "project-id",
			},
			expectedField: "TwilioFormNumber",
		},
		{
			name: "Missing SendGridApiKey",
			config: dto.ConfigResponse{
				HTTPListenAddress: "localhost",
				HTTPListenPort:    8080,
				GRPCListenAddress: "localhost",
				GRPCListenPort:    50051,
				DatabaseHost:      "db.example.com",
				DatabasePort:      5432,
				DatabaseUser:      "user",
				DatabasePassword:  "password",
				DatabaseName:      "mydb",
				
				
				TwilioAccountSID:  "ACxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
				TwilioAuthToken:   "your_auth_token",
				TwilioFormNumber:  "+15017122661",
				SendGridFromEmail: "test@example.com",
				SendGridFromName:  "Test User",
				SendWhatsAppMessageSID:        "ACxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
				SendWhatsAppMessageToken:      "your_auth_token",
				SendWhatsAppMessageFromNumber: "+15017122661",
				PushNotificationAccountCreds: "creds",
				PushNotificationProjectID:    "project-id",
			},
			expectedField: "SendGridApiKey",
		},
		{
			name: "Missing SendGridFromEmail",
			config: dto.ConfigResponse{
				HTTPListenAddress: "localhost",
				HTTPListenPort:    8080,
				GRPCListenAddress: "localhost",
				GRPCListenPort:    50051,
				DatabaseHost:      "db.example.com",
				DatabasePort:      5432,
				DatabaseUser:      "user",
				DatabasePassword:  "password",
				DatabaseName:      "mydb",
				
				
				TwilioAccountSID:  "ACxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
				TwilioAuthToken:   "your_auth_token",
				TwilioFormNumber:  "+15017122661",
				SendGridApiKey:    "SG.xxxxxxxxxxxxxxxxxxxxxxxxxxxx",
				SendGridFromName:  "Test User",
				SendWhatsAppMessageSID:        "ACxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
				SendWhatsAppMessageToken:      "your_auth_token",
				SendWhatsAppMessageFromNumber: "+15017122661",
				PushNotificationAccountCreds: "creds",
				PushNotificationProjectID:    "project-id",
			},
			expectedField: "SendGridFromEmail",
		},
		{
			name: "Missing SendGridFromName",
			config: dto.ConfigResponse{
				HTTPListenAddress: "localhost",
				HTTPListenPort:    8080,
				GRPCListenAddress: "localhost",
				GRPCListenPort:    50051,
				DatabaseHost:      "db.example.com",
				DatabasePort:      5432,
				DatabaseUser:      "user",
				DatabasePassword:  "password",
				DatabaseName:      "mydb",
				
				
				TwilioAccountSID:  "ACxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
				TwilioAuthToken:   "your_auth_token",
				TwilioFormNumber:  "+15017122661",
				SendGridApiKey:    "SG.xxxxxxxxxxxxxxxxxxxxxxxxxxxx",
				SendGridFromEmail: "test@example.com",
				SendWhatsAppMessageSID:        "ACxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
				SendWhatsAppMessageToken:      "your_auth_token",
				SendWhatsAppMessageFromNumber: "+15017122661",
				PushNotificationAccountCreds: "creds",
				PushNotificationProjectID:    "project-id",
			},
			expectedField: "SendGridFromName",
		},
		{
			name: "Missing SendWhatsAppMessageSID",
			config: dto.ConfigResponse{
				HTTPListenAddress: "localhost",
				HTTPListenPort:    8080,
				GRPCListenAddress: "localhost",
				GRPCListenPort:    50051,
				DatabaseHost:      "db.example.com",
				DatabasePort:      5432,
				DatabaseUser:      "user",
				DatabasePassword:  "password",
				DatabaseName:      "mydb",
				
				
				TwilioAccountSID:  "ACxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
				TwilioAuthToken:   "your_auth_token",
				TwilioFormNumber:  "+15017122661",
				SendGridApiKey:    "SG.xxxxxxxxxxxxxxxxxxxxxxxxxxxx",
				SendGridFromEmail: "test@example.com",
				SendGridFromName:  "Test User",
				SendWhatsAppMessageToken:      "your_auth_token",
				SendWhatsAppMessageFromNumber: "+15017122661",
				PushNotificationAccountCreds: "creds",
				PushNotificationProjectID:    "project-id",
			},
			expectedField: "SendWhatsAppMessageSID",
		},
		{
			name: "Missing SendWhatsAppMessageToken",
			config: dto.ConfigResponse{
				HTTPListenAddress: "localhost",
				HTTPListenPort:    8080,
				GRPCListenAddress: "localhost",
				GRPCListenPort:    50051,
				DatabaseHost:      "db.example.com",
				DatabasePort:      5432,
				DatabaseUser:      "user",
				DatabasePassword:  "password",
				DatabaseName:      "mydb",
				
				
				TwilioAccountSID:  "ACxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
				TwilioAuthToken:   "your_auth_token",
				TwilioFormNumber:  "+15017122661",
				SendGridApiKey:    "SG.xxxxxxxxxxxxxxxxxxxxxxxxxxxx",
				SendGridFromEmail: "test@example.com",
				SendGridFromName:  "Test User",
				SendWhatsAppMessageSID:        "ACxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
				SendWhatsAppMessageFromNumber: "+15017122661",
				PushNotificationAccountCreds: "creds",
				PushNotificationProjectID:    "project-id",
			},
			expectedField: "SendWhatsAppMessageToken",
		},
		{
			name: "Missing SendWhatsAppMessageFromNumber",
			config: dto.ConfigResponse{
				HTTPListenAddress: "localhost",
				HTTPListenPort:    8080,
				GRPCListenAddress: "localhost",
				GRPCListenPort:    50051,
				DatabaseHost:      "db.example.com",
				DatabasePort:      5432,
				DatabaseUser:      "user",
				DatabasePassword:  "password",
				DatabaseName:      "mydb",
				
				
				TwilioAccountSID:  "ACxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
				TwilioAuthToken:   "your_auth_token",
				TwilioFormNumber:  "+15017122661",
				SendGridApiKey:    "SG.xxxxxxxxxxxxxxxxxxxxxxxxxxxx",
				SendGridFromEmail: "test@example.com",
				SendGridFromName:  "Test User",
				SendWhatsAppMessageSID:        "ACxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
				SendWhatsAppMessageToken:      "your_auth_token",
				PushNotificationAccountCreds: "creds",
				PushNotificationProjectID:    "project-id",
			},
			expectedField: "SendWhatsAppMessageFromNumber",
		},
		{
			name: "Missing PushNotificationAccountCreds",
			config: dto.ConfigResponse{
				HTTPListenAddress: "localhost",
				HTTPListenPort:    8080,
				GRPCListenAddress: "localhost",
				GRPCListenPort:    50051,
				DatabaseHost:      "db.example.com",
				DatabasePort:      5432,
				DatabaseUser:      "user",
				DatabasePassword:  "password",
				DatabaseName:      "mydb",
				
				
				TwilioAccountSID:  "ACxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
				TwilioAuthToken:   "your_auth_token",
				TwilioFormNumber:  "+15017122661",
				SendGridApiKey:    "SG.xxxxxxxxxxxxxxxxxxxxxxxxxxxx",
				SendGridFromEmail: "test@example.com",
				SendGridFromName:  "Test User",
				SendWhatsAppMessageSID:        "ACxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
				SendWhatsAppMessageToken:      "your_auth_token",
				SendWhatsAppMessageFromNumber: "+15017122661",
				PushNotificationProjectID:    "project-id",
			},
			expectedField: "PushNotificationAccountCreds",
		},
		{
			name: "Missing PushNotificationProjectID",
			config: dto.ConfigResponse{
				HTTPListenAddress: "localhost",
				HTTPListenPort:    8080,
				GRPCListenAddress: "localhost",
				GRPCListenPort:    50051,
				DatabaseHost:      "db.example.com",
				DatabasePort:      5432,
				DatabaseUser:      "user",
				DatabasePassword:  "password",
				DatabaseName:      "mydb",
				
				
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
			},
			expectedField: "PushNotificationProjectID",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := cfgService.ValidateConfig(tc.config)

			assert.Error(t, err)
			validationErrors, ok := err.(validator.ValidationErrors)
			assert.True(t, ok)
			assert.Len(t, validationErrors, 1) // Expecting one error for the missing field
			assert.Equal(t, "required", validationErrors[0].Tag())
			assert.Equal(t, tc.expectedField, validationErrors[0].Field())
		})
	}
}
