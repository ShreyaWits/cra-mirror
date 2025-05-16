package routes

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"template-services/internal/models"
	appErrors "template-services/internal/pkg/errors"
	"template-services/internal/template/dto"
	"template-services/internal/template/handler"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockTemplateService implements the service interface for testing
type MockTemplateService struct {
	mock.Mock
}

func (m *MockTemplateService) CreateTemplate(ctx context.Context, template *models.Template) (*models.Template, error) {
	args := m.Called(ctx, template)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Template), args.Error(1)
}

func (m *MockTemplateService) GetTemplate(ctx context.Context, id, name, channel, language string) (*models.Template, error) {
	args := m.Called(ctx, id, name, channel, language)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Template), args.Error(1)
}

func (m *MockTemplateService) GetTemplateByID(ctx context.Context, id uuid.UUID) (*models.Template, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Template), args.Error(1)
}

func (m *MockTemplateService) UpdateTemplate(ctx context.Context, template *models.Template) (*models.Template, error) {
	args := m.Called(ctx, template)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Template), args.Error(1)
}

func (m *MockTemplateService) DeleteTemplate(ctx context.Context, id string) (*models.Template, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Template), args.Error(1)
}

func (m *MockTemplateService) ListTemplates(ctx context.Context) ([]models.Template, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Template), args.Error(1)
}

func setupTestRouter() (*fiber.App, *handler.TemplateHandler, *MockTemplateService) {
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{
				"success":    false,
				"message":    err.Error(),
				"error_code": appErrors.TmpErrInvalidRequestBody,
			})
		},
	})

	mockService := new(MockTemplateService)
	templateHandler := handler.NewTemplateHandler(mockService)
	SetupTemplateRoutes(app, templateHandler)

	return app, templateHandler, mockService
}

func decodeJSONResponse(t *testing.T, resp *http.Response) fiber.Map {
	bodyBytes, _ := io.ReadAll(resp.Body)
	var result fiber.Map
	err := json.Unmarshal(bodyBytes, &result)
	assert.NoError(t, err, "Response body: %s", string(bodyBytes))
	return result
}

func TestSetupTemplateRoutes_CreateTemplate(t *testing.T) {
	app, _, mockService := setupTestRouter()

	t.Run("Valid Request", func(t *testing.T) {
		payload := dto.CreateTemplateRequest{
			Name:           "Test Template",
			Channel:        "email",
			Language:       "en",
			RequiredFields: []string{"first_name", "last_name"},
			Content:        "Hello, {{first_name}} {{last_name}}!",
			IsActive:       true,
		}
		body, _ := json.Marshal(payload)

		// Setup mock expectations with mock.MatchedBy
		mockService.On("GetTemplate", mock.Anything, "", payload.Name, payload.Channel, payload.Language).Return(nil, nil)
		mockService.On("CreateTemplate", mock.Anything, mock.MatchedBy(func(template *models.Template) bool {
			return template.Name == payload.Name &&
				template.Channel == payload.Channel &&
				template.Language == payload.Language &&
				template.Content == payload.Content &&
				template.IsActive == payload.IsActive
		})).Return(&models.Template{
			ID:       uuid.New(),
			Name:     payload.Name,
			Channel:  payload.Channel,
			Language: payload.Language,
			Content:  payload.Content,
			IsActive: payload.IsActive,
		}, nil)

		req := httptest.NewRequest(http.MethodPost, "/v1/templates/create", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusCreated, resp.StatusCode)

		result := decodeJSONResponse(t, resp)
		assert.True(t, result["success"].(bool))
		mockService.AssertExpectations(t)
	})

	t.Run("Invalid Request - Missing Required Fields", func(t *testing.T) {
		payload := dto.CreateTemplateRequest{} // Missing required fields
		body, _ := json.Marshal(payload)

		req := httptest.NewRequest(http.MethodPost, "/v1/templates/create", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		result := decodeJSONResponse(t, resp)
		assert.False(t, result["success"].(bool))
		assert.NotNil(t, result["message"])
		assert.Equal(t, appErrors.TmpErrInvalidRequestBody, result["error_code"])
	})

	t.Run("Invalid Method - GET instead of POST", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/templates/create", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusMethodNotAllowed, resp.StatusCode)

		result := decodeJSONResponse(t, resp)
		assert.False(t, result["success"].(bool))
		assert.Equal(t, "Method Not Allowed", result["message"])
	})
}

