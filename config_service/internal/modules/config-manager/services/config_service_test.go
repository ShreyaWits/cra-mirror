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
	return m.MockRepository.CreateAdmin(admin)
}

type MockWebhookServiceWithContext struct {
	*serviceMock.MockWebhookService
}

func (m *MockWebhookServiceWithContext) DeleteAllWebhooks(ctx context.Context, env, serviceName string) error {
	return m.MockWebhookService.DeleteAllWebhooks(env, serviceName)
}

func setupTestService(t *testing.T) (IConfigService, *MockRepositoryWithContext, *MockWebhookServiceWithContext, *clientMocks.Client, *observability.ObservabilityStack) {
	mockRepo := &MockRepositoryWithContext{MockRepository: new(repoMock.MockRepository)}
	mockWebhook := &MockWebhookServiceWithContext{MockWebhookService: new(serviceMock.MockWebhookService)}
	mockTemporalClient := new(clientMocks.Client)
	obsStack := observability.NewObservabilityStack("config-service-test")
	service := NewConfigService(mockRepo, mockWebhook, mockTemporalClient, obsStack)
	return service, mockRepo, mockWebhook, mockTemporalClient, obsStack
}

func TestStoreConfigService_Success(t *testing.T) {
	service, mockRepo, mockWebhook, mockTemporalClient, _ := setupTestService(t)
	ctx := context.Background()

	env := "dev"
	serviceName := "test-service"
	req := map[string]interface{}{"key": "value"}

	storedResponse := dtos.SuccessResponse{StatusCode: 201, Message: "Config stored successfully", Data: dtos.SuccessResponse{StatusCode: 201, Message: "Config stored successfully", Data: dtos.SuccessResponse{StatusCode: 201, Message: "Config stored successfully", Data: map[string]interface{}{"status": "success"}}}}
	webhookData := `[{"url":"http://example.com/webhook"}]`

	mockRepo.On("StoreConfig", serviceName, env, req).Return(storedResponse, nil)
	mockRepo.On("Get", mock.Anything, "/webhooks/dev/test-service").Return(webhookData, nil)

	mockWebhook.On("NotifyWebhook", mock.Anything, req)
	mockTemporalClient.On("ExecuteWorkflow", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&clientMocks.WorkflowRun{}, nil)
	result, err := service.StoreConfigService(ctx, env, serviceName, req)

	assert.Nil(t, err)
	assert.Equal(t, storedResponse.StatusCode, result.StatusCode)

	mockRepo.AssertExpectations(t)
}

func TestStoreConfigService_InvalidWebhookData(t *testing.T) {
	service, mockRepo, mockWebhook, _, _ := setupTestService(t)
	ctx := context.Background()

	env := "dev"
	serviceName := "test-service"
	req := map[string]interface{}{"key": "value"}

	storedResponse := map[string]interface{}{"status": "success"}

	// Invalid webhook data (malformed JSON)
	webhookData := `[{"url":"http://example.com/webhook"` // Missing closing bracket

	mockRepo.On("StoreConfig", serviceName, env, req).Return(storedResponse, nil)
	mockRepo.On("Get", mock.Anything, "/webhooks/dev/test-service").Return(webhookData, nil)

	result, err := service.StoreConfigService(ctx, env, serviceName, req)

	assert.NotNil(t, err)
	assert.Nil(t, result)

	mockRepo.AssertExpectations(t)
	mockWebhook.AssertExpectations(t)
}

func TestStoreConfigService_StoreError(t *testing.T) {
	service, mockRepo, _, _, _ := setupTestService(t)
	ctx := context.Background()

	env := "dev"
	serviceName := "test-service"
	req := map[string]interface{}{"key": "value"}

	mockRepo.On("StoreConfig", serviceName, env, req).Return(nil, errors.New("store failed"))
	result, err := service.StoreConfigService(ctx, env, serviceName, req)

	assert.Nil(t, result)
	assert.NotNil(t, err)

	mockRepo.AssertExpectations(t)
}

func TestStoreConfigService_NoWebhookFound(t *testing.T) {
	service, mockRepo, _, _, _ := setupTestService(t)
	ctx := context.Background()

	env := "dev"
	serviceName := "test-service"
	req := map[string]interface{}{"key": "value"}

	storedResponse := map[string]interface{}{"status": "success"}

	mockRepo.On("StoreConfig", serviceName, env, req).Return(storedResponse, nil)
	mockRepo.On("Get", mock.Anything, "/webhooks/dev/test-service").Return("", errors.New("not found"))

	result, err := service.StoreConfigService(ctx, env, serviceName, req)

	assert.Nil(t, result)
	assert.NotNil(t, err)

	mockRepo.AssertExpectations(t)
}

func TestGetConfigService_Success(t *testing.T) {
	service, mockRepo, _, _, _ := setupTestService(t)
	ctx := context.Background()

	serviceName := "test-service"
	env := "dev"
	expectedConfig := map[string]interface{}{"key": "value"}

	mockRepo.On("GetConfig", serviceName, env).Return(expectedConfig, nil)

	result, err := service.GetConfigService(ctx, serviceName, env)

	assert.NoError(t, err)
	assert.Equal(t, expectedConfig, result)

	mockRepo.AssertExpectations(t)
}

func TestGetConfigValueService_Success(t *testing.T) {
	service, mockRepo, _, _, _ := setupTestService(t)
	ctx := context.Background()

	serviceName := "test-service"
	env := "dev"
	key := "some-key"
	expectedValue := "some-value"

	mockRepo.On("GetConfigValue", serviceName, env, key).Return(expectedValue, nil)

	result, err := service.GetConfigValueService(ctx, serviceName, env, key)

	assert.NoError(t, err)
	assert.Equal(t, expectedValue, result)

	mockRepo.AssertExpectations(t)
}
