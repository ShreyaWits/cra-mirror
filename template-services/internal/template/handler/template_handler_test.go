
package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"testing"
	"time"

	appErrors "template-services/internal/pkg/errors"
	"template-services/internal/models"
	"template-services/internal/template/dto"
	"template-services/internal/template/mocks"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupTestApp(handler *TemplateHandler) *fiber.App {
	app := fiber.New()
	app.Post("/template", handler.CreateTemplate)
	app.Put("/template", handler.UpdateTemplate)
	app.Delete("/template", handler.DeleteTemplate)
	app.Get("/template/:id", handler.GetTemplate)
	return app
}

func ptrBool(b bool) *bool { return &b }

// --- CreateTemplate ---

func TestCreateTemplate_InvalidBody(t *testing.T) {
	mockSvc := new(mocks.MockTemplateService)
	handler := NewTemplateHandler(mockSvc)
	app := setupTestApp(handler)

	req := httptest.NewRequest("POST", "/template", bytes.NewReader([]byte("{invalid json")))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

	var resBody dto.ErrorResponse
	json.NewDecoder(resp.Body).Decode(&resBody)
	assert.False(t, resBody.Success)
	assert.Equal(t, appErrors.TmpErrInvalidRequestBody, resBody.ErrorCode)
}

func TestCreateTemplate_DuplicateName(t *testing.T) {
	mockSvc := new(mocks.MockTemplateService)
	handler := NewTemplateHandler(mockSvc)
	app := setupTestApp(handler)

	reqBody := dto.CreateTemplateRequest{
		Name:     "TestTemplate",
		Channel:  "email",
		Language: "en",
		Content:  "Hello, {{name}}!",
		IsActive: true,
	}
	mockSvc.On("GetTemplate", mock.Anything, "", reqBody.Name, reqBody.Channel, reqBody.Language).Return(&models.Template{}, nil)

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/template", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)

	var resBody dto.ErrorResponse
	json.NewDecoder(resp.Body).Decode(&resBody)
	assert.False(t, resBody.Success)
	assert.Equal(t, appErrors.TmpErrTemplateNameAlreadyExists, resBody.ErrorCode)
	assert.Equal(t, "Template name already exists", resBody.ErrorMessage)
}

func TestCreateTemplate_ServiceError(t *testing.T) {
	mockSvc := new(mocks.MockTemplateService)
	handler := NewTemplateHandler(mockSvc)
	app := setupTestApp(handler)

	reqBody := dto.CreateTemplateRequest{
		Name:     "TestTemplate",
		Channel:  "email",
		Language: "en",
		Content:  "Hello, {{name}}!",
		IsActive: true,
	}
	mockSvc.On("GetTemplate", mock.Anything, "", reqBody.Name, reqBody.Channel, reqBody.Language).Return(nil, nil)
	mockSvc.On("CreateTemplate", mock.Anything, mock.AnythingOfType("*models.Template")).Return(nil, errors.New("service error"))

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/template", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)

	var resBody dto.ErrorResponse
	json.NewDecoder(resp.Body).Decode(&resBody)
	assert.False(t, resBody.Success)
	assert.Equal(t, appErrors.TmpErrTemplateCreate, resBody.ErrorCode)
}

func TestCreateTemplate_Success(t *testing.T) {
	mockSvc := new(mocks.MockTemplateService)
	handler := NewTemplateHandler(mockSvc)
	app := setupTestApp(handler)

	reqBody := dto.CreateTemplateRequest{
		Name:     "TestTemplate",
		Channel:  "email",
		Language: "en",
		Content:  "Hello, {{name}}!",
		IsActive: true,
	}
	createdID := uuid.New()
	mockSvc.On("GetTemplate", mock.Anything, "", reqBody.Name, reqBody.Channel, reqBody.Language).Return(nil, nil)
	mockSvc.On("CreateTemplate", mock.Anything, mock.AnythingOfType("*models.Template")).Return(&models.Template{ID: createdID}, nil)

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/template", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusCreated, resp.StatusCode)

	var resBody dto.CreateTemplateResponse
	json.NewDecoder(resp.Body).Decode(&resBody)
	assert.True(t, resBody.Success)
	assert.Equal(t, createdID.String(), resBody.Data.TemplateID)
	assert.Equal(t, "Template created successfully", resBody.Message)
}

