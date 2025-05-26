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
	"os"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAdminLoginHandler(t *testing.T) {
	os.Setenv("BYPASS_MIDDLEWARE", "true")
	adminService := new(mocks.MockAdminService)
	observabilityStack := observability.NewObservabilityStack("config_service")

	provider := handlerMock.NewMockConfigProvider("admin-secret", "jwt-secret")
	adminHandler := handler.NewAdminHandler(adminService, observabilityStack, provider)
	app := fiber.New()
	routes.RegisterAdminRoutes(app, adminHandler)

	app.Post("/noctx/login", adminHandler.FetchAdminHandler)

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
			Success:      true,
			Message:      "Admin login successful",
			AdminId:      "admin-id",
			Role:         "admin",
			Token:        "sample-token",
			RefreshToken: "sample-refresh-token",
		}, nil).Once()

		adminLoginDto := &dtos.AdminLoginDto{
			Username: "admin",
			Password: "password",
		}

		jsonData, err := json.Marshal(adminLoginDto)
		assert.NoError(t, err)
		assert.NotNil(t, jsonData)

		req := httptest.NewRequest(http.MethodPost, "/noctx/login", bytes.NewReader(jsonData))
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
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
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
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
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
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
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
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})
}

func TestAdminSignupHandler(t *testing.T) {
	os.Setenv("BYPASS_MIDDLEWARE", "true")
	adminService := new(mocks.MockAdminService)
	observabilityStack := observability.NewObservabilityStack("config_service")

	provider := handlerMock.NewMockConfigProvider("admin-secret", "jwt-secret")
	adminHandler := handler.NewAdminHandler(adminService, observabilityStack, provider)
	app := fiber.New()
	routes.RegisterAdminRoutes(app, adminHandler)

	app.Post("/noctx/signup", adminHandler.CreateAdminHandler)

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
			Role:     "ADMIN",
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
			Role:     "ADMIN",
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
			Role:     "ADMIN",
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
		adminService.On("CreateAdminService", mock.Anything, mock.Anything).Return(&dtos.ResponseAdminSignupDto{
			Success: true,
			Message: "Signup Successful",
			AdminId: "admin-id",
		}, nil)
		adminSignup := &dtos.AdminSignupDto{
			Secret:   "admin-secret",
			Username: "admin",
			Password: "password",
			Role:     "ADMIN",
		}

		jsonData, err := json.Marshal(adminSignup)
		assert.NoError(t, err)
		assert.NotNil(t, jsonData)

		req := httptest.NewRequest(http.MethodPost, "/noctx/signup", bytes.NewReader(jsonData))
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
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
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
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
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
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

func TestAdminListAdminHandler(t *testing.T) {
	err := os.Setenv("BYPASS_MIDDLEWARE", "true")
	assert.NoError(t, err)
	adminService := new(mocks.MockAdminService)
	observabilityStack := observability.NewObservabilityStack("config_service")

	provider := handlerMock.NewMockConfigProvider("admin-secret", "jwt-secret")
	adminHandler := handler.NewAdminHandler(adminService, observabilityStack, provider)
	app := fiber.New()

	// Admin only routes
	app.Get("/custom/admins",
		func(c *fiber.Ctx) error {
			c.Locals("userRole", "ADMIN")
			c.Locals("username", "admin")
			return c.Next()
		},
		adminHandler.ListAdminsHandler)

	app.Delete("/custom/admins/:username",
		func(c *fiber.Ctx) error {
			c.Locals("userRole", "ADMIN")
			c.Locals("username", "admin")
			return c.Next()
		},
		adminHandler.DeleteAdminHandler)
	routes.RegisterAdminRoutes(app, adminHandler)

	adminService.On("ListAdminsService", mock.Anything).Return(&dtos.ResponseListAdminsDto{
		Success: true,
		Message: "Admins fetched successfully",
		Admins: []*dtos.Admin{
			{
				ID:        "admin-id-1",
				UserName:  "admin1",
				Role:      "ADMIN",
				CreatedAt: "2023-10-01T00:00:00Z",
				UpdatedAt: "2023-10-01T00:00:00Z",
			}},
	}, nil).Once()

	t.Run("NewAdminHandler with nil observability stack", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/custom/admins", nil)
		// _req := httptest.NewRequest(http.MethodGet, "/admin/signup", nil)
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("NewAdminHandler with nil observability stack", func(t *testing.T) {
		err := os.Setenv("BYPASS_MIDDLEWARE", "true")
		assert.NoError(t, err)
		adminService := new(mocks.MockAdminService)
		observabilityStack := observability.NewObservabilityStack("config_service")

		provider := handlerMock.NewMockConfigProvider("admin-secret", "jwt-secret")
		adminHandler := handler.NewAdminHandler(adminService, observabilityStack, provider)
		app := fiber.New()
		routes.RegisterAdminRoutes(app, adminHandler)
		app.Get("/custom/admins",
			func(c *fiber.Ctx) error {
				c.Locals("userRole", "ADMIN")
				c.Locals("username", "admin")
				return c.Next()
			},
			adminHandler.ListAdminsHandler)

		adminService.On("ListAdminsService", mock.Anything).Return(&dtos.ResponseListAdminsDto{
			Success: true,
			Message: "Admins fetched successfully",
			Admins: []*dtos.Admin{
				{
					ID:        "admin-id-1",
					UserName:  "admin1",
					Role:      "ADMIN",
					CreatedAt: "2023-10-01T00:00:00Z",
					UpdatedAt: "2023-10-01T00:00:00Z",
				}},
		}, errors.New("abc")).Once()
		req := httptest.NewRequest(http.MethodGet, "/custom/admins", nil)
		// _req := httptest.NewRequest(http.MethodGet, "/admin/signup", nil)
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})
}

func TestAdminDeleteAdminHandler(t *testing.T) {
	adminService := new(mocks.MockAdminService)
	observabilityStack := observability.NewObservabilityStack("config_service")

	provider := handlerMock.NewMockConfigProvider("admin-secret", "jwt-secret")
	adminHandler := handler.NewAdminHandler(adminService, observabilityStack, provider)
	app := fiber.New()

	adminService.On("DeleteAdminService", mock.Anything, "P4R4MR").Return(&dtos.ResponseDeleteAdminDto{
		Success:  true,
		Message:  "Admin deleted successfully",
		Username: "P4R4MR",
	}, nil)

	adminService.On("DeleteAdminService", mock.Anything, "ParamR").Return(&dtos.ResponseDeleteAdminDto{}, errors.New("unknown error"))

	app.Delete("/custom/admins/:username",
		func(c *fiber.Ctx) error {
			c.Locals("userRole", "ADMIN")
			c.Locals("username", "rai")
			return c.Next()
		},
		adminHandler.DeleteAdminHandler)
	routes.RegisterAdminRoutes(app, adminHandler)

	t.Run("NewAdminHandler with nil observability stack", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/custom/admins/P4R4MR", nil)
		// _req := httptest.NewRequest(http.MethodGet, "/admin/signup", nil)
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("NewAdminHandler with nil observability stack", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/custom/admins/ParamR", nil)
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadGateway, resp.StatusCode)
	})

	t.Run("NewAdminHandler with nil observability stack", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/custom/admins/rai", nil)
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
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
