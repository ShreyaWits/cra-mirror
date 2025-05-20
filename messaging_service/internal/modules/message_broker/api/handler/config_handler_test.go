package handler_test

// import (
// 	"bytes"
// 	"encoding/json"
// 	"net/http/httptest"
// 	"testing"

// 	"github.com/gofiber/fiber/v2"
// 	"github.com/golang/mock/gomock"
// 	"github.com/stretchr/testify/assert"

// 	"messaging_service/internal/config"
// 	"messaging_service/internal/modules/message_broker/api/handler"
// 	mock_service "messaging_service/internal/modules/message_broker/mock"
// )

// func setupConfigHandler(t *testing.T) (*handler.ConfigHandler, *mock_service.MockConfigManagerService, *gomock.Controller) {
// 	ctrl := gomock.NewController(t)
// 	mockSvc := mock_service.NewMockConfigManagerService(ctrl)

// 	env := &config.Env{
// 		ServiceName:      "test-service",
// 		ConfigServiceUrl: "http://localhost:8080",
// 	}

// 	h, _ := handler.NewConfigHandler(mockSvc, env)

// 	return h, mockSvc, ctrl
// }

// func TestNewConfigHandler(t *testing.T) {
// 	ctrl := gomock.NewController(t)
// 	defer ctrl.Finish()

// 	mockSvc := mock_service.NewMockConfigManagerService(ctrl)
// 	validEnv := &config.Env{
// 		ServiceName:      "test-service",
// 		ConfigServiceUrl: "http://localhost:8080",
// 	}

// 	tests := []struct {
// 		name        string
// 		svc         *mock_service.MockConfigManagerService
// 		env         *config.Env
// 		expectError bool
// 	}{
// 		{
// 			name:        "valid parameters",
// 			svc:         mockSvc,
// 			env:         validEnv,
// 			expectError: false,
// 		},
// 		{
// 			name:        "nil service",
// 			svc:         nil,
// 			env:         validEnv,
// 			expectError: true,
// 		},
// 		{
// 			name:        "nil env",
// 			svc:         mockSvc,
// 			env:         nil,
// 			expectError: true,
// 		},
// 		{
// 			name: "empty service name",
// 			svc:  mockSvc,
// 			env: &config.Env{
// 				ConfigServiceUrl: "http://localhost:8080",
// 				// ServiceName is empty
// 			},
// 			expectError: true,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			h, err := handler.NewConfigHandler(tt.svc, tt.env)

// 			if tt.expectError {
// 				assert.Error(t, err)
// 				assert.Nil(t, h)
// 			} else {
// 				assert.NoError(t, err)
// 				assert.NotNil(t, h)
// 			}
// 		})
// 	}
// }

// func TestSetReinitCallback(t *testing.T) {
// 	h, _, ctrl := setupConfigHandler(t)
// 	defer ctrl.Finish()

// 	callbackExecuted := false
// 	callback := func(configuration *config.Config) error {
// 		callbackExecuted = true
// 		return nil
// 	}

// 	// Set the callback
// 	h.SetReinitCallback(callback)

// 	// Create a fiber app for testing
// 	app := fiber.New()
// 	app.Post("/config/update", h.UpdateConfigurations)

// 	// Create a test request with a valid configuration
// 	testConfig := &config.Config{
// 		KafkaBrokers:           []string{"localhost:9092"},
// 		KafkaNumPartitions:     3,
// 		KafkaReplicationFactor: 3,
// 	}

// 	reqBody, _ := json.Marshal(testConfig)
// 	req := httptest.NewRequest("POST", "/config/update", bytes.NewReader(reqBody))
// 	req.Header.Set("Content-Type", "application/json")

// 	// Execute the request
// 	resp, _ := app.Test(req)

// 	// Verify response
// 	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
// 	assert.True(t, callbackExecuted, "Callback should have been executed")
// }

// func TestUpdateConfigurations(t *testing.T) {
// 	h, mockSvc, ctrl := setupConfigHandler(t)
// 	defer ctrl.Finish()

