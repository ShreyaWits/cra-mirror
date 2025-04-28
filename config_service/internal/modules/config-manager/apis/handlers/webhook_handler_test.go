package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"nps-config-service/internal/modules/config-manager/apis/dtos"
	"nps-config-service/internal/modules/config-manager/services/mocks"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestWebhookHandler_RegisterWebhook(t *testing.T) {
	mockService := new(mocks.MockWebhookService)
	handler := NewWebhookHandler(mockService)

	tests := []struct {
		name           string
		requestBody    interface{}
		mockResponse   interface{}
		mockError      error
		expectedStatus int
		expectedBody   map[string]interface{}
		setupMock      func(*mocks.MockWebhookService)
	}{
		{
			name: "Success - Register webhook",
			requestBody: dtos.RegisterWebhookRequest{
				URL:         "http://example.com/webhook",
				Environment: "dev",
				ServiceName: "test-service",
				Method:      "POST",
			},
			mockResponse:   "Webhook registered successfully for service: test-service in environment: dev",
			mockError:      nil,
			expectedStatus: 200,
			expectedBody: map[string]interface{}{
				"status_code": float64(200),
				"message":     "Webhook registered successfully",
				"data":        "Webhook registered successfully for service: test-service in environment: dev",
			},
			setupMock: func(m *mocks.MockWebhookService) {
				m.On("RegisterWebhookService", mock.AnythingOfType("dtos.RegisterWebhookRequest")).
					Return("Webhook registered successfully for service: test-service in environment: dev", nil)
			},
		},
		{
			name:           "Error - Invalid request body",
			requestBody:    "invalid-json",
			mockResponse:   nil,
			mockError:      nil,
			expectedStatus: 400,
			expectedBody: map[string]interface{}{
				"status_code": float64(400),
				"message":     "Invalid config data",
				"error":       "Context data missing or invalid",
			},
			setupMock: func(m *mocks.MockWebhookService) {
				// No mock setup needed for invalid request
			},
		},
		{
			name: "Error - Webhook already exists",
			requestBody: dtos.RegisterWebhookRequest{
				URL:         "http://example.com/webhook",
				Environment: "dev",
				ServiceName: "test-service",
				Method:      "POST",
			},
			mockResponse:   nil,
			mockError:      errors.New("webhook already exists"),
			expectedStatus: 409,
			expectedBody: map[string]interface{}{
				"status_code": float64(409),
				"message":     "Webhook already exists",
				"error":       "webhook already exists",
			},
			setupMock: func(m *mocks.MockWebhookService) {
				m.On("RegisterWebhookService", mock.AnythingOfType("dtos.RegisterWebhookRequest")).
					Return(nil, errors.New("webhook already exists"))
			},
		},
		{
			name: "Error - Invalid URL format",
			requestBody: dtos.RegisterWebhookRequest{
				URL:         "invalid-url",
				Environment: "dev",
				ServiceName: "test-service",
				Method:      "POST",
			},
			mockResponse:   nil,
			mockError:      nil,
			expectedStatus: 400,
			expectedBody: map[string]interface{}{
				"status_code": float64(400),
				"message":     "Invalid webhook URL",
				"error":       "invalid URL format",
			},
			setupMock: func(m *mocks.MockWebhookService) {
				// No mock setup needed for invalid URL
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mock for each test
			mockService.ExpectedCalls = nil
			if tt.setupMock != nil {
				tt.setupMock(mockService)
			}

			app := fiber.New()
			app.Post("/webhook", handler.RegisterWebhook)

			// Create test request
			var req *http.Request
			switch body := tt.requestBody.(type) {
			case string:
				req = httptest.NewRequest("POST", "/webhook", bytes.NewBufferString(body))
			case dtos.RegisterWebhookRequest:
				jsonBody, err := json.Marshal(body)
				assert.NoError(t, err)
				req = httptest.NewRequest("POST", "/webhook", bytes.NewBuffer(jsonBody))
			}
			req.Header.Set("Content-Type", "application/json")

			// Set context data
			app.Use(func(c *fiber.Ctx) error {
				if req, ok := tt.requestBody.(dtos.RegisterWebhookRequest); ok {
					bodyMap := make(map[string]interface{})
					jsonBody, err := json.Marshal(req)
					assert.NoError(t, err)
					err = json.Unmarshal(jsonBody, &bodyMap)
					assert.NoError(t, err)
					c.Locals("contextData", &bodyMap)
				}
				return c.Next()
			})

			// Perform request
			resp, err := app.Test(req)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			var responseBody map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&responseBody)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedBody, responseBody)

			mockService.AssertExpectations(t)
		})
	}
}

