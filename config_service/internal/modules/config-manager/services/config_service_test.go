package services

import (
	"context"
	"errors"
	"nps-config-service/internal/modules/config-manager/apis/dtos"
	"nps-config-service/internal/modules/config-manager/models"
	repoMock "nps-config-service/internal/modules/config-manager/repositories/mocks"
	serviceMock "nps-config-service/internal/modules/config-manager/services/mocks"
	"nps-config-service/pkg/observability"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	clientMocks "go.temporal.io/sdk/mocks"
)

// Update mock interfaces to include context
type MockRepositoryWithContext struct {
	*repoMock.MockRepository
}

func (m *MockRepositoryWithContext) CreateAdmin(ctx context.Context, admin *models.Admin) (*models.Admin, error) {
	return m.MockRepository.CreateAdmin(ctx, admin)
}

type MockWebhookServiceWithContext struct {
	*serviceMock.MockWebhookService
}

func (m *MockWebhookServiceWithContext) DeleteAllWebhooks(ctx context.Context, env, serviceName string) error {
	return m.MockWebhookService.DeleteAllWebhooks(ctx, env, serviceName)
}

func setupConfigTestService(t *testing.T) (IConfigService, *repoMock.MockRepository, *serviceMock.MockWebhookService, *clientMocks.Client, *observability.ObservabilityStack) {
	mockRepo := new(repoMock.MockRepository)
	mockWebhook := new(serviceMock.MockWebhookService)
	mockTemporalClient := new(clientMocks.Client)
	obsStack := observability.NewObservabilityStack("config-service-test")
	service := NewConfigService(mockRepo, mockWebhook, mockTemporalClient, obsStack)
	return service, mockRepo, mockWebhook, mockTemporalClient, obsStack
}

func TestStoreConfigService(t *testing.T) {
	tests := []struct {
		name           string
		env            string
		serviceName    string
		configData     map[string]interface{}
		mockSetup      func(*repoMock.MockRepository, *serviceMock.MockWebhookService, *clientMocks.Client)
		expectedResult *dtos.SuccessResponse
		expectedError  *dtos.ServiceErrorResponse
	}{
		{
			name:        "successful store with webhooks",
			env:         "dev",
			serviceName: "test-service",
			configData:  map[string]interface{}{"key": "value"},
			mockSetup: func(repo *repoMock.MockRepository, webhook *serviceMock.MockWebhookService, temporal *clientMocks.Client) {
				// Mock successful config storage
				repo.On("StoreConfig", mock.Anything, "test-service", "dev", map[string]interface{}{"key": "value"}).
					Return(map[string]interface{}{"status": "success"}, nil)

				// Mock webhook retrieval
				webhookData := `[{"url":"http://example.com/webhook","method":"POST","environment":"dev","serviceName":"test-service"}]`
				repo.On("GetEtcdKey", mock.Anything, "/webhooks/dev/test-service").
					Return(webhookData, nil)

				// Mock temporal workflow execution
				temporal.On("ExecuteWorkflow", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(&clientMocks.WorkflowRun{}, nil)
			},
			expectedResult: &dtos.SuccessResponse{
				StatusCode: 201,
				Message:    "Config stored successfully",
				Data:       map[string]interface{}{"status": "success"},
			},
			expectedError: nil,
		},
		{
			name:        "successful store without webhooks",
			env:         "dev",
			serviceName: "test-service",
			configData:  map[string]interface{}{"key": "value"},
			mockSetup: func(repo *repoMock.MockRepository, webhook *serviceMock.MockWebhookService, temporal *clientMocks.Client) {
				// Mock successful config storage
				repo.On("StoreConfig", mock.Anything, "test-service", "dev", map[string]interface{}{"key": "value"}).
					Return(map[string]interface{}{"status": "success"}, nil)

				// Mock no webhooks found
				repo.On("GetEtcdKey", mock.Anything, "/webhooks/dev/test-service").
					Return("", errors.New("not found"))
			},
			expectedResult: &dtos.SuccessResponse{
				StatusCode: 201,
				Message:    "Config stored successfully: NOTE - No webhooks found",
				Data:       map[string]interface{}{"status": "success"},
			},
			expectedError: nil,
		},
		{
			name:        "store error",
			env:         "dev",
			serviceName: "test-service",
			configData:  map[string]interface{}{"key": "value"},
			mockSetup: func(repo *repoMock.MockRepository, webhook *serviceMock.MockWebhookService, temporal *clientMocks.Client) {
				repo.On("StoreConfig", mock.Anything, "test-service", "dev", map[string]interface{}{"key": "value"}).
					Return(nil, errors.New("store failed"))
			},
			expectedResult: nil,
			expectedError: &dtos.ServiceErrorResponse{
				StatusCode:   500,
				ErrorCode:    "Failed to store config",
				ErrorMessage: "store failed",
			},
		},
		{
			name:        "invalid webhook data",
			env:         "dev",
			serviceName: "test-service",
			configData:  map[string]interface{}{"key": "value"},
			mockSetup: func(repo *repoMock.MockRepository, webhook *serviceMock.MockWebhookService, temporal *clientMocks.Client) {
				// Mock successful config storage
				repo.On("StoreConfig", mock.Anything, "test-service", "dev", map[string]interface{}{"key": "value"}).
					Return(map[string]interface{}{"status": "success"}, nil)

				// Mock invalid webhook data
				repo.On("GetEtcdKey", mock.Anything, "/webhooks/dev/test-service").
					Return(`[{"url":"http://example.com/webhook"`, nil) // Malformed JSON
			},
			expectedResult: nil,
			expectedError: &dtos.ServiceErrorResponse{
				StatusCode:   500,
				ErrorCode:    "Failed to parse webhook data",
				ErrorMessage: "unexpected end of JSON input",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, mockRepo, mockWebhook, mockTemporalClient, _ := setupConfigTestService(t)
			tt.mockSetup(mockRepo, mockWebhook, mockTemporalClient)

			ctx := context.Background()
			result, err := service.StoreConfigService(ctx, tt.env, tt.serviceName, tt.configData)

			if tt.expectedError != nil {
				assert.NotNil(t, err)
				assert.Equal(t, tt.expectedError.StatusCode, err.StatusCode)
				assert.Equal(t, tt.expectedError.ErrorCode, err.ErrorCode)
				assert.Contains(t, err.ErrorMessage, tt.expectedError.ErrorMessage)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, tt.expectedResult.StatusCode, result.StatusCode)
				assert.Equal(t, tt.expectedResult.Message, result.Message)
				assert.Equal(t, tt.expectedResult.Data, result.Data)
			}

			mockRepo.AssertExpectations(t)
			mockWebhook.AssertExpectations(t)
			mockTemporalClient.AssertExpectations(t)
		})
	}
}