// 	tests := []struct {
// 		name              string
// 		requestBody       interface{}
// 		mockSetup         func()
// 		setCallback       bool
// 		callbackReturns   error
// 		expectedStatus    int
// 		expectedHasConfig bool
// 	}{
// 		{
// 			name: "valid config update",
// 			requestBody: config.Config{
// 				KafkaBrokers:           []string{"localhost:9092"},
// 				KafkaNumPartitions:     3,
// 				KafkaReplicationFactor: 3,
// 			},
// 			mockSetup: func() {
// 				mockSvc.EXPECT().
// 					SetDataToCache(gomock.Any(), "test-service", gomock.Any()).
// 					Return(nil)
// 			},
// 			setCallback:       false,
// 			expectedStatus:    fiber.StatusOK,
// 			expectedHasConfig: true,
// 		},
// 		{
// 			name: "valid config update with successful callback",
// 			requestBody: config.Config{
// 				KafkaBrokers:           []string{"localhost:9092"},
// 				KafkaNumPartitions:     3,
// 				KafkaReplicationFactor: 3,
// 			},
// 			mockSetup: func() {
// 				mockSvc.EXPECT().
// 					SetDataToCache(gomock.Any(), "test-service", gomock.Any()).
// 					Return(nil)
// 			},
// 			setCallback:       true,
// 			callbackReturns:   nil,
// 			expectedStatus:    fiber.StatusOK,
// 			expectedHasConfig: true,
// 		},
// 		{
// 			name:              "invalid request body",
// 			requestBody:       "this is not a valid json object",
// 			mockSetup:         func() {},
// 			setCallback:       false,
// 			expectedStatus:    fiber.StatusBadRequest,
// 			expectedHasConfig: false,
// 		},
// 		{
// 			name: "failed callback",
// 			requestBody: config.Config{
// 				KafkaBrokers:           []string{"localhost:9092"},
// 				KafkaNumPartitions:     3,
// 				KafkaReplicationFactor: 3,
// 			},
// 			mockSetup: func() {
// 				mockSvc.EXPECT().
// 					SetDataToCache(gomock.Any(), "test-service", gomock.Any()).
// 					Return(nil)
// 			},
// 			setCallback:       true,
// 			callbackReturns:   assert.AnError,
// 			expectedStatus:    fiber.StatusInternalServerError,
// 			expectedHasConfig: true,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			// Set up mocks
// 			tt.mockSetup()

// 			// Create a fiber app for testing
// 			app := fiber.New()

// 			// Set callback if needed
// 			if tt.setCallback {
// 				h.SetReinitCallback(func(configuration *config.Config) error {
// 					return tt.callbackReturns
// 				})
// 			} else {
// 				h.SetReinitCallback(nil)
// 			}

// 			app.Post("/config/update", h.UpdateConfigurations)

// 			// Create request body
// 			reqBody, _ := json.Marshal(tt.requestBody)
// 			req := httptest.NewRequest("POST", "/config/update", bytes.NewReader(reqBody))
// 			req.Header.Set("Content-Type", "application/json")

// 			// Execute request
// 			resp, _ := app.Test(req)

// 			// Verify response
// 			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

// 			// Get configuration if it was expected to be set
// 			if tt.expectedHasConfig {
// 				cfg := config.GetConfig()
// 				assert.NotNil(t, cfg)
// 			}
// 		})
// 	}
// }

// func TestHealth(t *testing.T) {
// 	h, _, ctrl := setupConfigHandler(t)
// 	defer ctrl.Finish()

// 	// Create a fiber app for testing
// 	app := fiber.New()
// 	app.Get("/health", h.Health)

// 	// Create a test request
// 	req := httptest.NewRequest("GET", "/health", nil)

// 	// Execute the request
// 	resp, _ := app.Test(req)

// 	// Verify response
// 	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

// 	// Parse response body
// 	var result struct {
// 		Status  string `json:"status"`
// 		Success bool   `json:"success"`
// 	}

// 	err := json.NewDecoder(resp.Body).Decode(&result)
// 	assert.NoError(t, err)
// 	assert.Equal(t, "OK", result.Status)
// 	assert.True(t, result.Success)
// }