func TestWebhookHandler_GetWebhooks(t *testing.T) {
	mockService := new(mocks.MockWebhookService)
	handler := NewWebhookHandler(mockService)

	tests := []struct {
		name           string
		environment    string
		service        string
		mockResponse   []dtos.RegisterWebhookRequest
		mockError      error
		expectedStatus int
		expectedBody   map[string]interface{}
		setupMock      func(*mocks.MockWebhookService)
	}{
		{
			name:        "Success - Get webhooks",
			environment: "dev",
			service:     "test-service",
			mockResponse: []dtos.RegisterWebhookRequest{
				{
					URL:         "http://example.com/webhook1",
					Environment: "dev",
					ServiceName: "test-service",
					Method:      "POST",
				},
				{
					URL:         "http://example.com/webhook2",
					Environment: "dev",
					ServiceName: "test-service",
					Method:      "GET",
				},
			},
			mockError:      nil,
			expectedStatus: 200,
			expectedBody: map[string]interface{}{
				"status_code": float64(200),
				"message":     "Webhooks retrieved successfully",
				"data": []interface{}{
					map[string]interface{}{
						"url":         "http://example.com/webhook1",
						"environment": "dev",
						"serviceName": "test-service",
						"method":      "POST",
					},
					map[string]interface{}{
						"url":         "http://example.com/webhook2",
						"environment": "dev",
						"serviceName": "test-service",
						"method":      "GET",
					},
				},
			},
			setupMock: func(m *mocks.MockWebhookService) {
				m.On("GetWebhooks", "dev", "test-service").
					Return([]dtos.RegisterWebhookRequest{
						{
							URL:         "http://example.com/webhook1",
							Environment: "dev",
							ServiceName: "test-service",
							Method:      "POST",
						},
						{
							URL:         "http://example.com/webhook2",
							Environment: "dev",
							ServiceName: "test-service",
							Method:      "GET",
						},
					}, nil)
			},
		},
		{
			name:           "Error - No webhooks found",
			environment:    "dev",
			service:        "test-service",
			mockResponse:   nil,
			mockError:      errors.New("no webhooks found"),
			expectedStatus: 404,
			expectedBody: map[string]interface{}{
				"status_code": float64(404),
				"message":     "Failed to retrieve webhooks",
				"error":       "no webhooks found",
			},
			setupMock: func(m *mocks.MockWebhookService) {
				m.On("GetWebhooks", "dev", "test-service").
					Return(nil, errors.New("no webhooks found"))
			},
		},
		{
			name:           "Error - Invalid environment",
			environment:    "",
			service:        "test-service",
			mockResponse:   nil,
			mockError:      nil,
			expectedStatus: 400,
			expectedBody: map[string]interface{}{
				"status_code": float64(400),
				"message":     "Invalid environment",
				"error":       "environment cannot be empty",
			},
			setupMock: func(m *mocks.MockWebhookService) {
				// No mock setup needed for invalid environment
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mock for each test
			mockService.ExpectedCalls = nil
			if tt.setupMock != nil {
				tt.setupMock(mockService)
			}

			app := fiber.New()
			app.Get("/:environment/:service/webhooks", handler.GetWebhooks)

			// Create test request
			req := httptest.NewRequest("GET", "/"+tt.environment+"/"+tt.service+"/webhooks", nil)

			// Perform request
			resp, err := app.Test(req)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			var responseBody map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&responseBody)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedBody, responseBody)

			mockService.AssertExpectations(t)
		})
	}
}

