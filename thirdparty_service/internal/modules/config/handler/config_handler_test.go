package handler_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"thirdparty_service/internal/modules/config/dto"
	"thirdparty_service/internal/modules/config/handler"
	configservice "thirdparty_service/internal/modules/config/service"
)

// MockConfigService is a mock implementation of the ConfigService interface
type MockConfigService struct {
	mock.Mock
}

func (m *MockConfigService) LoginToConfigService() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

func (m *MockConfigService) FetchDynamicConfig(token string) (*dto.ConfigResponse, error) {
	args := m.Called(token)
	// Check if the first argument is nil before attempting to cast
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.ConfigResponse), args.Error(1)
}

func (m *MockConfigService) ValidateConfig(config dto.ConfigResponse) error {
	args := m.Called(config)
	return args.Error(0)
}

func (m *MockConfigService) GetCacheConfig() (*dto.ConfigResponse, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.ConfigResponse), args.Error(1)
}

// Ensure MockConfigService implements the ConfigService interface
var _ configservice.ConfigService = (*MockConfigService)(nil)

func TestGetDynamicConfig(t *testing.T) {
	tests := []struct {
		name              string
		mockLoginToken    string
		mockLoginErr      error
		mockFetchConfig   *dto.ConfigResponse
		mockFetchErr      error
		mockValidateErr   error
		expectedConfig    *dto.ConfigResponse
		expectedErr       error
	}{
		{
			name:              "Success",
			mockLoginToken:    "test_token",
			mockLoginErr:      nil,
			mockFetchConfig:   &dto.ConfigResponse{HTTPListenAddress: "localhost", HTTPListenPort: 8080},
			mockFetchErr:      nil,
			mockValidateErr:   nil,
			expectedConfig:    &dto.ConfigResponse{HTTPListenAddress: "localhost", HTTPListenPort: 8080},
			expectedErr:       nil,
		},
		{
			name:              "LoginError",
			mockLoginToken:    "",
			mockLoginErr:      errors.New("login failed"),
			mockFetchConfig:   nil,
			mockFetchErr:      nil,
			mockValidateErr:   nil,
			expectedConfig:    nil,
			expectedErr:       errors.New("login failed"),
		},
		{
			name:              "FetchError",
			mockLoginToken:    "test_token",
			mockLoginErr:      nil,
			mockFetchConfig:   nil,
			mockFetchErr:      errors.New("fetch failed"),
			mockValidateErr:   nil,
			expectedConfig:    nil,
			expectedErr:       errors.New("fetch failed"),
		},
		{
			name:              "ValidationError",
			mockLoginToken:    "test_token",
			mockLoginErr:      nil,
			mockFetchConfig:   &dto.ConfigResponse{HTTPListenAddress: "localhost", HTTPListenPort: 8080},
			mockFetchErr:      nil,
			mockValidateErr:   errors.New("validation failed"),
			expectedConfig:    nil,
			expectedErr:       errors.New("validation failed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockConfigService)

			// Set up expectations for LoginToConfigService
			mockService.On("LoginToConfigService").Return(tt.mockLoginToken, tt.mockLoginErr)

			// Set up expectations for FetchDynamicConfig only if login is successful
			if tt.mockLoginErr == nil {
				mockService.On("FetchDynamicConfig", tt.mockLoginToken).Return(tt.mockFetchConfig, tt.mockFetchErr)
			}

			// Set up expectations for ValidateConfig only if login and fetch are successful and fetch returned a non-nil config
			if tt.mockLoginErr == nil && tt.mockFetchErr == nil && tt.mockFetchConfig != nil {
				mockService.On("ValidateConfig", *tt.mockFetchConfig).Return(tt.mockValidateErr)
			}

			configHandler := handler.NewConfigHandler(mockService)

			config, err := configHandler.GetDynamicConfig()

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.EqualError(t, err, tt.expectedErr.Error())
				assert.Nil(t, config)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedConfig, config)
			}

			mockService.AssertExpectations(t)
		})
	}
}
// func TestGetCacheConfig(t *testing.T) {
// 	tests := []struct {
// 		name           string
// 		mockConfig     *dto.ConfigResponse
// 		mockErr        error
// 		expectedConfig *dto.ConfigResponse
// 		expectedErr    error
// 	}{
// 		{
// 			name:           "Success",
// 			mockConfig:     &dto.ConfigResponse{HTTPListenAddress: "localhost", HTTPListenPort: 8080},
// 			mockErr:        nil,
// 			expectedConfig: &dto.ConfigResponse{HTTPListenAddress: "localhost", HTTPListenPort: 8080},
// 			expectedErr:    nil,
// 		},
// 		{
// 			name:           "Error",
// 			mockConfig:     nil,
// 			mockErr:        errors.New("cache fetch failed"),
// 			expectedConfig: nil,
// 			expectedErr:    errors.New("cache fetch failed"),
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			mockService := new(MockConfigService)

// 			// Set up expectations for GetCacheConfig
// 			mockService.On("GetCacheConfig").Return(tt.mockConfig, tt.mockErr)

// 			configHandler := handler.NewConfigHandler(mockService)

// 			config, err := configHandler.GetCacheConfig()

// 			if tt.expectedErr != nil {
// 				assert.Error(t, err)
// 				assert.EqualError(t, err, tt.expectedErr.Error())
// 				assert.Nil(t, config)
// 			} else {
// 				assert.NoError(t, err)
// 				assert.Equal(t, tt.expectedConfig, config)
// 			}

// 			mockService.AssertExpectations(t)
// 		})
// 	}
// }