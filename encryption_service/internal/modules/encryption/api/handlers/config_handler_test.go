package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"encryption_microservice/internal/config"
	"encryption_microservice/internal/modules/encryption/api/handlers"
	"encryption_microservice/pkg/observability"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockCacheClient is a mock implementation of the cache client interface
type MockCacheClient struct {
	mock.Mock
}

func (m *MockCacheClient) SetDataToCache(ctx context.Context, key string, value interface{}) error {
	args := m.Called(ctx, key, value)
	return args.Error(0)
}

func TestUpdateConfigurations(t *testing.T) {
	// Initialize mock dependencies
	mockCache := new(MockCacheClient)
	env := &config.EnvConfig{}
	obs := observability.NewObservabilityStack(env)

	// Create the handler
	handler, err := handlers.NewConfigHandler(mockCache, env, obs)
	assert.NoError(t, err)

	// Setup test
	app := fiber.New()
	app.Post("/config", handler.UpdateConfigurations)

	// Create a valid request
	configPayload := map[string]interface{}{
		"environment": "test",
		"method":      "update",
		"serviceName": "encryption_service",
		"values": map[string]interface{}{
			"key_type":                "AES256",
			"key_rotation":            "30d",
			"HASHICORP_VAULT_ADDRess": "http://localhost:8200",
		},
	}
	jsonBytes, _ := json.Marshal(configPayload)

	// Setup mock expectations
	mockCache.On("SetDataToCache", mock.Anything, "config", mock.Anything).Return(nil)

	// Make the request
	req := httptest.NewRequest("POST", "/config", bytes.NewReader(jsonBytes))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	// Verify mock was called
	mockCache.AssertExpectations(t)
}