func TestWebhookHandler_DeleteWebhook(t *testing.T) {
	mockService := new(mocks.MockWebhookService)
	handler := NewWebhookHandler(mockService)

	tests := []struct {
		name           string
		environment    string
		service        string
		requestBody    dtos.RegisterWebhookRequest
		mockResponse   string
		mockError      error
		expectedStatus int
		expectedBody   map[string]interface{}
		setupMock      func(*mocks.MockWebhookService)
	}{
		{
			name:        "Success - Delete webhook",
			environment: "dev",
			service:     "test-service",
			requestBody: dtos.RegisterWebhookRequest{
				URL:         "http://example.com/webhook",
				Environment: "dev",
				ServiceName: "test-service",
				Method:      "POST",
			},
			mockResponse:   "Webhook deleted successfully",
			mockError:      nil,
			expectedStatus: 200,
			expectedBody: map[string]interface{}{
				"status_code": float64(200),
				"message":     "Webhook deleted successfully",
				"data":        "Webhook deleted successfully",
			},
			setupMock: func(m *mocks.MockWebhookService) {
				m.On("DeleteWebhook", "dev", "test-service", "http://example.com/webhook", "POST").
					Return("Webhook deleted successfully", nil)
			},
		},
		{
			name:           "Error - Invalid request body",
			environment:    "dev",
			service:        "test-service",
			requestBody:    dtos.RegisterWebhookRequest{},
			mockResponse:   "",
			mockError:      nil,
			expectedStatus: 400,
			expectedBody: map[string]interface{}{
				"status_code": float64(400),
				"message":     "Invalid delete request",
				"error":       "Context data missing or invalid",
			},
			setupMock: func(m *mocks.MockWebhookService) {
				// No mock setup needed for invalid request
			},
		},
		{
			name:        "Error - Webhook not found",
			environment: "dev",
			service:     "test-service",
			requestBody: dtos.RegisterWebhookRequest{
				URL:         "http://example.com/webhook",
				Environment: "dev",
				ServiceName: "test-service",
				Method:      "POST",
			},
			mockResponse:   "",
			mockError:      errors.New("webhook not found"),
			expectedStatus: 500,
			expectedBody: map[string]interface{}{
				"status_code": float64(500),
				"message":     "Failed to delete webhook",
				"error":       "webhook not found",
			},
			setupMock: func(m *mocks.MockWebhookService) {
				m.On("DeleteWebhook", "dev", "test-service", "http://example.com/webhook", "POST").
					Return("", errors.New("webhook not found"))
			},
		},
		{
			name:        "Error - Invalid URL format",
			environment: "dev",
			service:     "test-service",
			requestBody: dtos.RegisterWebhookRequest{
				URL:         "invalid-url",
				Environment: "dev",
				ServiceName: "test-service",
				Method:      "POST",
			},
			mockResponse:   "",
			mockError:      nil,
			expectedStatus: 400,
			expectedBody: map[string]interface{}{
				"status_code": float64(400),
				"message":     "Invalid webhook URL",
				"error":       "invalid URL format",
			},
			setupMock: func(m *mocks.MockWebhookService) {
				// No mock setup needed for invalid URL
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mock for each test
			mockService.ExpectedCalls = nil
			if tt.setupMock != nil {
				tt.setupMock(mockService)
			}

			app := fiber.New()
			app.Delete("/:environment/:service/webhook", handler.DeleteWebhook)

			// Create test request
			jsonBody, err := json.Marshal(tt.requestBody)
			assert.NoError(t, err)
			req := httptest.NewRequest("DELETE", "/"+tt.environment+"/"+tt.service+"/webhook", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			// Set context data
			app.Use(func(c *fiber.Ctx) error {
				bodyMap := make(map[string]interface{})
				err := json.Unmarshal(jsonBody, &bodyMap)
				assert.NoError(t, err)
				c.Locals("contextData", &bodyMap)
				return c.Next()
			})

			// Perform request
			resp, err := app.Test(req)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			var responseBody map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&responseBody)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedBody, responseBody)

			mockService.AssertExpectations(t)
		})
	}
}
