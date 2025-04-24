// services/config_service_test.go

package services

import (
	"fmt"
	"nps-config-service/internal/modules/config-manager/apis/dtos"
	"nps-config-service/internal/modules/config-manager/repositories/mocks"
	"nps-config-service/internal/modules/config-manager/services"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRegisterWebhookService(t *testing.T) {
	mockRepo := new(mocks.MockRepository)
	// services.ConfigService
	configService := &services.WebhookService{Repo: mockRepo}

	// Test data
	req := dtos.RegisterWebhookRequest{
		URL:         "http://example.com/webhook",
		Method:      "POST",
		Environment: "prod",
		ServiceName: "service1",
	}

	key := "/webhooks/prod/service1"
	mockRepo.On("Get", mock.Anything, key).Return("[]", nil).Once()
	mockRepo.On("Set", mock.Anything, key, mock.Anything, mock.Anything).Return(nil).Once()

	// Test success case
	result, err := configService.RegisterWebhookService(req)
	assert.Nil(t, err)
	assert.Equal(t, "Webhook registered successfully for service: service1 in environment: prod", result)

	// Test duplicate webhook
	mockRepo.On("Get", mock.Anything, key).Return(`[{"url":"http://example.com/webhook","method":"POST"}]`, nil).Once()

	result, err = configService.RegisterWebhookService(req)
	assert.NotNil(t, err)
	assert.Equal(t, "webhook with URL 'http://example.com/webhook' and method 'POST' already exists", err.Error())

	// Assert expectations
	mockRepo.AssertExpectations(t)
}

func TestGetWebhooks(t *testing.T) {
	mockRepo := new(mocks.MockRepository)
	configService := &services.WebhookService{Repo: mockRepo}

	// Test data
	env := "prod"
	service := "service1"
	key := fmt.Sprintf("/webhooks/%s/%s", env, service)
	expectedWebhooks := `[{"url":"http://example.com/webhook","method":"POST","environment":"prod","serviceName":"service1"}]`

	mockRepo.On("Get", mock.Anything, key).Return(expectedWebhooks, nil).Once()

	// Test success case
	webhooks, err := configService.GetWebhooks(env, service)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(webhooks))
	assert.Equal(t, "http://example.com/webhook", webhooks[0].URL)
	assert.Equal(t, "POST", webhooks[0].Method)

	// Test no webhooks found
	mockRepo.On("Get", mock.Anything, key).Return("", fmt.Errorf("no webhooks found")).Once()
	webhooks, err = configService.GetWebhooks(env, service)
	assert.NotNil(t, err)
	assert.Equal(t, "no webhooks found for prod/service1", err.Error())

	// Assert expectations
	mockRepo.AssertExpectations(t)
}

func TestDeleteWebhook(t *testing.T) {
	mockRepo := new(mocks.MockRepository)
	configService := &services.WebhookService{Repo: mockRepo}

	// Test data
	env := "prod"
	service := "service1"
	url := "http://example.com/webhook"
	method := "POST"
	key := fmt.Sprintf("/webhooks/%s/%s", env, service)
	webhooks := `[{"url":"http://example.com/webhook","method":"POST","environment":"prod","serviceName":"service1"}]`

	mockRepo.On("Get", mock.Anything, key).Return(webhooks, nil)

	// Test success case
	mockRepo.On("Set", mock.Anything, key, mock.Anything, mock.Anything).Return(nil)

	result, err := configService.DeleteWebhook(env, service, url, method)
	assert.Nil(t, err)
	assert.Equal(t, "Webhook deleted successfully", result)

	// Test webhook not found
	result, err = configService.DeleteWebhook(env, service, "http://notfound.com/webhook", method)
	assert.NotNil(t, err)
	assert.Equal(t, "webhook with URL 'http://notfound.com/webhook' and method 'POST' not found", err.Error())

	// Assert expectations
	mockRepo.AssertExpectations(t)
}

func TestDeleteAllWebhooks(t *testing.T) {
	mockRepo := new(mocks.MockRepository)
	configService := &services.WebhookService{Repo: mockRepo}

	// Test data
	env := "prod"
	service := "service1"
	key := fmt.Sprintf("/webhooks/%s/%s", env, service)

	// Test success case
	mockRepo.On("Delete", mock.Anything, key).Return(nil).Once()

	err := configService.DeleteAllWebhooks(env, service)
	assert.Nil(t, err)

	// Test failure case
	mockRepo.On("Delete", mock.Anything, key).Return(fmt.Errorf("failed to delete")).Once()

	err = configService.DeleteAllWebhooks(env, service)
	assert.NotNil(t, err)

	// Assert expectations
	mockRepo.AssertExpectations(t)
}
