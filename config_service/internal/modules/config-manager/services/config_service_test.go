package services

import (
	"errors"
	"nps-config-service/internal/modules/config-manager/models"
	repoMock "nps-config-service/internal/modules/config-manager/repositories/mocks"
	serviceMock "nps-config-service/internal/modules/config-manager/services/mocks"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestStoreConfigService_Success(t *testing.T) {
	mockRepo := new(repoMock.MockRepository)
	mockWebhook := new(serviceMock.MockWebhookService)

	service := NewConfigService(mockRepo, mockWebhook)

	env := "dev"
	serviceName := "test-service"
	req := map[string]interface{}{"key": "value"}

	storedResponse := map[string]interface{}{"status": "success"}
	webhookData := `[{"url":"http://example.com/webhook"}]`

	mockRepo.On("StoreConfig", env, serviceName, req).Return(storedResponse, nil)
	mockRepo.On("Get", mock.Anything, "/webhooks/dev/test-service").Return(webhookData, nil)

	mockWebhook.On("NotifyWebhook", mock.Anything, req)

	result, err := service.StoreConfigService(env, serviceName, req)

	assert.NoError(t, err)
	assert.Equal(t, storedResponse, result)

	mockRepo.AssertExpectations(t)

}

func TestStoreConfigService_InvalidWebhookData(t *testing.T) {
	mockRepo := new(repoMock.MockRepository)
	mockWebhook := new(serviceMock.MockWebhookService)

	service := NewConfigService(mockRepo, mockWebhook)

	env := "dev"
	serviceName := "test-service"
	req := map[string]interface{}{"key": "value"}

	storedResponse := map[string]interface{}{"status": "success"}

	// Invalid webhook data (malformed JSON)
	webhookData := `[{"url":"http://example.com/webhook"` // Missing closing bracket

	// Set up the mocks
	mockRepo.On("StoreConfig", env, serviceName, req).Return(storedResponse, nil)
	mockRepo.On("Get", mock.Anything, "/webhooks/dev/test-service").Return(webhookData, nil)

	// Run the service method
	result, err := service.StoreConfigService(env, serviceName, req)

	// Assert error occurs due to invalid JSON unmarshalling
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "invalid webhook data stored for /webhooks/dev/test-service")

	// Ensure expectations for mockRepo are met
	mockRepo.AssertExpectations(t)
	mockWebhook.AssertExpectations(t)
}

func TestStoreConfigService_StoreError(t *testing.T) {
	mockRepo := new(repoMock.MockRepository)
	mockWebhook := new(serviceMock.MockWebhookService)
	service := NewConfigService(mockRepo, mockWebhook)

	env := "dev"
	serviceName := "test-service"
	req := map[string]interface{}{"key": "value"}

	mockRepo.On("StoreConfig", env, serviceName, req).Return(nil, errors.New("store failed"))

	result, err := service.StoreConfigService(env, serviceName, req)

	assert.Nil(t, result)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no webhook registered")

	mockRepo.AssertExpectations(t)
}

func TestStoreConfigService_NoWebhookFound(t *testing.T) {
	mockRepo := new(repoMock.MockRepository)
	mockWebhook := new(serviceMock.MockWebhookService)
	service := NewConfigService(mockRepo, mockWebhook)

	env := "dev"
	serviceName := "test-service"
	req := map[string]interface{}{"key": "value"}

	storedResponse := map[string]interface{}{"status": "success"}

	mockRepo.On("StoreConfig", env, serviceName, req).Return(storedResponse, nil)
	mockRepo.On("Get", mock.Anything, "/webhooks/dev/test-service").Return("", errors.New("not found"))

	result, err := service.StoreConfigService(env, serviceName, req)

	assert.Nil(t, result)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no webhook registered")

	mockRepo.AssertExpectations(t)
}

func TestGetConfigService_Success(t *testing.T) {
	mockRepo := new(repoMock.MockRepository)
	mockWebhook := new(serviceMock.MockWebhookService)
	service := NewConfigService(mockRepo, mockWebhook)

	serviceName := "test-service"
	env := "dev"
	expectedConfig := map[string]interface{}{"key": "value"}

	mockRepo.On("GetConfig", serviceName, env).Return(expectedConfig, nil)

	result, err := service.GetConfigService(serviceName, env)

	assert.NoError(t, err)
	assert.Equal(t, expectedConfig, result)

	mockRepo.AssertExpectations(t)
}

func TestGetConfigValueService_Success(t *testing.T) {
	mockRepo := new(repoMock.MockRepository)
	mockWebhook := new(serviceMock.MockWebhookService)
	service := NewConfigService(mockRepo, mockWebhook)

	serviceName := "test-service"
	env := "dev"
	key := "some-key"
	expectedValue := "some-value"

	mockRepo.On("GetConfigValue", serviceName, env, key).Return(expectedValue, nil)

	result, err := service.GetConfigValueService(serviceName, env, key)

	assert.NoError(t, err)
	assert.Equal(t, expectedValue, result)

	mockRepo.AssertExpectations(t)
}

func TestGetConfigMetadataService_Success(t *testing.T) {
	mockRepo := new(repoMock.MockRepository)
	mockWebhook := new(serviceMock.MockWebhookService)
	service := NewConfigService(mockRepo, mockWebhook)

	serviceName := "test-service"
	env := "dev"

	// Define the expected metadata as a *models.ConfigMetadata type (assuming it’s a struct)
	expectedMetadata := &models.ConfigMetadata{
		LastModifiedBy: "admin",
		ChangeHistory:  []string{"initial commit", "updated config"},
		LastModifiedAt: time.Now(),
	}

	// Mock the repository to return the correct type
	mockRepo.On("GetConfigMetadata", serviceName, env).Return(expectedMetadata, nil)

	// Call the service method
	result, err := service.GetConfigMetadataService(serviceName, env)

	// Assert no errors and check the result
	assert.NoError(t, err)
	assert.Equal(t, expectedMetadata, result)

	// Assert that the mock expectations were met
	mockRepo.AssertExpectations(t)
}
