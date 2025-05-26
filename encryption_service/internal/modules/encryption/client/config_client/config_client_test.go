package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"encryption_microservice/internal/config"
	httpclient "encryption_microservice/pkg/http"
)

// MockHTTPClient is a mock implementation of HTTPClient interface
type MockHTTPClient struct {
	doFunc func(ctx context.Context, req httpclient.Request) (*http.Response, error)
}

// Do mocks the Do method
func (m *MockHTTPClient) Do(ctx context.Context, req httpclient.Request) (*http.Response, error) {
	return m.doFunc(ctx, req)
}

// TestNewConfigClient tests the NewConfigClient function
func TestNewConfigClient(t *testing.T) {
	validEnv := &config.EnvConfig{

		ConfigServiceUrl:   "http://config-service",
		ConfigServiceToken: "test-token",
		Environment:        "test",
	}

	mockHTTPClient := &MockHTTPClient{}

	tests := []struct {
		name          string
		httpClient    httpclient.HTTPClient
		env           *config.EnvConfig
		expectedError bool
	}{
		{
			name:          "valid parameters",
			httpClient:    mockHTTPClient,
			env:           validEnv,
			expectedError: false,
		},
		{
			name:          "nil HTTP client",
			httpClient:    nil,
			env:           validEnv,
			expectedError: true,
		},
		{
			name:          "nil environment",
			httpClient:    mockHTTPClient,
			env:           nil,
			expectedError: true,
		},
		{
			name:       "missing config service URL",
			httpClient: mockHTTPClient,
			env: &config.EnvConfig{

				ConfigServiceToken: "test-token",
				Environment:        "test",
				// ConfigServiceUrl is missing
			},
			expectedError: true,
		},
		{
			name:       "missing config service token",
			httpClient: mockHTTPClient,
			env: &config.EnvConfig{
				ConfigServiceUrl: "http://config-service",
				Environment:      "test",
				// ConfigServiceToken is missing
			},
			expectedError: true,
		},
		{
			name:       "missing environment",
			httpClient: mockHTTPClient,
			env: &config.EnvConfig{

				ConfigServiceUrl:   "http://config-service",
				ConfigServiceToken: "test-token",
				// Environment is missing
			},
			expectedError: true,
		},
		{
			name:       "missing service name",
			httpClient: mockHTTPClient,
			env: &config.EnvConfig{
				ConfigServiceUrl:   "http://config-service",
				ConfigServiceToken: "test-token",
				Environment:        "test",
				// ServiceName is missing
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewConfigClient(tt.httpClient, tt.env)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, client)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, client)
			}
		})
	}
}

// TestFetchConfig tests the FetchConfig method
func TestFetchConfig(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Valid config response
	validConfig := config.Config{
		// ServerPort:       "8082",
		// GrpcPort:         "50051",
		// UserServiceURL:   "http://host.docker.internal:8080",
		VaultAddr:                 "https://host.docker.internal:8200",
		VaultToken:                "test-token",
		VaultPath:                 "transit",
		OtelCollectorGrpcEndpoint: "localhost:4317",
	}

	// Define response structure for test
	type ConfigServiceResponse struct {
		Success bool          `json:"success"`
		Message string        `json:"message"`
		Data    config.Config `json:"data"`
	}

	validResponse := ConfigServiceResponse{
		Success: true,
		Message: "OK",
		Data:    validConfig,
	}

	validResponseBytes, _ := json.Marshal(validResponse)

	// Empty config response (invalid)
	emptyConfig := config.Config{
		// VaultAddr is missing, which makes this config invalid
		// ServerPort:     "8082",
		// GrpcPort:       "50051",
		// UserServiceURL: "http://localhost:8080",
	}

	emptyResponse := ConfigServiceResponse{
		Success: true,
		Message: "OK",
		Data:    emptyConfig,
	}

	emptyResponseBytes, _ := json.Marshal(emptyResponse)

	// Common test environment
	env := &config.EnvConfig{

		ConfigServiceUrl:   "http://config-service",
		ConfigServiceToken: "test-token",
		Environment:        "test",
	}

	tests := []struct {
		name           string
		setupMockHTTP  func() *MockHTTPClient
		expectedConfig *config.Config
		expectedError  bool
	}{
		{
			name: "successful config fetch",
			setupMockHTTP: func() *MockHTTPClient {
				return &MockHTTPClient{
					doFunc: func(ctx context.Context, req httpclient.Request) (*http.Response, error) {
						// Verify request
						assert.Equal(t, httpclient.GET, req.Method)
						assert.Equal(t, "http://config-service/config/test/encryption_service", req.URL)
						assert.Equal(t, "Bearer test-token", req.Headers["Authorization"])

						// Return successful response
						return &http.Response{
							StatusCode: http.StatusOK,
							Body:       io.NopCloser(bytes.NewReader(validResponseBytes)),
						}, nil
					},
				}
			},
			expectedConfig: &validConfig,
			expectedError:  false,
		},
		{
			name: "HTTP client error",
			setupMockHTTP: func() *MockHTTPClient {
				return &MockHTTPClient{
					doFunc: func(ctx context.Context, req httpclient.Request) (*http.Response, error) {
						return nil, errors.New("HTTP error")
					},
				}
			},
			expectedConfig: nil,
			expectedError:  true,
		},
		{
			name: "non-OK status code",
			setupMockHTTP: func() *MockHTTPClient {
				return &MockHTTPClient{
					doFunc: func(ctx context.Context, req httpclient.Request) (*http.Response, error) {
						return &http.Response{
							StatusCode: http.StatusInternalServerError,
							Body:       io.NopCloser(bytes.NewReader([]byte(`{"message": "Internal server error"}`))),
						}, nil
					},
				}
			},
			expectedConfig: nil,
			expectedError:  true,
		},
		{
			name: "invalid JSON response",
			setupMockHTTP: func() *MockHTTPClient {
				return &MockHTTPClient{
					doFunc: func(ctx context.Context, req httpclient.Request) (*http.Response, error) {
						return &http.Response{
							StatusCode: http.StatusOK,
							Body:       io.NopCloser(bytes.NewReader([]byte(`{"malformed json`))),
						}, nil
					},
				}
			},
			expectedConfig: nil,
			expectedError:  true,
		},
		{
			name: "missing required fields in config",
			setupMockHTTP: func() *MockHTTPClient {
				return &MockHTTPClient{
					doFunc: func(ctx context.Context, req httpclient.Request) (*http.Response, error) {
						return &http.Response{
							StatusCode: http.StatusOK,
							Body:       io.NopCloser(bytes.NewReader(emptyResponseBytes)),
						}, nil
					},
				}
			},
			expectedConfig: nil,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockHTTP := tt.setupMockHTTP()

			// Create the client
			client, _ := NewConfigClient(mockHTTP, env)

			// Call the method
			config, err := client.FetchConfig(context.Background())

			// Verify results
			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, config)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, config)
				assert.Equal(t, tt.expectedConfig.VaultAddr, config.VaultAddr)
				assert.Equal(t, tt.expectedConfig.VaultToken, config.VaultToken)
				assert.Equal(t, tt.expectedConfig.VaultPath, config.VaultPath)
			}
		})
	}
}
