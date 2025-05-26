package services_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"thirdparty_service/internal/config"
	"thirdparty_service/internal/modules/config/dto"
	cacheclient "thirdparty_service/internal/modules/execute/clients/cache_client"
	cache "thirdparty_service/internal/modules/execute/services"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockRedisClient is a mock implementation of the cacheclient.RedisClient interface
type MockRedisClient struct {
	mock.Mock
}

func (m *MockRedisClient) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockRedisClient) SetCache(ctx context.Context, namespace, key, value string, ttl time.Duration, trackingID string) error {
	args := m.Called(ctx, namespace, key, value, ttl, trackingID)
	return args.Error(0)
}

func (m *MockRedisClient) GetCache(ctx context.Context, namespace, key, trackingID string) (string, bool, error) {
	args := m.Called(ctx, namespace, key, trackingID)
	return args.String(0), args.Bool(1), args.Error(2)
}

func (m *MockRedisClient) InvalidateCache(ctx context.Context, namespace, key, trackingID string) error {
	args := m.Called(ctx, namespace, key, trackingID)
	return args.Error(0)
}

func TestNewCacheManager(t *testing.T) {
	mockClient := new(MockRedisClient)

	tests := []struct {
		name          string
		cacheClient   cacheclient.RedisClient
		expectedError error
	}{
		{
			name:          "Success",
			cacheClient:   mockClient,
			expectedError: nil,
		},
		{
			name:          "Nil Cache Client",
			cacheClient:   nil,
			expectedError: errors.New("cacheClient cannot be zero in NewCacheManager"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager, err := cache.NewCacheManager(tt.cacheClient)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
				assert.Nil(t, manager)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, manager)
			}
		})
	}
}

func TestSetDataToCache(t *testing.T) {
	mockClient := new(MockRedisClient)
	service, _ := cache.NewCacheManager(mockClient)
	ctx := context.Background()
	key := "test-key"
	cfg := &dto.ConfigResponse{
		HTTPListenAddress: "localhost",
		HTTPListenPort:    8080,
		GRPCListenAddress: "localhost",
		GRPCListenPort:    8081,
		DatabaseHost:      "db.example.com",
		DatabasePort:      5432,
		DatabaseUser:      "user",
		DatabasePassword:  "password",
		DatabaseName:      "mydatabase",
		TwilioAccountSID:  "sid",
		TwilioAuthToken:   "token",
		TwilioFormNumber:  "+15005550006",
		SendGridApiKey:    "sendgridkey",
		SendGridFromEmail: "from@example.com",
		SendGridFromName:  "From Name",
		SendWhatsAppMessageSID: "whatsapp_sid",
		SendWhatsAppMessageToken: "whatsapp_token",
		SendWhatsAppMessageFromNumber: "+15005550007",
		PushNotificationAccountCreds: "push_creds",
		PushNotificationProjectID: "push_project_id",
	}
	cfgJSON, _ := json.Marshal(cfg)
	cfgString := string(cfgJSON)

	// Mock config.AppConfig.ServiceName
	originalServiceName := config.AppConfig.ServiceName
	config.AppConfig.ServiceName = "test-service"
	defer func() { config.AppConfig.ServiceName = originalServiceName }()

	tests := []struct {
		name          string
		service       *cache.CacheManagerService
		mockSetup     func()
		expectedError error
	}{
		{
			name:    "Success",
			service: service,
			mockSetup: func() {
				mockClient.On("SetCache",
					mock.Anything, // context
					config.AppConfig.ServiceName,
					key,
					cfgString,
					mock.AnythingOfType("time.Duration"), // ttl
					mock.AnythingOfType("string"),        // trackingID
				).Return(nil).Once()
			},
			expectedError: nil,
		},
		{
			name:    "SetCache Error",
			service: service,
			mockSetup: func() {
				mockClient.On("SetCache",
					mock.Anything, // context
					config.AppConfig.ServiceName,
					key,
					cfgString,
					mock.AnythingOfType("time.Duration"), // ttl
					mock.AnythingOfType("string"),        // trackingID
				).Return(errors.New("mock set cache error")).Once()
			},
			expectedError: errors.New("mock set cache error"),
		},
		// Note: Marshaling error is hard to test directly here as json.Marshal is used directly.
		// If json.Marshal was abstracted, we could mock it.
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.mockSetup != nil {
				tt.mockSetup()
			}

			err := tt.service.SetDataToCache(ctx, key, cfg)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}
			mockClient.AssertExpectations(t)
		})
	}
}