// --- GetTemplate ---

func TestGetTemplate_InvalidUUIDParam(t *testing.T) {
	mockSvc := new(mocks.MockTemplateService)
	handler := NewTemplateHandler(mockSvc)
	app := setupTestApp(handler)

	req := httptest.NewRequest("GET", "/template/invalid-uuid", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

	var resBody dto.ErrorResponse
	json.NewDecoder(resp.Body).Decode(&resBody)
	assert.False(t, resBody.Success)
	assert.Equal(t, appErrors.TmpErrUUIDParsing, resBody.ErrorCode)
}

func TestGetTemplate_NotFound(t *testing.T) {
	mockSvc := new(mocks.MockTemplateService)
	handler := NewTemplateHandler(mockSvc)
	app := setupTestApp(handler)

	id := uuid.New()
	mockSvc.On("GetTemplate", mock.Anything, id.String(), "", "", "").Return(nil, errors.New("not found"))

	req := httptest.NewRequest("GET", "/template/"+id.String(), nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)

	var resBody dto.ErrorResponse
	json.NewDecoder(resp.Body).Decode(&resBody)
	assert.False(t, resBody.Success)
	assert.Equal(t, appErrors.TmpErrTemplateNotFound, resBody.ErrorCode)
}

func TestGetTemplate_Success(t *testing.T) {
	mockSvc := new(mocks.MockTemplateService)
	handler := NewTemplateHandler(mockSvc)
	app := setupTestApp(handler)

	id := uuid.New()
	now := time.Now()
	expected := &models.Template{
		ID:        id,
		Name:      "Test",
		Channel:   "email",
		Language:  "en",
		Version:   1,
		IsActive:  true,
		Content:   "content",
		CreatedAt: now,
		UpdatedAt: now,
	}
	mockSvc.On("GetTemplate", mock.Anything, id.String(), "", "", "").Return(expected, nil)

	req := httptest.NewRequest("GET", "/template/"+id.String(), nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var resBody dto.GetTemplateResponse
	json.NewDecoder(resp.Body).Decode(&resBody)
	assert.True(t, resBody.Success)
	assert.Equal(t, expected.ID.String(), resBody.Data.ID)
	assert.Equal(t, "Template retrieved successfully", resBody.Message)
}

// --- UpdateTemplate ---

func TestUpdateTemplate_InvalidBody(t *testing.T) {
	mockSvc := new(mocks.MockTemplateService)
	handler := NewTemplateHandler(mockSvc)
	app := setupTestApp(handler)

	req := httptest.NewRequest("PUT", "/template", bytes.NewReader([]byte("{invalid json")))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

	var resBody dto.ErrorResponse
	json.NewDecoder(resp.Body).Decode(&resBody)
	assert.False(t, resBody.Success)
	assert.Equal(t, appErrors.TmpErrInvalidRequestBody, resBody.ErrorCode)
}

func TestUpdateTemplate_InvalidUUID(t *testing.T) {
	mockSvc := new(mocks.MockTemplateService)
	handler := NewTemplateHandler(mockSvc)
	app := setupTestApp(handler)

	reqBody := dto.UpdateTemplateRequest{
		TemplateID: "not-a-uuid",
		IsActive:   ptrBool(true),
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("PUT", "/template", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

	var resBody dto.ErrorResponse
	json.NewDecoder(resp.Body).Decode(&resBody)
	assert.False(t, resBody.Success)
	assert.Equal(t, appErrors.TmpErrUUIDParsing, resBody.ErrorCode)
}

func TestUpdateTemplate_NotFound(t *testing.T) {
	mockSvc := new(mocks.MockTemplateService)
	handler := NewTemplateHandler(mockSvc)
	app := setupTestApp(handler)

	tmplID := uuid.New()
	reqBody := dto.UpdateTemplateRequest{
		TemplateID: tmplID.String(),
		IsActive:   ptrBool(true),
	}
	mockSvc.On("GetTemplateByID", mock.Anything, tmplID).Return(nil, errors.New("not found"))

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("PUT", "/template", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)

	var resBody dto.ErrorResponse
	json.NewDecoder(resp.Body).Decode(&resBody)
	assert.False(t, resBody.Success)
	assert.Equal(t, appErrors.TmpErrTemplateNotFound, resBody.ErrorCode)
}

func TestUpdateTemplate_ServiceError(t *testing.T) {
	mockSvc := new(mocks.MockTemplateService)
	handler := NewTemplateHandler(mockSvc)
	app := setupTestApp(handler)

	tmplID := uuid.New()
	reqBody := dto.UpdateTemplateRequest{
		TemplateID: tmplID.String(),
		IsActive:   ptrBool(true),
	}
	existing := &models.Template{
		ID:        tmplID,
		Name:      "TestTemplate",
		Channel:   "email",
		Language:  "en",
		Version:   1,
		IsActive:  false,
		Content:   "Old content",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	mockSvc.On("GetTemplateByID", mock.Anything, tmplID).Return(existing, nil)
	mockSvc.On("UpdateTemplate", mock.Anything, mock.AnythingOfType("*models.Template")).Return(nil, errors.New("update error"))

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("PUT", "/template", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)

	var resBody dto.ErrorResponse
	json.NewDecoder(resp.Body).Decode(&resBody)
	assert.False(t, resBody.Success)
	assert.Equal(t, appErrors.TmpErrTemplateUpdate, resBody.ErrorCode)
}

func TestUpdateTemplate_Success(t *testing.T) {
	mockSvc := new(mocks.MockTemplateService)
	handler := NewTemplateHandler(mockSvc)
	app := setupTestApp(handler)

	tmplID := uuid.New()
	reqBody := dto.UpdateTemplateRequest{
		TemplateID: tmplID.String(),
		IsActive:   ptrBool(true),
	}
	existing := &models.Template{
		ID:        tmplID,
		Name:      "TestTemplate",
		Channel:   "email",
		Language:  "en",
		Version:   1,
		IsActive:  false,
		Content:   "Old content",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	updated := *existing
	updated.IsActive = true

	mockSvc.On("GetTemplateByID", mock.Anything, tmplID).Return(existing, nil)
	mockSvc.On("UpdateTemplate", mock.Anything, mock.AnythingOfType("*models.Template")).Return(&updated, nil)

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("PUT", "/template", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var resBody dto.UpdateTemplateResponse
	json.NewDecoder(resp.Body).Decode(&resBody)
	assert.True(t, resBody.Success)
	assert.Equal(t, updated.ID.String(), resBody.Data.ID)
	assert.True(t, resBody.Data.IsActive)
	assert.Equal(t, "Template updated successfully", resBody.Message)
}

// --- DeleteTemplate ---

func TestDeleteTemplate_InvalidBody(t *testing.T) {
	mockSvc := new(mocks.MockTemplateService)
	handler := NewTemplateHandler(mockSvc)
	app := setupTestApp(handler)

	req := httptest.NewRequest("DELETE", "/template", bytes.NewReader([]byte("{invalid json")))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

	var resBody dto.ErrorResponse
	json.NewDecoder(resp.Body).Decode(&resBody)
	assert.False(t, resBody.Success)
	assert.Equal(t, appErrors.TmpErrInvalidRequestBody, resBody.ErrorCode)
}

func TestDeleteTemplate_InvalidUUID(t *testing.T) {
	mockSvc := new(mocks.MockTemplateService)
	handler := NewTemplateHandler(mockSvc)
	app := setupTestApp(handler)

	reqBody := dto.DeleteTemplateRequest{
		TemplateID: "not-a-uuid",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("DELETE", "/template", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

	var resBody dto.ErrorResponse
	json.NewDecoder(resp.Body).Decode(&resBody)
	assert.False(t, resBody.Success)
	assert.Equal(t, appErrors.TmpErrUUIDParsing, resBody.ErrorCode)
}

func TestDeleteTemplate_NotFound(t *testing.T) {
	mockSvc := new(mocks.MockTemplateService)
	handler := NewTemplateHandler(mockSvc)
	app := setupTestApp(handler)

	tmplID := uuid.New()
	reqBody := dto.DeleteTemplateRequest{
		TemplateID: tmplID.String(),
	}
	mockSvc.On("GetTemplateByID", mock.Anything, tmplID).Return(nil, errors.New("not found"))

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("DELETE", "/template", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)

	var resBody dto.ErrorResponse
	json.NewDecoder(resp.Body).Decode(&resBody)
	assert.False(t, resBody.Success)
	assert.Equal(t, appErrors.TmpErrTemplateNotFound, resBody.ErrorCode)
}

func TestDeleteTemplate_ServiceError(t *testing.T) {
	mockSvc := new(mocks.MockTemplateService)
	handler := NewTemplateHandler(mockSvc)
	app := setupTestApp(handler)

	tmplID := uuid.New()
	reqBody := dto.DeleteTemplateRequest{
		TemplateID: tmplID.String(),
	}
	existing := &models.Template{
		ID: tmplID,
	}
	mockSvc.On("GetTemplateByID", mock.Anything, tmplID).Return(existing, nil)
	mockSvc.On("DeleteTemplate", mock.Anything, tmplID.String()).Return(nil, errors.New("delete error"))

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("DELETE", "/template", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)

	var resBody dto.ErrorResponse
	json.NewDecoder(resp.Body).Decode(&resBody)
	assert.False(t, resBody.Success)
	assert.Equal(t, appErrors.TmpErrTemplateDelete, resBody.ErrorCode)
}

func TestDeleteTemplate_Success(t *testing.T) {
	mockSvc := new(mocks.MockTemplateService)
	handler := NewTemplateHandler(mockSvc)
	app := setupTestApp(handler)

	tmplID := uuid.New()
	reqBody := dto.DeleteTemplateRequest{
		TemplateID: tmplID.String(),
	}
	existing := &models.Template{
		ID: tmplID,
	}
	mockSvc.On("GetTemplateByID", mock.Anything, tmplID).Return(existing, nil)
	mockSvc.On("DeleteTemplate", mock.Anything, tmplID.String()).Return(existing, nil)

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("DELETE", "/template", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var resBody dto.DeleteTemplateResponse
	json.NewDecoder(resp.Body).Decode(&resBody)
	assert.True(t, resBody.Success)
	assert.Equal(t, "Template deleted successfully", resBody.Message)
}

// package handler

// import (
// 	"bytes"
// 	"encoding/json"
// 	"net/http/httptest"
// 	"testing"
// 	"time"
// 	"errors"
// 	appErrors "template-services/internal/pkg/errors"
// 	"template-services/internal/models"
// 	"template-services/internal/template/dto"
// 	"template-services/internal/template/mocks"
// 	"github.com/gofiber/fiber/v2"
// 	"github.com/google/uuid"
// 	"github.com/stretchr/testify/assert"
// 	"github.com/stretchr/testify/mock"
// )

// func setupTestApp(handler *TemplateHandler) *fiber.App {
// 	app := fiber.New()
// 	app.Post("/template", handler.CreateTemplate)
// 	app.Put("/template", handler.UpdateTemplate)
// 	app.Delete("/template", handler.DeleteTemplate)
// 	app.Get("/template/:id", handler.GetTemplate)
// 	return app
// }

// // --- Tests ---

// func TestCreateTemplate_ServiceError(t *testing.T) {
// 	mockSvc := new(mocks.MockTemplateService)
// 	handler := NewTemplateHandler(mockSvc)
// 	app := setupTestApp(handler)

// 	reqBody := dto.CreateTemplateRequest{
// 		Name:     "TestTemplate",
// 		Channel:  "email",
// 		Language: "en",
// 		Content:  "Hello, {{name}}!",
// 		IsActive: true,
// 	}
// 	mockSvc.On("GetTemplate", mock.Anything, "", reqBody.Name, reqBody.Channel, reqBody.Language).Return(nil, nil)
// 	mockSvc.On("CreateTemplate", mock.Anything, mock.AnythingOfType("*models.Template")).Return(nil, errors.New("service error"))

// 	body, _ := json.Marshal(reqBody)
// 	req := httptest.NewRequest("POST", "/template", bytes.NewReader(body))
// 	req.Header.Set("Content-Type", "application/json")
// 	resp, err := app.Test(req)
// 	assert.NoError(t, err)
// 	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)

// 	var resBody dto.ErrorResponse
// 	json.NewDecoder(resp.Body).Decode(&resBody)
// 	assert.False(t, resBody.Success)
// 	assert.Equal(t, appErrors.TmpErrTemplateCreate, resBody.ErrorCode)
// }

// func TestGetTemplate_InvalidUUID(t *testing.T) {
// 	mockSvc := new(mocks.MockTemplateService)
// 	handler := NewTemplateHandler(mockSvc)
// 	app := setupTestApp(handler)

// 	req := httptest.NewRequest("GET", "/template/invalid-uuid", nil)
// 	resp, err := app.Test(req)
// 	assert.NoError(t, err)
// 	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

// 	var resBody dto.ErrorResponse
// 	json.NewDecoder(resp.Body).Decode(&resBody)
// 	assert.False(t, resBody.Success)
// 	assert.Equal(t, appErrors.TmpErrUUIDParsing, resBody.ErrorCode)
// }

// func TestGetTemplate_NotFound(t *testing.T) {
// 	mockSvc := new(mocks.MockTemplateService)
// 	handler := NewTemplateHandler(mockSvc)
// 	app := setupTestApp(handler)

// 	id := uuid.New()
// 	mockSvc.On("GetTemplate", mock.Anything, id.String(), "", "", "").Return(nil, errors.New("not found"))

// 	req := httptest.NewRequest("GET", "/template/"+id.String(), nil)
// 	resp, err := app.Test(req)
// 	assert.NoError(t, err)
// 	assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)

// 	var resBody dto.ErrorResponse
// 	json.NewDecoder(resp.Body).Decode(&resBody)
// 	assert.False(t, resBody.Success)
// 	assert.Equal(t, appErrors.TmpErrTemplateNotFound, resBody.ErrorCode)
// }

// func TestGetTemplate_Success(t *testing.T) {
// 	mockSvc := new(mocks.MockTemplateService)
// 	handler := NewTemplateHandler(mockSvc)
// 	app := setupTestApp(handler)

// 	id := uuid.New()
// 	now := time.Now()
// 	expected := &models.Template{
// 		ID:        id,
// 		Name:      "Test",
// 		Channel:   "email",
// 		Language:  "en",
// 		Version:   1,
// 		IsActive:  true,
// 		Content:   "content",
// 		CreatedAt: now,
// 		UpdatedAt: now,
// 	}
// 	mockSvc.On("GetTemplate", mock.Anything, id.String(), "", "", "").Return(expected, nil)

// 	req := httptest.NewRequest("GET", "/template/"+id.String(), nil)
// 	resp, err := app.Test(req)
// 	assert.NoError(t, err)
// 	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

// 	var resBody dto.GetTemplateResponse
// 	json.NewDecoder(resp.Body).Decode(&resBody)
// 	assert.True(t, resBody.Success)
// 	assert.Equal(t, expected.ID.String(), resBody.Data.ID)
// }

// func TestUpdateTemplate_InvalidBody(t *testing.T) {
// 	mockSvc := new(mocks.MockTemplateService)
// 	handler := NewTemplateHandler(mockSvc)
// 	app := setupTestApp(handler)

// 	req := httptest.NewRequest("PUT", "/template", bytes.NewReader([]byte("{invalid json")))
// 	req.Header.Set("Content-Type", "application/json")
// 	resp, err := app.Test(req)
// 	assert.NoError(t, err)
// 	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

// 	var resBody dto.ErrorResponse
// 	json.NewDecoder(resp.Body).Decode(&resBody)
// 	assert.False(t, resBody.Success)
// 	assert.Equal(t, appErrors.TmpErrInvalidRequestBody, resBody.ErrorCode)
// }

// func TestUpdateTemplate_ServiceError(t *testing.T) {
// 	mockSvc := new(mocks.MockTemplateService)
// 	handler := NewTemplateHandler(mockSvc)
// 	app := setupTestApp(handler)

// 	tmplID := uuid.New()
// 	reqBody := dto.UpdateTemplateRequest{
// 		TemplateID: tmplID.String(),
// 		IsActive:   ptrBool(true),
// 	}
// 	existing := &models.Template{
// 		ID:        tmplID,
// 		Name:      "TestTemplate",
// 		Channel:   "email",
// 		Language:  "en",
// 		Version:   1,
// 		IsActive:  false,
// 		Content:   "Old content",
// 		CreatedAt: time.Now(),
// 		UpdatedAt: time.Now(),
// 	}
// 	mockSvc.On("GetTemplateByID", mock.Anything, tmplID).Return(existing, nil)
// 	mockSvc.On("UpdateTemplate", mock.Anything, mock.AnythingOfType("*models.Template")).Return(nil, errors.New("update error"))

// 	body, _ := json.Marshal(reqBody)
// 	req := httptest.NewRequest("PUT", "/template", bytes.NewReader(body))
// 	req.Header.Set("Content-Type", "application/json")
// 	resp, err := app.Test(req)
// 	assert.NoError(t, err)
// 	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)

// 	var resBody dto.ErrorResponse
// 	json.NewDecoder(resp.Body).Decode(&resBody)
// 	assert.False(t, resBody.Success)
// 	assert.Equal(t, appErrors.TmpErrTemplateUpdate, resBody.ErrorCode)
// }

// func TestDeleteTemplate_InvalidBody(t *testing.T) {
// 	mockSvc := new(mocks.MockTemplateService)
// 	handler := NewTemplateHandler(mockSvc)
// 	app := setupTestApp(handler)

// 	req := httptest.NewRequest("DELETE", "/template", bytes.NewReader([]byte("{invalid json")))
// 	req.Header.Set("Content-Type", "application/json")
// 	resp, err := app.Test(req)
// 	assert.NoError(t, err)
// 	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

// 	var resBody dto.ErrorResponse
// 	json.NewDecoder(resp.Body).Decode(&resBody)
// 	assert.False(t, resBody.Success)
// 	assert.Equal(t, appErrors.TmpErrInvalidRequestBody, resBody.ErrorCode)
// }

// func TestDeleteTemplate_ServiceError(t *testing.T) {
// 	mockSvc := new(mocks.MockTemplateService)
// 	handler := NewTemplateHandler(mockSvc)
// 	app := setupTestApp(handler)

// 	tmplID := uuid.New()
// 	reqBody := dto.DeleteTemplateRequest{
// 		TemplateID: tmplID.String(),
// 	}
// 	existing := &models.Template{
// 		ID: tmplID,
// 	}
// 	mockSvc.On("GetTemplateByID", mock.Anything, tmplID).Return(existing, nil)
// 	mockSvc.On("DeleteTemplate", mock.Anything, tmplID.String()).Return(nil, errors.New("delete error"))

// 	body, _ := json.Marshal(reqBody)
// 	req := httptest.NewRequest("DELETE", "/template", bytes.NewReader(body))
// 	req.Header.Set("Content-Type", "application/json")
// 	resp, err := app.Test(req)
// 	assert.NoError(t, err)
// 	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)

// 	var resBody dto.ErrorResponse
// 	json.NewDecoder(resp.Body).Decode(&resBody)
// 	assert.False(t, resBody.Success)
// 	assert.Equal(t, appErrors.TmpErrTemplateDelete, resBody.ErrorCode)
// }

// // --- Existing tests for success and other error branches are assumed to be present here ---

// func ptrBool(b bool) *bool {
// 	return &b
// }