func TestSetupTemplateRoutes_GetTemplate(t *testing.T) {
	app, _, mockService := setupTestRouter()

	t.Run("Valid Request", func(t *testing.T) {
		payload := dto.GetTemplateRequest{
			Name:     "Test Template",
			Channel:  "email",
			Language: "en",
		}
		body, _ := json.Marshal(payload)

		// Setup mock expectations with mock.MatchedBy to properly match the context
		mockService.On("GetTemplate", mock.Anything, "", payload.Name, payload.Channel, payload.Language).Return(&models.Template{
			ID:       uuid.New(),
			Name:     payload.Name,
			Channel:  payload.Channel,
			Language: payload.Language,
		}, nil)

		req := httptest.NewRequest(http.MethodPost, "/v1/templates/get", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		result := decodeJSONResponse(t, resp)
		assert.True(t, result["success"].(bool))
		mockService.AssertExpectations(t)
	})

	t.Run("Invalid Request - Missing Required Fields", func(t *testing.T) {
		payload := dto.GetTemplateRequest{} // Missing required fields
		body, _ := json.Marshal(payload)

		req := httptest.NewRequest(http.MethodPost, "/v1/templates/get", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		result := decodeJSONResponse(t, resp)
		assert.False(t, result["success"].(bool))
		assert.NotNil(t, result["message"])
		assert.Equal(t, appErrors.TmpErrInvalidRequestBody, result["error_code"])
	})

	t.Run("Invalid Method - GET instead of POST", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/templates/get", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusMethodNotAllowed, resp.StatusCode)

		result := decodeJSONResponse(t, resp)
		assert.False(t, result["success"].(bool))
		assert.Equal(t, "Method Not Allowed", result["message"])
	})
}

func TestSetupTemplateRoutes_UpdateTemplate(t *testing.T) {
	app, _, mockService := setupTestRouter()

	tmplID := uuid.New()
	isActive := true

	t.Run("Valid Request", func(t *testing.T) {
		payload := dto.UpdateTemplateRequest{
			TemplateID: tmplID.String(),
			IsActive:   &isActive,
		}
		body, _ := json.Marshal(payload)

		// Setup mock expectations
		existingTemplate := &models.Template{
			ID:       tmplID,
			IsActive: false,
		}
		mockService.On("GetTemplateByID", mock.Anything, tmplID).Return(existingTemplate, nil)
		mockService.On("UpdateTemplate", mock.Anything, mock.MatchedBy(func(template *models.Template) bool {
			return template.ID == tmplID && template.IsActive == isActive
		})).Return(&models.Template{
			ID:       tmplID,
			IsActive: true,
		}, nil)

		req := httptest.NewRequest(http.MethodPost, "/v1/templates/update", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		result := decodeJSONResponse(t, resp)
		assert.True(t, result["success"].(bool))
		mockService.AssertExpectations(t)
	})

	t.Run("Invalid Request - Missing Template ID", func(t *testing.T) {
		payload := dto.UpdateTemplateRequest{
			IsActive: &isActive,
		}
		body, _ := json.Marshal(payload)

		req := httptest.NewRequest(http.MethodPost, "/v1/templates/update", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		result := decodeJSONResponse(t, resp)
		assert.False(t, result["success"].(bool))
		assert.NotNil(t, result["message"])
	})
}

func TestSetupTemplateRoutes_DeleteTemplate(t *testing.T) {
	app, _, mockService := setupTestRouter()

	tmplID := uuid.New()

	t.Run("Valid Request", func(t *testing.T) {
		payload := dto.DeleteTemplateRequest{
			TemplateID: tmplID.String(),
		}
		body, _ := json.Marshal(payload)

		// Setup mock expectations
		existingTemplate := &models.Template{
			ID: tmplID,
		}
		mockService.On("GetTemplateByID", mock.Anything, tmplID).Return(existingTemplate, nil)
		mockService.On("DeleteTemplate", mock.Anything, tmplID.String()).Return(existingTemplate, nil)

		req := httptest.NewRequest(http.MethodPost, "/v1/templates/delete", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		result := decodeJSONResponse(t, resp)
		assert.True(t, result["success"].(bool))
		mockService.AssertExpectations(t)
	})

	t.Run("Invalid Request - Missing Template ID", func(t *testing.T) {
		payload := dto.DeleteTemplateRequest{} // Missing template ID
		body, _ := json.Marshal(payload)

		req := httptest.NewRequest(http.MethodPost, "/v1/templates/delete", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		result := decodeJSONResponse(t, resp)
		assert.False(t, result["success"].(bool))
		assert.NotNil(t, result["message"])
	})
}
