// services/config_service_test.go

package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"nps-config-service/internal/modules/config-manager/apis/dtos"
	"nps-config-service/internal/modules/config-manager/repositories/mocks"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRegisterWebhookService(t *testing.T) {
	mockRepo := new(mocks.MockRepository)
	// services.ConfigService
	// webHookService := &services.WebhookService{Repo: mockRepo}
	serviceInterface := NewWebhookService(mockRepo)
	webHookService := serviceInterface.(*WebhookService) // type assertion

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
	result, err := webHookService.RegisterWebhookService(req)
	assert.Nil(t, err)
	assert.Equal(t, "Webhook registered successfully for service: service1 in environment: prod", result)

	// Test duplicate webhook
	mockRepo.On("Get", mock.Anything, key).Return(`[{"url":"http://example.com/webhook","method":"POST"}]`, nil).Once()

	result, err = webHookService.RegisterWebhookService(req)
	assert.NotNil(t, err)
	assert.Equal(t, "webhook with URL 'http://example.com/webhook' and method 'POST' already exists", err.Error())

	// Assert expectations
	mockRepo.AssertExpectations(t)
}

func TestGetWebhooks(t *testing.T) {
	mockRepo := new(mocks.MockRepository)
	webHookService := &WebhookService{Repo: mockRepo}

	// Test data
	env := "prod"
	service := "service1"
	key := fmt.Sprintf("/webhooks/%s/%s", env, service)
	expectedWebhooks := `[{"url":"http://example.com/webhook","method":"POST","environment":"prod","serviceName":"service1"}]`

	mockRepo.On("Get", mock.Anything, key).Return(expectedWebhooks, nil).Once()

	// Test success case
	webhooks, err := webHookService.GetWebhooks(env, service)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(webhooks))
	assert.Equal(t, "http://example.com/webhook", webhooks[0].URL)
	assert.Equal(t, "POST", webhooks[0].Method)

	// Test no webhooks found
	mockRepo.On("Get", mock.Anything, key).Return("", fmt.Errorf("no webhooks found")).Once()
	webhooks, err = webHookService.GetWebhooks(env, service)
	assert.NotNil(t, err)
	assert.Equal(t, "no webhooks found for prod/service1", err.Error())

	// Assert expectations
	mockRepo.AssertExpectations(t)
}

func TestDeleteWebhook(t *testing.T) {
	mockRepo := new(mocks.MockRepository)
	webHookService := &WebhookService{Repo: mockRepo}

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

	result, err := webHookService.DeleteWebhook(env, service, url, method)
	assert.Nil(t, err)
	assert.Equal(t, "Webhook deleted successfully", result)

	// Test webhook not found
	result, err = webHookService.DeleteWebhook(env, service, "http://notfound.com/webhook", method)
	assert.NotNil(t, err)
	assert.Equal(t, "webhook with URL 'http://notfound.com/webhook' and method 'POST' not found", err.Error())

	// Assert expectations
	mockRepo.AssertExpectations(t)
}

func TestDeleteAllWebhooks(t *testing.T) {
	mockRepo := new(mocks.MockRepository)
	webHookService := &WebhookService{Repo: mockRepo}

	// Test data
	env := "prod"
	service := "service1"
	key := fmt.Sprintf("/webhooks/%s/%s", env, service)

	// Test success case
	mockRepo.On("Delete", mock.Anything, key).Return(nil).Once()

	err := webHookService.DeleteAllWebhooks(env, service)
	assert.Nil(t, err)

	// Test failure case
	mockRepo.On("Delete", mock.Anything, key).Return(fmt.Errorf("failed to delete")).Once()

	err = webHookService.DeleteAllWebhooks(env, service)
	assert.NotNil(t, err)

	// Assert expectations
	mockRepo.AssertExpectations(t)
}
func TestNotifyWebhook_Success(t *testing.T) {
	hook := dtos.RegisterWebhookRequest{
		URL:         "http://example.com/webhook",
		Method:      "POST",
		Environment: "dev",
		ServiceName: "test-service",
	}

	data := map[string]interface{}{"key": "value"}

	// Mocking HTTP server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		var payload map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Errorf("Failed to decode request body: %v", err)
		}
		if status := r.Response.StatusCode; status != http.StatusOK {
			t.Errorf("Expected status 200 OK, got %d", status)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	// Replace webhook URL with mock server URL
	hook.URL = ts.URL

	// Create the WebhookService and invoke the method
	mockRepo := new(mocks.MockRepository)

	webHookService := &WebhookService{Repo: mockRepo}
	webHookService.NotifyWebhook(hook, data)

	// You could add more assertions to ensure retry behavior and logging, if necessary.
}
