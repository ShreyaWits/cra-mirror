package handler

import (
	"errors"
	customErr "nps-config-service/internal/common/errors"
	"nps-config-service/internal/modules/config-manager/apis/dtos"
	"testing"
	mocks_service "nps-config-service/internal/modules/config-manager/services/mocks"
	"bytes"
	"encoding/json"
	"net/http/httptest"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock service implementing IWebhookService

func setupFiberWithHandler(h *WebhookHandler) *fiber.App {
	app := fiber.New()

	app.Post("/register", func(c *fiber.Ctx) error {
		var body map[string]interface{}
		if err := c.BodyParser(&body); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
		}
		c.Locals("contextData", &body)
		return h.RegisterWebhook(c)
	})

	app.Get("/webhooks/:environment/:service", h.GetWebhooks)

	app.Delete("/webhooks/:environment/:service", func(c *fiber.Ctx) error {
		var body map[string]interface{}
		if err := c.BodyParser(&body); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
		}
		c.Locals("contextData", &body)
		return h.DeleteWebhook(c)
	})

	return app
}

// --- TESTS ---

func TestRegisterWebhook_Success(t *testing.T) {
	mockService := new(mocks_service.MockWebhookService)
	handler := NewWebhookHandler(mockService)
	app := setupFiberWithHandler(handler)

	request := dtos.RegisterWebhookRequest{
		URL:         "http://example.com",
		Method:      "POST",
		ServiceName: "my-service",
		Environment: "dev",
	}

	mockService.On("RegisterWebhookService", request).Return(request, nil)

	body, _ := json.Marshal(request)

	req := httptest.NewRequest("POST", "/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	mockService.AssertExpectations(t)
}

func TestRegisterWebhook_ConflictError(t *testing.T) {
	mockService := new(mocks_service.MockWebhookService)
	handler := NewWebhookHandler(mockService)
	app := setupFiberWithHandler(handler)

	request := dtos.RegisterWebhookRequest{
		URL:    "http://example.com",
		Method: "POST",
	}

	mockService.On("RegisterWebhookService", request).
		Return(nil, &customErr.ConflictError{Message: "Webhook already exists"})

	body, _ := json.Marshal(request)

	req := httptest.NewRequest("POST", "/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, fiber.StatusConflict, resp.StatusCode)
	mockService.AssertExpectations(t)
}

func TestGetWebhooks_Success(t *testing.T) {
	mockService := new(mocks_service.MockWebhookService)
	handler := NewWebhookHandler(mockService)
	app := setupFiberWithHandler(handler)

	mockService.On("GetWebhooks", "dev", "my-service").
		Return([]dtos.RegisterWebhookRequest{
			{
				URL:         "http://example.com",
				Method:      "POST",
				Environment: "dev",
				ServiceName: "my-service",
			},
		}, nil)

	req := httptest.NewRequest("GET", "/webhooks/dev/my-service", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	mockService.AssertExpectations(t)
}

func TestGetWebhooks_NotFound(t *testing.T) {
	mockService := new(mocks_service.MockWebhookService)
	handler := NewWebhookHandler(mockService)
	app := setupFiberWithHandler(handler)

	mockService.On("GetWebhooks", "dev", "unknown-service").
		Return([]dtos.RegisterWebhookRequest{}, errors.New("service not found"))

	req := httptest.NewRequest("GET", "/webhooks/dev/unknown-service", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)
	mockService.AssertExpectations(t)
}

func TestDeleteWebhook_Success(t *testing.T) {
	mockService := new(mocks_service.MockWebhookService)
	handler := NewWebhookHandler(mockService)
	app := setupFiberWithHandler(handler)

	request := dtos.RegisterWebhookRequest{
		URL:    "http://example.com",
		Method: "POST",
	}

	mockService.On("DeleteWebhook", "dev", "my-service", request.URL, request.Method).
		Return("deleted", nil)

	body, _ := json.Marshal(request)

	req := httptest.NewRequest("DELETE", "/webhooks/dev/my-service", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	mockService.AssertExpectations(t)
}

func TestDeleteWebhook_NotFound(t *testing.T) {
	mockService := new(mocks_service.MockWebhookService)
	handler := NewWebhookHandler(mockService)
	app := setupFiberWithHandler(handler)

	request := dtos.RegisterWebhookRequest{
		URL:         "http://nonexistent.com",
		Method:      "POST",
		Environment: "dev",
		ServiceName: "my-service",
	}

	mockService.On("DeleteWebhook", "dev", "my-service", request.URL, request.Method).
		Return("", &customErr.ConflictError{Message: "Webhook not found"})

	body, _ := json.Marshal(request)

	req := httptest.NewRequest("DELETE", "/webhooks/dev/my-service", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)
	mockService.AssertExpectations(t)
}
