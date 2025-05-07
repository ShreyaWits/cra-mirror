package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"nps-config-service/internal/modules/config-manager/services/mocks"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func TestConfigHandler_StoreConfigHandler(t *testing.T) {
	// Test cases
	tests := []struct {
		name           string
		environment    string
		service        string
		requestBody    map[string]interface{}
		mockResponse   any
		mockError      error
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name:        "Success - Store config",
			environment: "dev",
			service:     "test-service",
			requestBody: map[string]interface{}{
				"key1": "value1",
				"key2": "value2",
			},
			mockResponse: map[string]interface{}{
				"message": "Config stored successfully",
			},
			mockError:      nil,
			expectedStatus: 200,
			expectedBody: map[string]interface{}{
				"status_code": float64(200),
				"message":     "Config stored successfully",
				"data": map[string]interface{}{
					"message": "Config stored successfully",
				},
			},
		},
		{
			name:           "Error - Invalid request body",
			environment:    "dev",
			service:        "test-service",
			requestBody:    nil,
			mockResponse:   nil,
			mockError:      nil,
			expectedStatus: 400,
			expectedBody: map[string]interface{}{
				"status_code": float64(400),
				"message":     "Invalid config data",
				"error":       "Invalid config data",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Always create and assign the mock service
			mockService := new(mocks.MockConfigService)
			handler := NewConfigHandler(mockService)

			app := fiber.New()
			app.Put("/:environment/:service", func(c *fiber.Ctx) error {
				if tt.requestBody != nil {
					bodyMap := make(map[string]interface{})
					for k, v := range tt.requestBody {
						bodyMap[k] = v
					}
					c.Locals("contextData", &bodyMap)
				} else {
					c.Locals("contextData", nil)
				}
				return handler.StoreConfigHandler(c)
			})

			if tt.requestBody != nil {
				mockService.On("StoreConfigService", tt.environment, tt.service, tt.requestBody).
					Return(tt.mockResponse, tt.mockError).Once()
			}

			var req *http.Request
			if tt.requestBody != nil {
				body, _ := json.Marshal(tt.requestBody)
				req = httptest.NewRequest("PUT", "/"+tt.environment+"/"+tt.service, bytes.NewBuffer(body))
				req.Header.Set("Content-Type", "application/json")
			} else {
				req = httptest.NewRequest("PUT", "/"+tt.environment+"/"+tt.service, nil)
			}

			resp, err := app.Test(req)
			assert.NoError(t, err)

			var responseBody map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&responseBody)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
			assert.Equal(t, tt.expectedBody, responseBody)

			mockService.AssertExpectations(t)
		})
	}

}

func TestConfigHandler_GetfullConfig(t *testing.T) {
	// Test cases
	tests := []struct {
		name           string
		environment    string
		service        string
		mockResponse   interface{}
		mockError      error
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name:        "Success - Get config",
			environment: "dev",
			service:     "test-service",
			mockResponse: map[string]interface{}{
				"key1": "value1",
				"key2": "value2",
			},
			mockError:      nil,
			expectedStatus: 200,
			expectedBody: map[string]interface{}{
				"status_code": float64(200),
				"message":     "Config fetched successfully",
				"data": map[string]interface{}{
					"key1": "value1",
					"key2": "value2",
				},
			},
		},
		{
			name:           "Error - Config not found",
			environment:    "dev",
			service:        "test-service",
			mockResponse:   nil,
			mockError:      assert.AnError,
			expectedStatus: 500,
			expectedBody: map[string]interface{}{
				"status_code": float64(500),
				"message":     "Failed to fetch config",
				"error":       assert.AnError.Error(),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Recreate mockService and handler per test to avoid leaks
			mockService := new(mocks.MockConfigService)
			handler := NewConfigHandler(mockService)

			// Create a new Fiber app for testing
			app := fiber.New()

			// Set up the route
			app.Get("/:environment/:service", handler.GetfullConfig)

			// Set up mock expectations
			mockService.On("GetConfigService", tt.service, tt.environment).
				Return(tt.mockResponse, tt.mockError)

			// Create test request
			req := httptest.NewRequest("GET", "/"+tt.environment+"/"+tt.service, nil)

			// Perform request
			resp, err := app.Test(req)
			assert.NoError(t, err)

			// Assert response
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			var responseBody map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&responseBody)
			assert.NoError(t, err)

			assert.Equal(t, tt.expectedBody, responseBody)

			// Verify mock expectations
			mockService.AssertExpectations(t)
		})
	}
}

