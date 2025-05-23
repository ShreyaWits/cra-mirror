package services

import (
	"context"
	"nps-config-service/internal/modules/config-manager/apis/dtos"
	"nps-config-service/internal/modules/config-manager/models"
	repoMock "nps-config-service/internal/modules/config-manager/repositories/mocks"
	"nps-config-service/pkg/observability"
	"testing"

	"github.com/golang-jwt/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// setupAdminTestService creates a new test service with mock repository and observability stack
func setupAdminTestService(t *testing.T) (IAdminService, *repoMock.MockRepository, *observability.ObservabilityStack) {
	mockRepo := new(repoMock.MockRepository)
	obsStack := observability.NewObservabilityStack("admin-service-test")
	service := NewAdminService(mockRepo, obsStack)
	return service, mockRepo, obsStack
}

func TestCreateAdminService(t *testing.T) {
	tests := []struct {
		name           string
		request        *dtos.AdminSignupDto
		mockSetup      func(*repoMock.MockRepository)
		expectedResult *dtos.ResponseAdminSignupDto
		expectedError  bool
	}{
		{
			name: "successful creation",
			request: &dtos.AdminSignupDto{
				Username: "testuser",
				Password: "testpass",
				Secret:   "validsecret",
			},
			mockSetup: func(m *repoMock.MockRepository) {
				m.On("CreateAdmin", mock.Anything, mock.AnythingOfType("*models.Admin")).Return(&models.Admin{
					UserName: "testuser",
					Password: "testpass",
				}, nil)
			},
			expectedResult: &dtos.ResponseAdminSignupDto{
				Success: true,
				Message: "Admin created successfully",
				AdminId: "testuser",
			},
			expectedError: false,
		},
		{
			name: "repository error",
			request: &dtos.AdminSignupDto{
				Username: "testuser",
				Password: "testpass",
				Secret:   "validsecret",
			},
			mockSetup: func(m *repoMock.MockRepository) {
				m.On("CreateAdmin", mock.Anything, mock.AnythingOfType("*models.Admin")).Return(nil, assert.AnError)
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, mockRepo, _ := setupAdminTestService(t)
			tt.mockSetup(mockRepo)

			ctx := context.Background()
			result, err := service.CreateAdminService(ctx, tt.request)

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
	tests := []struct {
		name          string
		request       *dtos.AdminLoginDto
		secret        string
		mockSetup     func(*repoMock.MockRepository)
		expectedError bool
		validateToken bool
	}{
		{
			name: "successful fetch",
			request: &dtos.AdminLoginDto{
				Username: "testuser",
				Password: "testpass",
			},
			secret: "testsecret",
			mockSetup: func(m *repoMock.MockRepository) {
				m.On("GetAdminByCredentials", mock.Anything, "testuser", "testpass").Return(&models.Admin{
					UserName: "testuser",
					Password: "testpass",
				}, nil)
			},
			expectedError: false,
			validateToken: true,
		},
		{
			name: "invalid credentials",
			request: &dtos.AdminLoginDto{
				Username: "testuser",
				Password: "wrongpass",
			},
			secret: "testsecret",
			mockSetup: func(m *repoMock.MockRepository) {
				m.On("GetAdminByCredentials", mock.Anything, "testuser", "wrongpass").Return(nil, assert.AnError)
			},
			expectedError: true,
			validateToken: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, mockRepo, _ := setupAdminTestService(t)
			tt.mockSetup(mockRepo)

			ctx := context.Background()
			result, err := service.FetchAdminService(ctx, tt.request, tt.secret)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.True(t, result.Success)
				assert.NotEmpty(t, result.Token)
				assert.NotEmpty(t, result.RefreshToken)

				if tt.validateToken {
					// Verify token
					token, err := jwt.Parse(result.Token, func(token *jwt.Token) (interface{}, error) {
						return []byte(tt.secret), nil
					})
					assert.NoError(t, err)
					assert.True(t, token.Valid)
				}
			}
			mockRepo.AssertExpectations(t)
		})
	}
}
