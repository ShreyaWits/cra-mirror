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

	"messaging_service/internal/config"
	"messaging_service/internal/modules/message_broker/models"
	httpclient "messaging_service/pkg/http"
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
	validEnv := &config.Env{

		ConfigServiceUrl:   "http://config-service",
		ConfigServiceToken: "test-token",
		Environment:        "test",
	}

	mockHTTPClient := &MockHTTPClient{}

	tests := []struct {
		name          string
		httpClient    httpclient.HTTPClient
		env           *config.Env
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
			env: &config.Env{

				ConfigServiceToken: "test-token",
				Environment:        "test",
				// ConfigServiceUrl is missing
			},
			expectedError: true,
		},
		{
			name:       "missing config service token",
			httpClient: mockHTTPClient,
			env: &config.Env{
				ConfigServiceUrl: "http://config-service",
				Environment:      "test",
				// ConfigServiceToken is missing
			},
			expectedError: true,
		},
		{
			name:       "missing environment",
			httpClient: mockHTTPClient,
			env: &config.Env{

				ConfigServiceUrl:   "http://config-service",
				ConfigServiceToken: "test-token",
				// Environment is missing
			},
			expectedError: true,
		},
		{
			name:       "missing service name",
			httpClient: mockHTTPClient,
			env: &config.Env{
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
	validConfig := models.MessaggingConfigResponse{
		KafkaBrokers:           []string{"localhost:9092"},
		KafkaNumPartitions:     3,
		KafkaReplicationFactor: 3,
		KafkaBatchSize:         100,
		KafkaBatchBytes:        1048576,
		KafkaBatchTimeoutMs:    500,
		KafkaCompressionCodec:  "snappy",
		KafkaMaxAttempts:       3,
		KafkaRetryBackoffMs:    100,
		KafkaReadTimeoutMs:     5000,
		KafkaWriteTimeoutMs:    5000,
		KafkaRetentionMs:       3000,

		KafkaConsumerMaxWaitMs:        5000,
		KafkaConsumerCommitIntervalMs: 5000,
		KafkaConsumerSessionTimeoutMs: 30000,
		KafkaConsumerHeartbeatMs:      1000,
		KafkaConsumerMaxPollRecords:   1000,
		KafkaConsumerAutoOffsetReset:  "earliest",
		KafkaIsolationLevel:           "read_committed",
	}

	validResponse := models.ConfigServiceResponse{
		StatusCode: 200,
		Message:    "OK",
		Data:       validConfig,
	}

	validResponseBytes, _ := json.Marshal(validResponse)

	// Empty config response (invalid)
	emptyConfig := models.MessaggingConfigResponse{
		// KafkaBrokers is missing, which makes this config invalid
	}

	emptyResponse := models.ConfigServiceResponse{
		StatusCode: 200,
		Message:    "OK",
		Data:       emptyConfig,
	}

	emptyResponseBytes, _ := json.Marshal(emptyResponse)

	// Common test environment
	env := &config.Env{

		ConfigServiceUrl:   "http://config-service",
		ConfigServiceToken: "test-token",
		Environment:        "test",
	}

	tests := []struct {
		name           string
		setupMockHTTP  func() *MockHTTPClient
		expectedConfig *models.MessaggingConfigResponse
		expectedError  bool
	}{
		{
			name: "successful config fetch",
			setupMockHTTP: func() *MockHTTPClient {
				return &MockHTTPClient{
					doFunc: func(ctx context.Context, req httpclient.Request) (*http.Response, error) {
						// Verify request
						assert.Equal(t, httpclient.GET, req.Method)
						assert.Equal(t, "http://config-service/config/test/test-service", req.URL)
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
				assert.Equal(t, tt.expectedConfig, config)
			}
		})
	}
}
