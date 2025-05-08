package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"template-services/internal/models"
	"template-services/internal/template/dto"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// --- Mock service ---
type MockTemplateService struct {
	mock.Mock
}

func (m *MockTemplateService) CreateTemplate(ctx context.Context, tmpl *models.Template) (*models.Template, error) {
	args := m.Called(ctx, tmpl)
	return args.Get(0).(*models.Template), args.Error(1)
}

func (m *MockTemplateService) GetTemplate(ctx context.Context, id, name, channel, language string) (*models.Template, error) {
	args := m.Called(ctx, id, name, channel, language)
	return args.Get(0).(*models.Template), args.Error(1)
}

func (m *MockTemplateService) GetTemplateByID(ctx context.Context, id uuid.UUID) (*models.Template, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*models.Template), args.Error(1)
}

func (m *MockTemplateService) UpdateTemplate(ctx context.Context, tmpl *models.Template) (*models.Template, error) {
	args := m.Called(ctx, tmpl)
	return args.Get(0).(*models.Template), args.Error(1)
}

func (m *MockTemplateService) DeleteTemplate(ctx context.Context, id string) (*models.Template, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*models.Template), args.Error(1)
}

// --- Tests ---
func TestCreateTemplate_Success(t *testing.T) {
	app := fiber.New()
	mockService := new(MockTemplateService)
	h := NewTemplateHandler(mockService)

	app.Post("/template", h.CreateTemplate)

	reqBody := dto.CreateTemplateRequest{
		Name:     "Welcome",
		Channel:  "email",
		Language: "en",
		Content:  "Hello!",
		IsActive: true,
	}
	reqBytes, _ := json.Marshal(reqBody)

	createdID := uuid.New()
	mockResp := &models.Template{
		ID:        createdID,
		Name:      reqBody.Name,
		Channel:   reqBody.Channel,
		Language:  reqBody.Language,
		Content:   reqBody.Content,
		IsActive:  reqBody.IsActive,
		Version:   1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	mockService.On("CreateTemplate", mock.Anything, mock.AnythingOfType("*models.Template")).
		Return(mockResp, nil)

	req := httptest.NewRequest("POST", "/template", bytes.NewReader(reqBytes))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusCreated, resp.StatusCode)
}

func TestCreateTemplate_InvalidBody(t *testing.T) {
	app := fiber.New()
	mockService := new(MockTemplateService)
	h := NewTemplateHandler(mockService)

	app.Post("/template", h.CreateTemplate)

	req := httptest.NewRequest("POST", "/template", bytes.NewBufferString("{invalid json"))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestGetTemplate_NotFound(t *testing.T) {
	app := fiber.New()
	mockService := new(MockTemplateService)
	h := NewTemplateHandler(mockService)

	app.Get("/template/:id", h.GetTemplate)

	mockService.On("GetTemplate", mock.Anything, "invalid", "", "", "").
		Return(nil, errors.New("not found"))

	req := httptest.NewRequest("GET", "/template/invalid", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
}