func TestGetConfigService(t *testing.T) {
	tests := []struct {
		name           string
		serviceName    string
		env            string
		mockSetup      func(*repoMock.MockRepository)
		expectedResult interface{}
		expectedError  error
	}{
		{
			name:        "successful get",
			serviceName: "test-service",
			env:         "dev",
			mockSetup: func(repo *repoMock.MockRepository) {
				expectedConfig := map[string]interface{}{"key": "value"}
				repo.On("GetConfig", mock.Anything, "test-service", "dev").
					Return(expectedConfig, nil)
			},
			expectedResult: map[string]interface{}{"key": "value"},
			expectedError:  nil,
		},
		{
			name:        "config not found",
			serviceName: "test-service",
			env:         "dev",
			mockSetup: func(repo *repoMock.MockRepository) {
				repo.On("GetConfig", mock.Anything, "test-service", "dev").
					Return(nil, nil)
			},
			expectedResult: nil,
			expectedError:  nil,
		},
		{
			name:        "repository error",
			serviceName: "test-service",
			env:         "dev",
			mockSetup: func(repo *repoMock.MockRepository) {
				repo.On("GetConfig", mock.Anything, "test-service", "dev").
					Return(nil, errors.New("get failed"))
			},
			expectedResult: nil,
			expectedError:  errors.New("failed to get config: get failed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, mockRepo, _, _, _ := setupConfigTestService(t)
			tt.mockSetup(mockRepo)

			ctx := context.Background()
			result, err := service.GetConfigService(ctx, tt.serviceName, tt.env)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult, result)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestGetConfigValueService(t *testing.T) {
	tests := []struct {
		name           string
		serviceName    string
		env            string
		key            string
		mockSetup      func(*repoMock.MockRepository)
		expectedResult interface{}
		expectedError  error
	}{
		{
			name:        "successful get value",
			serviceName: "test-service",
			env:         "dev",
			key:         "test-key",
			mockSetup: func(repo *repoMock.MockRepository) {
				repo.On("GetConfigValue", mock.Anything, "test-service", "dev", "test-key").
					Return("test-value", nil)
			},
			expectedResult: "test-value",
			expectedError:  nil,
		},
		{
			name:        "value not found",
			serviceName: "test-service",
			env:         "dev",
			key:         "test-key",
			mockSetup: func(repo *repoMock.MockRepository) {
				repo.On("GetConfigValue", mock.Anything, "test-service", "dev", "test-key").
					Return(nil, nil)
			},
			expectedResult: nil,
			expectedError:  nil,
		},
		{
			name:        "repository error",
			serviceName: "test-service",
			env:         "dev",
			key:         "test-key",
			mockSetup: func(repo *repoMock.MockRepository) {
				repo.On("GetConfigValue", mock.Anything, "test-service", "dev", "test-key").
					Return(nil, errors.New("get failed"))
			},
			expectedResult: nil,
			expectedError:  errors.New("failed to get config value: get failed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, mockRepo, _, _, _ := setupConfigTestService(t)
			tt.mockSetup(mockRepo)

			ctx := context.Background()
			result, err := service.GetConfigValueService(ctx, tt.serviceName, tt.env, tt.key)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult, result)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