func TestConfigHandler_GetByValue(t *testing.T) {
	// Test cases
	tests := []struct {
		name           string
		environment    string
		service        string
		key            string
		mockResponse   interface{}
		mockError      error
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name:           "Success - Get config value",
			environment:    "dev",
			service:        "test-service",
			key:            "key1",
			mockResponse:   "value1",
			mockError:      nil,
			expectedStatus: 200,
			expectedBody: map[string]interface{}{
				"status_code": float64(200),
				"message":     "Value fetched successfully",
				"data": map[string]interface{}{
					"key1": "value1",
				},
			},
		},
		{
			name:           "Error - Value not found",
			environment:    "dev",
			service:        "test-service",
			key:            "key1",
			mockResponse:   nil,
			mockError:      assert.AnError,
			expectedStatus: 500,
			expectedBody: map[string]interface{}{
				"status_code": float64(500),
				"message":     "Failed to get config value",
				"error":       assert.AnError.Error(),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create new mockService and handler per test
			mockService := new(mocks.MockConfigService)
			handler := NewConfigHandler(mockService)

			// Create a new Fiber app for testing
			app := fiber.New()

			// Set up the route
			app.Get("/:environment/:service/:key", handler.GetByValue)

			// Set up mock expectations
			mockService.On("GetConfigValueService", tt.service, tt.environment, tt.key).
				Return(tt.mockResponse, tt.mockError)

			// Create test request
			req := httptest.NewRequest("GET", "/"+tt.environment+"/"+tt.service+"/"+tt.key, nil)

			// Perform request
			resp, err := app.Test(req)
			assert.NoError(t, err)

			// Assert response
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			var responseBody map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&responseBody)
			assert.NoError(t, err)

			assert.Equal(t, tt.expectedBody, responseBody)

			// Verify mock expectations
			mockService.AssertExpectations(t)
		})
	}
}

func TestConfigHandler_GetByMetadata(t *testing.T) {
	// Test cases
	tests := []struct {
		name           string
		environment    string
		service        string
		mockResponse   interface{}
		mockError      error
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name:        "Success - Get metadata",
			environment: "dev",
			service:     "test-service",
			mockResponse: map[string]interface{}{
				"version":     "1.0",
				"lastUpdated": "2024-03-20",
			},
			mockError:      nil,
			expectedStatus: 200,
			expectedBody: map[string]interface{}{
				"status_code": float64(200),
				"message":     "Metadata fetched successfully",
				"data": map[string]interface{}{
					"version":     "1.0",
					"lastUpdated": "2024-03-20",
				},
			},
		},
		{
			name:           "Error - Metadata not found",
			environment:    "dev",
			service:        "test-service",
			mockResponse:   nil,
			mockError:      assert.AnError,
			expectedStatus: 500,
			expectedBody: map[string]interface{}{
				"status_code": float64(500),
				"message":     "Failed to fetch metadata",
				"error":       assert.AnError.Error(),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create fresh mockService and handler per test
			mockService := new(mocks.MockConfigService)
			handler := NewConfigHandler(mockService)

			app := fiber.New()
			app.Get("/:environment/:service/metadata", handler.GetByMetadata)

			// Set up mock expectations
			mockService.On("GetConfigMetadataService", tt.service, tt.environment).
				Return(tt.mockResponse, tt.mockError)

			// Create and send request
			req := httptest.NewRequest("GET", "/"+tt.environment+"/"+tt.service+"/metadata", nil)
			resp, err := app.Test(req)
			assert.NoError(t, err)

			// Check response
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			var responseBody map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&responseBody)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedBody, responseBody)

			mockService.AssertExpectations(t)
		})
	}
}
