package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"nps-config-service/internal/modules/config-manager/apis/dtos"
	handler "nps-config-service/internal/modules/config-manager/apis/handlers"
	"nps-config-service/internal/modules/config-manager/apis/routes"
	"nps-config-service/internal/modules/config-manager/services/mocks"
	"nps-config-service/pkg/observability"
	"os"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestWebHookHandler(t *testing.T) {
	wbService := new(mocks.MockWebhookService)

	wbHandler := handler.NewWebhookHandler(wbService, observability.NewObservabilityStack("webhook-handler"))

	app := fiber.New()
	routes.RegisterWebHookRoutes(app, wbHandler)

	wbService.On("RegisterWebhookService", mock.Anything, mock.Anything).Return(dtos.SuccessResponse{
		StatusCode: 200,
		Message:    "Webhook registered successfully",
	}, nil)

	app.Post("/noctx/webhook", wbHandler.RegisterWebhook)

	t.Run("wbHandler", func(t *testing.T) {

		registerWbRequest := dtos.RegisterWebhookRequest{
			URL:         "http://sampleurl.com/",
			Environment: "sampleenv",
			ServiceName: "sampleService",
			Method:      "POST",
		}

		jsonData, err := json.Marshal(registerWbRequest)
		assert.NoError(t, err)
		assert.NotNil(t, jsonData)

		req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(jsonData))
		// _req := httptest.NewRequest(http.MethodGet, "/admin/signup", nil)
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("wbHandler", func(t *testing.T) {

		registerWbRequest := dtos.RegisterWebhookRequest{
			URL:         "http://sampleurl.com/",
			Environment: "sampleenv",
			ServiceName: "sampleService",
			Method:      "POST",
		}

		jsonData, err := json.Marshal(registerWbRequest)
		assert.NoError(t, err)
		assert.NotNil(t, jsonData)

		req := httptest.NewRequest(http.MethodPost, "/noctx/webhook", bytes.NewReader(jsonData))
		// _req := httptest.NewRequest(http.MethodGet, "/admin/signup", nil)
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("wbHandler", func(t *testing.T) {
		wbService := new(mocks.MockWebhookService)

		wbHandler := handler.NewWebhookHandler(wbService, observability.NewObservabilityStack("webhook-handler"))

		app := fiber.New()
		routes.RegisterWebHookRoutes(app, wbHandler)

		wbService.On("RegisterWebhookService", mock.Anything, mock.Anything).Return(dtos.SuccessResponse{
			StatusCode: 400,
			Message:    "Webhook error",
		}, &dtos.ServiceErrorResponse{
			ErrorCode:    "ABC",
			ErrorMessage: "adkcndka",
		})

		registerWbRequest := dtos.RegisterWebhookRequest{
			URL:         "http://sampleurl.com/",
			Environment: "sampleenv",
			ServiceName: "sampleService",
			Method:      "POST",
		}

		jsonData, err := json.Marshal(registerWbRequest)
		assert.NoError(t, err)
		assert.NotNil(t, jsonData)

		req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(jsonData))
		// _req := httptest.NewRequest(http.MethodGet, "/admin/signup", nil)
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}

func TestWebHookHandler_GetWebhooks(t *testing.T) {
	wbService := new(mocks.MockWebhookService)

	wbHandler := handler.NewWebhookHandler(wbService, observability.NewObservabilityStack("webhook-handler"))

	app := fiber.New()
	routes.RegisterWebHookRoutes(app, wbHandler)

	wbService.On("GetWebhooks", mock.Anything, mock.Anything, mock.Anything).Return([]dtos.RegisterWebhookRequest{
		{
			URL:         "http://sampleurl.com/",
			Environment: "sampleenv",
			ServiceName: "sampleService",
			Method:      "POST",
		},
	}, nil)

	t.Run("wbHandler", func(t *testing.T) {
		registerWbRequest := dtos.RegisterWebhookRequest{
			URL:         "http://sampleurl.com/",
			Environment: "sampleenv",
			ServiceName: "sampleService",
			Method:      "POST",
		}

		jsonData, err := json.Marshal(registerWbRequest)
		assert.NoError(t, err)
		assert.NotNil(t, jsonData)

		req := httptest.NewRequest(http.MethodGet, "/env1/srv/webhooks", nil)
		// _req := httptest.NewRequest(http.MethodGet, "/admin/signup", nil)
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("wbHandler", func(t *testing.T) {
		wbService := new(mocks.MockWebhookService)
		wbHandler := handler.NewWebhookHandler(wbService, observability.NewObservabilityStack("webhook-handler"))

		wbService.On("GetWebhooks", mock.Anything, mock.Anything, mock.Anything).Return([]dtos.RegisterWebhookRequest{}, &dtos.ServiceErrorResponse{
			ErrorCode:    "ERR100",
			ErrorMessage: "No webhooks found",
			StatusCode:   http.StatusNotFound,
		})
		app := fiber.New()
		routes.RegisterWebHookRoutes(app, wbHandler)

		req := httptest.NewRequest(http.MethodGet, "/env1/srv/webhooks", nil)
		// _req := httptest.NewRequest(http.MethodGet, "/admin/signup", nil)
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})
}

func TestWebHookHandler_DeleteWebhooks(t *testing.T) {
	os.Setenv("BYPASS_MIDDLEWARE", "true")

	wbService := new(mocks.MockWebhookService)

	wbHandler := handler.NewWebhookHandler(wbService, observability.NewObservabilityStack("webhook-handler"))

	app := fiber.New()
	routes.RegisterWebHookRoutes(app, wbHandler)
	app.Delete("/noctx/:environment/:service/webhook", wbHandler.DeleteWebhook)

	wbService.On("DeleteWebhook", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return("OK", nil)

	t.Run("wbHandler", func(t *testing.T) {
		registerWbRequest := dtos.RegisterWebhookRequest{
			URL:         "http://sampleurl.com/",
			Environment: "sampleenv",
			ServiceName: "sampleService",
			Method:      "POST",
		}

		jsonData, err := json.Marshal(registerWbRequest)
		assert.NoError(t, err)
		assert.NotNil(t, jsonData)

		req := httptest.NewRequest(http.MethodDelete, "/env1/srv/webhook", bytes.NewReader(jsonData))
		// _req := httptest.NewRequest(http.MethodDelete, "/admin/signup", nil)
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("wbHandler", func(t *testing.T) {
		wbService := new(mocks.MockWebhookService)
		wbHandler := handler.NewWebhookHandler(wbService, observability.NewObservabilityStack("webhook-handler"))

		wbService.On("DeleteWebhook", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return("OK", &dtos.ServiceErrorResponse{
			ErrorCode:    "ERR100",
			ErrorMessage: "Webhook not found",
			StatusCode:   http.StatusNotFound,
		})
		app := fiber.New()
		routes.RegisterWebHookRoutes(app, wbHandler)

		registerWbRequest := dtos.RegisterWebhookRequest{
			URL:         "http://sampleurl.com/",
			Environment: "sampleenv",
			ServiceName: "sampleService",
			Method:      "POST",
		}

		jsonData, err := json.Marshal(registerWbRequest)
		assert.NoError(t, err)
		assert.NotNil(t, jsonData)

		req := httptest.NewRequest(http.MethodDelete, "/env1/srv/webhook", bytes.NewReader(jsonData))
		// _req := httptest.NewRequest(http.MethodDelete, "/admin/signup", nil)
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})
	t.Run("wbHandler", func(t *testing.T) {
		registerWbRequest := dtos.RegisterWebhookRequest{
			URL:         "http://sampleurl.com/",
			Environment: "sampleenv",
			ServiceName: "sampleService",
			Method:      "POST",
		}

		jsonData, err := json.Marshal(registerWbRequest)
		assert.NoError(t, err)
		assert.NotNil(t, jsonData)

		req := httptest.NewRequest(http.MethodDelete, "/noctx/env1/srv/webhook", bytes.NewReader(jsonData))
		// _req := httptest.NewRequest(http.MethodDelete, "/admin/signup", nil)
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("wbHandler", func(t *testing.T) {
		jsonData := []byte("abc")
		assert.NotNil(t, jsonData)

		req := httptest.NewRequest(http.MethodDelete, "/env1/srv/webhook", bytes.NewReader(jsonData))
		// _req := httptest.NewRequest(http.MethodDelete, "/admin/signup", nil)
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

}

func TestInitWebhookHandler(t *testing.T) {
	assert.Panicsf(t, func() {
		_ = handler.NewWebhookHandler(nil, nil)
	}, "InitWebhookHandler should panic with nil ObservabilityStack")
}
