package services

import (
	"context"
	"nps-config-service/internal/modules/config-manager/apis/dtos"
	"nps-config-service/internal/modules/config-manager/models"
	"testing"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockConfigRepo struct {
	mock.Mock
}

func (m *MockConfigRepo) CreateAdmin(admin *models.Admin) (*models.Admin, error) {
	args := m.Called(admin)
	return args.Get(0).(*models.Admin), args.Error(1)
}

func (m *MockConfigRepo) GetAdminByCredentials(username, password string) (*models.Admin, error) {
	args := m.Called(username, password)
	return args.Get(0).(*models.Admin), args.Error(1)
}

// Implement other required interface methods
func (m *MockConfigRepo) StoreConfig(serviceName, environment string, configData map[string]interface{}) (interface{}, error) {
	return nil, nil
}

func (m *MockConfigRepo) GetConfig(serviceName, environment string) (map[string]interface{}, error) {
	return nil, nil
}

func (m *MockConfigRepo) GetAllKeys(prefix string) (map[string]string, error) {
	return nil, nil
}

func (m *MockConfigRepo) GetConfigMetadata(serviceName, environment string) (*models.ConfigMetadata, error) {
	return nil, nil
}

func (m *MockConfigRepo) GetConfigValue(serviceName, environment, key string) (interface{}, error) {
	return nil, nil
}

func (m *MockConfigRepo) SetEtcdKey(ctx context.Context, key string, data string, ttl time.Duration) error {
	return nil
}

func (m *MockConfigRepo) GetEtcdKey(ctx context.Context, key string) (string, error) {
	return "", nil
}

func (m *MockConfigRepo) DeleteEtcdKey(ctx context.Context, key string) error {
	return nil
}

func TestCreateAdminService(t *testing.T) {
	// Setup
	mockRepo := new(MockConfigRepo)
	service := NewAdminService(mockRepo)

	// Test cases
	tests := []struct {
		name           string
		request        *dtos.AdminSignupDto
		mockAdmin      *models.Admin
		mockError      error
		expectedResult *dtos.ResponseAdminSignupDto
		expectedError  bool
	}{
		{
			name: "Successful admin creation",
			request: &dtos.AdminSignupDto{
				Username: "testadmin",
				Password: "testpass",
				Secret:   "validsecret",
			},
			mockAdmin: &models.Admin{
				UserName: "testadmin",
				Password: "testpass",
			},
			expectedResult: &dtos.ResponseAdminSignupDto{
				Success: true,
				Message: "Admin created successfully",
				AdminId: "testadmin",
			},
		},
		{
			name: "Admin already exists",
			request: &dtos.AdminSignupDto{
				Username: "existingadmin",
				Password: "testpass",
				Secret:   "validsecret",
			},
			mockAdmin: &models.Admin{
				UserName: "existingadmin",
				Password: "testpass",
			},
			mockError:     assert.AnError,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock expectations
			mockRepo.On("CreateAdmin", tt.mockAdmin).Return(tt.mockAdmin, tt.mockError)

			// Call service
			result, err := service.CreateAdminService(tt.request)

			// Assertions
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult, result)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestFetchAdminService(t *testing.T) {
	// Setup
	mockRepo := new(MockConfigRepo)
	service := NewAdminService(mockRepo)

	// Test cases
	tests := []struct {
		name          string
		request       *dtos.AdminLoginDto
		secret        string
		mockAdmin     *models.Admin
		mockError     error
		expectedError bool
	}{
		{
			name: "Successful admin login",
			request: &dtos.AdminLoginDto{
				Username: "testadmin",
				Password: "testpass",
			},
			secret: "testsecret",
			mockAdmin: &models.Admin{
				UserName: "testadmin",
				Password: "testpass",
			},
		},
		{
			name: "Invalid credentials",
			request: &dtos.AdminLoginDto{
				Username: "wrongadmin",
				Password: "wrongpass",
			},
			secret: "testsecret",
			mockAdmin: &models.Admin{
				UserName: "wrongadmin",
				Password: "wrongpass",
			},
			mockError:     assert.AnError,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock expectations
			mockRepo.On("GetAdminByCredentials", tt.request.Username, tt.request.Password).Return(tt.mockAdmin, tt.mockError)

			// Call service
			result, err := service.FetchAdminService(tt.request, tt.secret)

			// Assertions
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.True(t, result.Success)
				assert.NotEmpty(t, result.Token)
				assert.NotEmpty(t, result.RefreshToken)

				// Verify token
				token, err := jwt.Parse(result.Token, func(token *jwt.Token) (interface{}, error) {
					return []byte(tt.secret), nil
				})
				assert.NoError(t, err)
				assert.True(t, token.Valid)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