func TestGetDataToCache(t *testing.T) {
	mockClient := new(MockRedisClient)
	service, _ := cache.NewCacheManager(mockClient)
	ctx := context.Background()
	key := "test-key"
	cfg := &dto.ConfigResponse{
		HTTPListenAddress: "localhost",
		HTTPListenPort:    8080,
		GRPCListenAddress: "localhost",
		GRPCListenPort:    8081,
		DatabaseHost:      "db.example.com",
		DatabasePort:      5432,
		DatabaseUser:      "user",
		DatabasePassword:  "password",
		DatabaseName:      "mydatabase",
		TwilioAccountSID:  "sid",
		TwilioAuthToken:   "token",
		TwilioFormNumber:  "+15005550006",
		SendGridApiKey:    "sendgridkey",
		SendGridFromEmail: "from@example.com",
		SendGridFromName:  "From Name",
		SendWhatsAppMessageSID: "whatsapp_sid",
		SendWhatsAppMessageToken: "whatsapp_token",
		SendWhatsAppMessageFromNumber: "+15005550007",
		PushNotificationAccountCreds: "push_creds",
		PushNotificationProjectID: "push_project_id",
	}
	cfgJSON, _ := json.Marshal(cfg)
	cfgString := string(cfgJSON)

	// Mock config.AppConfig.ServiceName
	originalServiceName := config.AppConfig.ServiceName
	config.AppConfig.ServiceName = "test-service"
	defer func() { config.AppConfig.ServiceName = originalServiceName }()

	tests := []struct {
		name          string
		service       *cache.CacheManagerService
		mockSetup     func()
		expectedCfg   *dto.ConfigResponse
		expectedError error
	}{
		{
			name:    "Success",
			service: service,
			mockSetup: func() {
				mockClient.On("GetCache",
					mock.Anything, // context
					config.AppConfig.ServiceName,
					key,
					mock.AnythingOfType("string"), // trackingID
				).Return(cfgString, true, nil).Once()
			},
			expectedCfg:   cfg,
			expectedError: nil,
		},
		{
			name:    "GetCache Error",
			service: service,
			mockSetup: func() {
				mockClient.On("GetCache",
					mock.Anything, // context
					config.AppConfig.ServiceName,
					key,
					mock.AnythingOfType("string"), // trackingID
				).Return("", false, errors.New("mock get cache error")).Once()
			},
			expectedCfg:   nil,
			expectedError: errors.New("failed to get cache: mock get cache error"),
		},
		{
			name:    "Not Found",
			service: service,
			mockSetup: func() {
				mockClient.On("GetCache",
					mock.Anything, // context
					config.AppConfig.ServiceName,
					key,
					mock.AnythingOfType("string"), // trackingID
				).Return("", false, nil).Once()
			},
			expectedCfg:   nil,
			expectedError: errors.New("config not found in cache"),
		},
		{
			name:    "Unmarshal Error",
			service: service,
			mockSetup: func() {
				mockClient.On("GetCache",
					mock.Anything, // context
					config.AppConfig.ServiceName,
					key,
					mock.AnythingOfType("string"), // trackingID
				).Return("invalid json", true, nil).Once()
			},
			expectedCfg:   nil,
			expectedError: errors.New("failed to unmarshal cached config"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.mockSetup != nil {
				tt.mockSetup()
			}

			resultCfg, err := tt.service.GetDataToCache(ctx, key)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
				assert.Nil(t, resultCfg)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedCfg, resultCfg)
			}
			mockClient.AssertExpectations(t)
		})
	}
}