package handler_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"nps-config-service/internal/configs"
	"nps-config-service/internal/modules/config-manager/apis/dtos"
	handler "nps-config-service/internal/modules/config-manager/apis/handlers"
	handlerMock "nps-config-service/internal/modules/config-manager/apis/handlers/mocks"
	"nps-config-service/internal/modules/config-manager/apis/routes"
	"nps-config-service/internal/modules/config-manager/services/mocks"
	"nps-config-service/pkg/observability"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAdminLoginHandler(t *testing.T) {
	adminService := new(mocks.MockAdminService)
	observabilityStack := observability.NewObservabilityStack("config_service")

	provider := handlerMock.NewMockConfigProvider("admin-secret", "jwt-secret")
	adminHandler := handler.NewAdminHandler(adminService, observabilityStack, provider)
	app := fiber.New()
	routes.RegisterAdminRoutes(app, adminHandler)

	app.Post("/admin/login-no-ctxdata", adminHandler.FetchAdminHandler)

	t.Run("NewAdminHandler with nil observability stack", func(t *testing.T) {
		adminService.On("FetchAdminService", mock.Anything, mock.Anything, mock.Anything).Return(&dtos.ResponseAdminDto{
			Success: true,
			Message: "Admin login successful",
			AdminId: "admin-id",
		}, nil).Once()

		adminLoginDto := &dtos.AdminLoginDto{
			Username: "admin",
			Password: "password",
		}

		jsonData, err := json.Marshal(adminLoginDto)
		assert.NoError(t, err)
		assert.NotNil(t, jsonData)

		req := httptest.NewRequest(http.MethodPost, "/admin/login", bytes.NewReader(jsonData))
		// _req := httptest.NewRequest(http.MethodGet, "/admin/signup", nil)
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("NewAdminHandler with nil observability stack", func(t *testing.T) {
		adminService.On("FetchAdminService", mock.Anything, mock.Anything, mock.Anything).Return(&dtos.ResponseAdminDto{
			Success: true,
			Message: "Admin login successful",
			AdminId: "admin-id",
		}, nil).Once()

		adminLoginDto := &dtos.AdminLoginDto{
			Username: "admin",
			Password: "password",
		}

		jsonData, err := json.Marshal(adminLoginDto)
		assert.NoError(t, err)
		assert.NotNil(t, jsonData)

		req := httptest.NewRequest(http.MethodPost, "/admin/login-no-ctxdata", bytes.NewReader(jsonData))
		// _req := httptest.NewRequest(http.MethodGet, "/admin/signup", nil)
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("NewAdminHandler with nil observability stack", func(t *testing.T) {
		adminService := new(mocks.MockAdminService)
		observabilityStack := observability.NewObservabilityStack("config_service")

		provider := handlerMock.NewMockConfigProvider("admin-secret", "")
		adminHandler := handler.NewAdminHandler(adminService, observabilityStack, provider)
		app := fiber.New()
		routes.RegisterAdminRoutes(app, adminHandler)

		adminService.On("FetchAdminService", mock.Anything, mock.Anything, mock.Anything).Return(&dtos.ResponseAdminDto{
			Success: true,
			Message: "Admin login successful",
			AdminId: "admin-id",
		}, nil).Once()

		adminLoginDto := &dtos.AdminLoginDto{
			Username: "admin",
			Password: "password",
		}

		jsonData, err := json.Marshal(adminLoginDto)
		assert.NoError(t, err)
		assert.NotNil(t, jsonData)

		req := httptest.NewRequest(http.MethodPost, "/admin/login", bytes.NewReader(jsonData))
		// _req := httptest.NewRequest(http.MethodGet, "/admin/signup", nil)
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})

	t.Run("NewAdminHandler with nil observability stack", func(t *testing.T) {
		adminService.On("FetchAdminService", mock.Anything, mock.Anything, mock.Anything).Return(&dtos.ResponseAdminDto{
			Success: true,
			Message: "Admin login successful",
			AdminId: "admin-id",
		}, nil).Once()

		adminLoginDto := &dtos.AdminLoginDto{
			Username: "",
			Password: "password",
		}

		jsonData, err := json.Marshal(adminLoginDto)
		assert.NoError(t, err)
		assert.NotNil(t, jsonData)

		req := httptest.NewRequest(http.MethodPost, "/admin/login", bytes.NewReader(jsonData))
		// _req := httptest.NewRequest(http.MethodGet, "/admin/signup", nil)
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})

	t.Run("NewAdminHandler with nil observability stack", func(t *testing.T) {
		adminService.On("FetchAdminService", mock.Anything, mock.Anything, mock.Anything).Return(&dtos.ResponseAdminDto{
			Success: true,
			Message: "Admin login successful",
			AdminId: "admin-id",
		}, nil).Once()

		adminLoginDto := &dtos.AdminLoginDto{
			Username: "user",
			Password: "",
		}

		jsonData, err := json.Marshal(adminLoginDto)
		assert.NoError(t, err)
		assert.NotNil(t, jsonData)

		req := httptest.NewRequest(http.MethodPost, "/admin/login", bytes.NewReader(jsonData))
		// _req := httptest.NewRequest(http.MethodGet, "/admin/signup", nil)
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})

	t.Run("NewAdminHandler with nil observability stack", func(t *testing.T) {
		adminService := new(mocks.MockAdminService)
		observabilityStack := observability.NewObservabilityStack("config_service")

		provider := handlerMock.NewMockConfigProvider("admin-secret", "1234")
		adminHandler := handler.NewAdminHandler(adminService, observabilityStack, provider)
		app := fiber.New()
		routes.RegisterAdminRoutes(app, adminHandler)

		adminService.On("FetchAdminService", mock.Anything, mock.Anything, mock.Anything).Return(&dtos.ResponseAdminDto{
			Success: false,
			Message: "Error occurred",
			AdminId: "no",
			Token:   "LoL",
		}, errors.New("sample error"))

		adminLoginDto := &dtos.AdminLoginDto{
			Username: "admin",
			Password: "password",
		}

		jsonData, err := json.Marshal(adminLoginDto)
		assert.NoError(t, err)
		assert.NotNil(t, jsonData)

		req := httptest.NewRequest(http.MethodPost, "/admin/login", bytes.NewReader(jsonData))
		// _req := httptest.NewRequest(http.MethodGet, "/admin/signup", nil)
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		assert.NoError(t, err)

		fmt.Println(resp.Body)
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})
}

func TestAdminSignupHandler(t *testing.T) {
	adminService := new(mocks.MockAdminService)
	observabilityStack := observability.NewObservabilityStack("config_service")

	provider := handlerMock.NewMockConfigProvider("admin-secret", "jwt-secret")
	adminHandler := handler.NewAdminHandler(adminService, observabilityStack, provider)
	app := fiber.New()
	routes.RegisterAdminRoutes(app, adminHandler)

	app.Post("/admin/signup-no-ctxdata", adminHandler.CreateAdminHandler)

	t.Run("NewAdminHandler with nil observability stack", func(t *testing.T) {
		// adminService.On("FetchAdminService", mock.Anything, mock.Anything, mock.Anything).Return(&dtos.ResponseAdminDto{
		// 	Success: true,
		// 	Message: "Admin login successful",
		// 	AdminId: "admin-id",
		// }, nil).Once()

		adminService.On("CreateAdminService", mock.Anything, mock.Anything).Return(&dtos.ResponseAdminSignupDto{
			Success: true,
			Message: "Signup Successful",
			AdminId: "admin-id",
		}, nil)

		adminSignup := &dtos.AdminSignupDto{
			Secret:   "admin-secret",
			Username: "admin",
			Password: "password",
		}

		jsonData, err := json.Marshal(adminSignup)
		assert.NoError(t, err)
		assert.NotNil(t, jsonData)

		req := httptest.NewRequest(http.MethodPost, "/admin/signup", bytes.NewReader(jsonData))
		// _req := httptest.NewRequest(http.MethodGet, "/admin/signup", nil)
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("NewAdminHandler with nil observability stack", func(t *testing.T) {
		// adminService.On("FetchAdminService", mock.Anything, mock.Anything, mock.Anything).Return(&dtos.ResponseAdminDto{
		// 	Success: true,
		// 	Message: "Admin login successful",
		// 	AdminId: "admin-id",
		// }, nil).Once()

		adminService.On("CreateAdminService", mock.Anything, mock.Anything).Return(&dtos.ResponseAdminSignupDto{
			Success: true,
			Message: "Signup Successful",
			AdminId: "admin-id",
		}, nil)

		adminSignup := &dtos.AdminSignupDto{
			Secret:   "admin-secret",
			Username: "",
			Password: "password",
		}

		jsonData, err := json.Marshal(adminSignup)
		assert.NoError(t, err)
		assert.NotNil(t, jsonData)

		req := httptest.NewRequest(http.MethodPost, "/admin/signup", bytes.NewReader(jsonData))
		// _req := httptest.NewRequest(http.MethodGet, "/admin/signup", nil)
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})

	t.Run("NewAdminHandler with nil observability stack", func(t *testing.T) {
		// adminService.On("FetchAdminService", mock.Anything, mock.Anything, mock.Anything).Return(&dtos.ResponseAdminDto{
		// 	Success: true,
		// 	Message: "Admin login successful",
		// 	AdminId: "admin-id",
		// }, nil).Once()

		adminService.On("CreateAdminService", mock.Anything, mock.Anything).Return(&dtos.ResponseAdminSignupDto{
			Success: true,
			Message: "Signup Successful",
			AdminId: "admin-id",
		}, nil)

		adminSignup := &dtos.AdminSignupDto{
			Secret:   "admin-secret-invalid",
			Username: "admin",
			Password: "password",
		}

		jsonData, err := json.Marshal(adminSignup)
		assert.NoError(t, err)
		assert.NotNil(t, jsonData)

		req := httptest.NewRequest(http.MethodPost, "/admin/signup", bytes.NewReader(jsonData))
		// _req := httptest.NewRequest(http.MethodGet, "/admin/signup", nil)
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})

	t.Run("NewAdminHandler with nil observability stack", func(t *testing.T) {
		adminService.On("CreateAdminService", mock.Anything, mock.Anything).Return(&dtos.ResponseAdminSignupDto{
			Success: true,
			Message: "Signup Successful",
			AdminId: "admin-id",
		}, nil)
		adminSignup := &dtos.AdminSignupDto{
			Secret:   "admin-secret",
			Username: "admin",
			Password: "password",
		}

		jsonData, err := json.Marshal(adminSignup)
		assert.NoError(t, err)
		assert.NotNil(t, jsonData)

		req := httptest.NewRequest(http.MethodPost, "/admin/signup-no-ctxdata", bytes.NewReader(jsonData))
		// _req := httptest.NewRequest(http.MethodGet, "/admin/signup", nil)
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("NewAdminHandler with nil observability stack", func(t *testing.T) {
		adminService := new(mocks.MockAdminService)
		observabilityStack := observability.NewObservabilityStack("config_service")

		provider := handlerMock.NewMockConfigProvider("admin-secret", "abc")
		adminHandler := handler.NewAdminHandler(adminService, observabilityStack, provider)
		app := fiber.New()
		routes.RegisterAdminRoutes(app, adminHandler)

		adminService.On("CreateAdminService", mock.Anything, mock.Anything).Return(&dtos.ResponseAdminSignupDto{
			Success: true,
			Message: "Signup Successful",
			AdminId: "admin-id",
		}, errors.New("abcd"))
		adminSignup := &dtos.AdminSignupDto{
			Secret:   "admin-secret",
			Username: "admin",
			Password: "password",
		}

		jsonData, err := json.Marshal(adminSignup)
		assert.NoError(t, err)
		assert.NotNil(t, jsonData)

		req := httptest.NewRequest(http.MethodPost, "/admin/signup", bytes.NewReader(jsonData))
		// _req := httptest.NewRequest(http.MethodGet, "/admin/signup", nil)
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("NewAdminHandler with nil observability stack", func(t *testing.T) {
		adminService := new(mocks.MockAdminService)
		observabilityStack := observability.NewObservabilityStack("config_service")

		provider := handlerMock.NewMockConfigProvider("", "abc")
		adminHandler := handler.NewAdminHandler(adminService, observabilityStack, provider)
		app := fiber.New()
		routes.RegisterAdminRoutes(app, adminHandler)

		adminService.On("CreateAdminService", mock.Anything, mock.Anything).Return(&dtos.ResponseAdminSignupDto{
			Success: true,
			Message: "Signup Successful",
			AdminId: "admin-id",
		}, errors.New("abcd"))
		adminSignup := &dtos.AdminSignupDto{
			Secret:   "admin-secret",
			Username: "admin",
			Password: "password",
		}

		jsonData, err := json.Marshal(adminSignup)
		assert.NoError(t, err)
		assert.NotNil(t, jsonData)

		req := httptest.NewRequest(http.MethodPost, "/admin/signup", bytes.NewReader(jsonData))
		// _req := httptest.NewRequest(http.MethodGet, "/admin/signup", nil)
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})

	t.Run("NewAdminHandler with nil observability stack", func(t *testing.T) {
		adminService.On("CreateAdminService", mock.Anything, mock.Anything).Return(&dtos.ResponseAdminSignupDto{
			Success: true,
			Message: "Signup Successful",
			AdminId: "admin-id",
		}, nil)
		adminSignup := &dtos.AdminSignupDto{
			Secret:   "",
			Username: "admin",
			Password: "password",
		}

		jsonData, err := json.Marshal(adminSignup)
		assert.NoError(t, err)
		assert.NotNil(t, jsonData)

		req := httptest.NewRequest(http.MethodPost, "/admin/signup", bytes.NewReader(jsonData))
		// _req := httptest.NewRequest(http.MethodGet, "/admin/signup", nil)
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})

	t.Run("NewAdminHandler with nil observability stack", func(t *testing.T) {
		adminService.On("CreateAdminService", mock.Anything, mock.Anything).Return(&dtos.ResponseAdminSignupDto{
			Success: true,
			Message: "Signup Successful",
			AdminId: "admin-id",
		}, nil)
		adminSignup := &dtos.AdminSignupDto{
			Secret:   "admin-secret",
			Username: "admin",
			Password: "",
		}

		jsonData, err := json.Marshal(adminSignup)
		assert.NoError(t, err)
		assert.NotNil(t, jsonData)

		req := httptest.NewRequest(http.MethodPost, "/admin/signup", bytes.NewReader(jsonData))
		// _req := httptest.NewRequest(http.MethodGet, "/admin/signup", nil)
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})
}

func TestInitWithNil(t *testing.T) {
	assert.Panicsf(t, func() {
		_ = handler.NewAdminHandler(nil, nil, nil)
	}, "NewAdminHandler should panic if ObservabilityStack is nil")

	handler := handler.NewAdminHandler(nil, observability.NewObservabilityStack("abc"), nil)

	assert.NotNil(t, handler, "NewAdminHandler should not panic if ObservabilityStack is provided")
}

func TestConfigProvider(t *testing.T) {
	provider := handler.DefaultConfigProvider{}

	configs.AppConfig.AdminSecret = "1"
	secret := provider.GetAdminSecret()
	assert.NotEmpty(t, secret)
	configs.AppConfig.JWTSecret = "1"
	jwtSecret := provider.GetJWTSecret()
	assert.NotEmpty(t, jwtSecret)
}
