// services/config_service_test.go

package services

import (
	"context"
	"fmt"
	"nps-config-service/internal/modules/config-manager/apis/dtos"
	"nps-config-service/internal/modules/config-manager/repositories/mocks"
	"nps-config-service/pkg/observability"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupWebhookTestService(t *testing.T) (IWebhookService, *mocks.MockRepository, *observability.ObservabilityStack) {
	mockRepo := new(mocks.MockRepository)
	obsStack := observability.NewObservabilityStack("webhook-service-test")
	service := NewWebhookService(mockRepo, obsStack)
	return service, mockRepo, obsStack
}

func TestRegisterWebhookService(t *testing.T) {
	tests := []struct {
		name           string
		request        dtos.RegisterWebhookRequest
		mockSetup      func(*mocks.MockRepository)
		expectedResult *dtos.SuccessResponse
		expectedError  *dtos.ServiceErrorResponse
	}{
		{
			name: "successful registration",
			request: dtos.RegisterWebhookRequest{
				URL:         "http://example.com/webhook",
				Method:      "POST",
				Environment: "dev",
				ServiceName: "test-service",
			},
			mockSetup: func(m *mocks.MockRepository) {
				m.On("GetEtcdKey", mock.Anything, "/webhooks/dev/test-service").Return("[]", nil)
				m.On("SetEtcdKey", mock.Anything, "/webhooks/dev/test-service", mock.Anything, mock.Anything).Return(nil)
			},
			expectedResult: &dtos.SuccessResponse{
				StatusCode: 201,
				Message:    "Webhook registered successfully",
				Data: map[string]interface{}{
					"webhook": dtos.RegisterWebhookRequest{
						URL:         "http://example.com/webhook",
						Method:      "POST",
						Environment: "dev",
						ServiceName: "test-service",
					},
				},
			},
			expectedError: nil,
		},
		{
			name: "webhook already exists",
			request: dtos.RegisterWebhookRequest{
				URL:         "http://example.com/webhook",
				Method:      "POST",
				Environment: "dev",
				ServiceName: "test-service",
			},
			mockSetup: func(m *mocks.MockRepository) {
				existingWebhooks := `[{"url":"http://example.com/webhook","method":"POST","environment":"dev","serviceName":"test-service"}]`
				m.On("GetEtcdKey", mock.Anything, "/webhooks/dev/test-service").Return(existingWebhooks, nil)
			},
			expectedResult: nil,
			expectedError: &dtos.ServiceErrorResponse{
				StatusCode:   409,
				ErrorCode:    "Webhook already exists",
				ErrorMessage: "webhook with URL 'http://example.com/webhook' and method 'POST' already exists",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, mockRepo, _ := setupWebhookTestService(t)
			tt.mockSetup(mockRepo)

			ctx := context.Background()
			result, err := service.RegisterWebhookService(ctx, tt.request)

			if tt.expectedError != nil {
				assert.NotNil(t, err)
				assert.Equal(t, tt.expectedError.StatusCode, err.StatusCode)
				assert.Equal(t, tt.expectedError.ErrorCode, err.ErrorCode)
				assert.Equal(t, tt.expectedError.ErrorMessage, err.ErrorMessage)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, tt.expectedResult.StatusCode, result.StatusCode)
				assert.Equal(t, tt.expectedResult.Message, result.Message)
				assert.Equal(t, tt.expectedResult.Data, result.Data)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestGetWebhooks(t *testing.T) {
	tests := []struct {
		name           string
		env            string
		service        string
		mockSetup      func(*mocks.MockRepository)
		expectedResult []dtos.RegisterWebhookRequest
		expectedError  *dtos.ServiceErrorResponse
	}{
		{
			name:    "successful get",
			env:     "dev",
			service: "test-service",
			mockSetup: func(m *mocks.MockRepository) {
				webhooks := `[{"url":"http://example.com/webhook","method":"POST","environment":"dev","serviceName":"test-service"}]`
				m.On("GetEtcdKey", mock.Anything, "/webhooks/dev/test-service").Return(webhooks, nil)
			},
			expectedResult: []dtos.RegisterWebhookRequest{
				{
					URL:         "http://example.com/webhook",
					Method:      "POST",
					Environment: "dev",
					ServiceName: "test-service",
				},
			},
			expectedError: nil,
		},
		{
			name:    "no webhooks found",
			env:     "dev",
			service: "test-service",
			mockSetup: func(m *mocks.MockRepository) {
				m.On("GetEtcdKey", mock.Anything, "/webhooks/dev/test-service").Return("", fmt.Errorf("no webhooks found"))
			},
			expectedResult: nil,
			expectedError: &dtos.ServiceErrorResponse{
				StatusCode:   404,
				ErrorCode:    "No webhooks found",
				ErrorMessage: "no webhooks found for dev/test-service: no webhooks found",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, mockRepo, _ := setupWebhookTestService(t)
			tt.mockSetup(mockRepo)

			ctx := context.Background()
			result, err := service.GetWebhooks(ctx, tt.env, tt.service)

			if tt.expectedError != nil {
				assert.NotNil(t, err)
				assert.Equal(t, tt.expectedError.StatusCode, err.StatusCode)
				assert.Equal(t, tt.expectedError.ErrorCode, err.ErrorCode)
				assert.Contains(t, err.ErrorMessage, tt.expectedError.ErrorMessage)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, tt.expectedResult, result)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestDeleteWebhook(t *testing.T) {
	tests := []struct {
		name           string
		env            string
		service        string
		url            string
		method         string
		mockSetup      func(*mocks.MockRepository)
		expectedResult string
		expectedError  *dtos.ServiceErrorResponse
	}{
		{
			name:    "successful delete",
			env:     "dev",
			service: "test-service",
			url:     "http://example.com/webhook",
			method:  "POST",
			mockSetup: func(m *mocks.MockRepository) {
				webhooks := `[{"url":"http://example.com/webhook","method":"POST","environment":"dev","serviceName":"test-service"}]`
				m.On("GetEtcdKey", mock.Anything, "/webhooks/dev/test-service").Return(webhooks, nil)
				m.On("SetEtcdKey", mock.Anything, "/webhooks/dev/test-service", "[]", mock.Anything).Return(nil)
			},
			expectedResult: "Webhook deleted successfully",
			expectedError:  nil,
		},
		{
			name:    "webhook not found",
			env:     "dev",
			service: "test-service",
			url:     "http://example.com/webhook",
			method:  "POST",
			mockSetup: func(m *mocks.MockRepository) {
				webhooks := `[{"url":"http://other.com/webhook","method":"POST","environment":"dev","serviceName":"test-service"}]`
				m.On("GetEtcdKey", mock.Anything, "/webhooks/dev/test-service").Return(webhooks, nil)
			},
			expectedResult: "",
			expectedError: &dtos.ServiceErrorResponse{
				StatusCode:   404,
				ErrorCode:    "Webhook not found",
				ErrorMessage: "webhook with URL 'http://example.com/webhook' and method 'POST' not found",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, mockRepo, _ := setupWebhookTestService(t)
			tt.mockSetup(mockRepo)

			ctx := context.Background()
			result, err := service.DeleteWebhook(ctx, tt.env, tt.service, tt.url, tt.method)

			if tt.expectedError != nil {
				assert.NotNil(t, err)
				assert.Equal(t, tt.expectedError.StatusCode, err.StatusCode)
				assert.Equal(t, tt.expectedError.ErrorCode, err.ErrorCode)
				assert.Equal(t, tt.expectedError.ErrorMessage, err.ErrorMessage)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, tt.expectedResult, result)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestDeleteAllWebhooks(t *testing.T) {
	tests := []struct {
		name          string
		env           string
		service       string
		mockSetup     func(*mocks.MockRepository)
		expectedError error
	}{
		{
			name:    "successful delete all",
			env:     "dev",
			service: "test-service",
			mockSetup: func(m *mocks.MockRepository) {
				m.On("DeleteEtcdKey", mock.Anything, "/webhooks/dev/test-service").Return(nil)
			},
			expectedError: nil,
		},
		{
			name:    "delete error",
			env:     "dev",
			service: "test-service",
			mockSetup: func(m *mocks.MockRepository) {
				m.On("DeleteEtcdKey", mock.Anything, "/webhooks/dev/test-service").Return(fmt.Errorf("failed to delete"))
			},
			expectedError: fmt.Errorf("failed to delete"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, mockRepo, _ := setupWebhookTestService(t)
			tt.mockSetup(mockRepo)

			ctx := context.Background()
			err := service.DeleteAllWebhooks(ctx, tt.env, tt.service)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestNotifyWebhook(t *testing.T) {
	hook := dtos.RegisterWebhookRequest{
		URL:         "http://example.com/webhook",
		Method:      "POST",
		Environment: "dev",
		ServiceName: "test-service",
	}
	data := map[string]interface{}{
		"key": "value",
	}

	service, mockRepo, _ := setupWebhookTestService(t)
	ctx := context.Background()

	// NotifyWebhook is an async operation that makes HTTP calls
	// We can't easily test the actual HTTP calls, but we can verify it doesn't panic
	assert.NotPanics(t, func() {
		service.NotifyWebhook(ctx, hook, data)
	})

	mockRepo.AssertExpectations(t)
}
