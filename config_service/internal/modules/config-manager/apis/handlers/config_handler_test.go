package handler_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"nps-config-service/internal/modules/config-manager/apis/dtos"
	handler "nps-config-service/internal/modules/config-manager/apis/handlers"
	"nps-config-service/internal/modules/config-manager/apis/routes"
	"nps-config-service/internal/modules/config-manager/services/mocks"
	"nps-config-service/pkg/observability"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestConfigHandler(t *testing.T) {
	confService := new(mocks.MockConfigService)

	confHandler := handler.NewConfigHandler(confService, observability.NewObservabilityStack("config-handler"))

	app := fiber.New()
	routes.RegisterConfigRoutes(app, confHandler)

	app.Put("/noctx/:environment/:service", confHandler.StoreConfigHandler)

	t.Run("NewConfigHandler should not panic with nil ObservabilityStack", func(t *testing.T) {
		confService.On("StoreConfigService", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&dtos.SuccessResponse{
			StatusCode: 200,
			Message:    "Success",
		}, nil)

		configsData := map[string]string{
			"conf1": "conf1value",
		}

		jsonData, err := json.Marshal(configsData)
		assert.NoError(t, err)
		assert.NotNil(t, jsonData)

		req := httptest.NewRequest(http.MethodPut, "/env/srv1", bytes.NewReader(jsonData))
		// _req := httptest.NewRequest(http.MethodGet, "/admin/signup", nil)
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("NewConfigHandler should not panic with nil ObservabilityStack", func(t *testing.T) {
		confService.On("StoreConfigService", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&dtos.SuccessResponse{
			StatusCode: 200,
			Message:    "Success",
		}, nil)

		configsData := map[string]string{
			"conf1": "conf1value",
		}

		jsonData, err := json.Marshal(configsData)
		assert.NoError(t, err)
		assert.NotNil(t, jsonData)

		req := httptest.NewRequest(http.MethodPut, "/noctx/env/srv1", bytes.NewReader(jsonData))
		// _req := httptest.NewRequest(http.MethodGet, "/admin/signup", nil)
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("NewConfigHandler should not panic with nil ObservabilityStack", func(t *testing.T) {
		confService := new(mocks.MockConfigService)

		confHandler := handler.NewConfigHandler(confService, observability.NewObservabilityStack("config-handler"))

		app := fiber.New()
		routes.RegisterConfigRoutes(app, confHandler)

		confService.On("StoreConfigService", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&dtos.SuccessResponse{
			StatusCode: 400,
			Message:    "Fail",
		}, &dtos.ServiceErrorResponse{
			ErrorCode:    "ERR100",
			ErrorMessage: "unknown error",
			StatusCode:   401,
		})

		configsData := map[string]string{
			"conf1": "conf1value",
		}

		jsonData, err := json.Marshal(configsData)
		assert.NoError(t, err)
		assert.NotNil(t, jsonData)

		req := httptest.NewRequest(http.MethodPut, "/env/srv1", bytes.NewReader(jsonData))
		// _req := httptest.NewRequest(http.MethodGet, "/admin/signup", nil)
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})

}

func TestConfigHandler_GetFullConfig(t *testing.T) {
	confService := new(mocks.MockConfigService)

	confHandler := handler.NewConfigHandler(confService, observability.NewObservabilityStack("config-handler"))

	app := fiber.New()
	routes.RegisterConfigRoutes(app, confHandler)

	t.Run("NewConfigHandler should not panic with nil ObservabilityStack", func(t *testing.T) {
		confService.On("GetConfigService", mock.Anything, mock.Anything, mock.Anything).Return(&dtos.SuccessResponse{
			StatusCode: 200,
			Message:    "Success",
		}, nil)

		confService.On("GetConfigValueService", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&dtos.SuccessResponse{
			StatusCode: 200,
			Message:    "Success",
		}, nil)

		req := httptest.NewRequest(http.MethodGet, "/env/srv1", nil)
		// _req := httptest.NewRequest(http.MethodGet, "/admin/signup", nil)
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("NewConfigHandler should not panic with nil ObservabilityStack", func(t *testing.T) {
		confService := new(mocks.MockConfigService)

		confHandler := handler.NewConfigHandler(confService, observability.NewObservabilityStack("config-handler"))

		app := fiber.New()
		routes.RegisterConfigRoutes(app, confHandler)

		confService.On("GetConfigService", mock.Anything, mock.Anything, mock.Anything).Return(&dtos.SuccessResponse{
			StatusCode: 400,
			Message:    "Fail",
		}, errors.New("sample error"))

		req := httptest.NewRequest(http.MethodGet, "/env/srv1", nil)
		// _req := httptest.NewRequest(http.MethodGet, "/admin/signup", nil)
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})

}

func TestConfigHandler_GetByVal(t *testing.T) {
	confService := new(mocks.MockConfigService)

	confHandler := handler.NewConfigHandler(confService, observability.NewObservabilityStack("config-handler"))

	app := fiber.New()

	routes.RegisterConfigRoutes(app, confHandler)

	t.Run("NewConfigHandler should not panic with nil ObservabilityStack", func(t *testing.T) {
		confService.On("GetConfigValueService", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&dtos.SuccessResponse{
			StatusCode: 200,
			Message:    "Success",
		}, nil)
		confService.On("GetConfigService", mock.Anything, mock.Anything, mock.Anything).Return(&dtos.SuccessResponse{
			StatusCode: 200,
			Message:    "Success",
		}, nil)

		req := httptest.NewRequest(http.MethodGet, "/env/srv1/val1", nil)
		// _req := httptest.NewRequest(http.MethodGet, "/admin/signup", nil)
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("NewConfigHandler should not panic with nil ObservabilityStack", func(t *testing.T) {
		confService := new(mocks.MockConfigService)

		confHandler := handler.NewConfigHandler(confService, observability.NewObservabilityStack("config-handler"))

		confService.On("GetConfigValueService", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&dtos.SuccessResponse{
			StatusCode: 200,
			Message:    "Success",
		}, errors.New("sample error"))
		app := fiber.New()
		routes.RegisterConfigRoutes(app, confHandler)

		req := httptest.NewRequest(http.MethodGet, "/env/srv1/val1", nil)
		// _req := httptest.NewRequest(http.MethodGet, "/admin/signup", nil)
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})

}

func TestInitConfigHandler(t *testing.T) {
	assert.Panicsf(t, func() {
		handler.NewConfigHandler(nil, nil)
	}, "obs error")
}
