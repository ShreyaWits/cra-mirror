package handler

import (
	"template-services/internal/models"
	"template-services/internal/pkg/errors"
	"template-services/internal/template/dto"
	"template-services/internal/template/service"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type TemplateHandler struct {
	service service.TemplateServiceInterface
}

func NewTemplateHandler(service service.TemplateServiceInterface) *TemplateHandler {
	return &TemplateHandler{
		service: service,
	}
}

// CreateTemplate handles template creation
func (h *TemplateHandler) CreateTemplate(c *fiber.Ctx) error {
	var req dto.CreateTemplateRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Success:      false,
			ErrorMessage: errors.GetAppErrorMessage(errors.TmpErrInvalidRequestBody),
			ErrorCode:    errors.TmpErrInvalidRequestBody,
		})
	}

	// Convert to proto request
	protoReq := &models.Template{
		Name:           req.Name,
		Channel:        req.Channel,
		Language:       req.Language,
		Content:        req.Content,
		RequiredFields: req.RequiredFields,
		IsActive:       req.IsActive,
	}

	// Call service
	resp, err := h.service.CreateTemplate(c.Context(), protoReq)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Success:      false,
			ErrorMessage: errors.GetAppErrorMessage(errors.TmpErrTemplateCreate),
			ErrorCode:    errors.TmpErrTemplateCreate,
		})
	}

	return c.Status(fiber.StatusCreated).JSON(dto.CreateTemplateResponse{
		Success: true,
		Message: "Template created successfully",
		Data: struct {
			TemplateID string `json:"template_id"`
		}{
			TemplateID: resp.ID.String(),
		},
	})
}

// GetTemplate handles template retrieval
func (h *TemplateHandler) GetTemplate(c *fiber.Ctx) error {
	req := dto.GetTemplateRequest{
		Name:     c.Query("name"),
		Channel:  c.Query("channel"),
		Language: c.Query("language"),
		ID:       c.Params("id"),
	}

	// Call service
	resp, err := h.service.GetTemplate(c.Context(), req.ID, req.Name, req.Channel, req.Language)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Success:      false,
			ErrorMessage: errors.GetAppErrorMessage(errors.TmpErrTemplateFetch),
			ErrorCode:    errors.TmpErrTemplateFetch,
		})
	}

	return c.JSON(dto.GetTemplateResponse{
		Success: true,
		Message: "Template retrieved successfully",
		Data: dto.TemplateResponse{
			ID:             resp.ID.String(),
			Name:           resp.Name,
			Channel:        resp.Channel,
			Language:       resp.Language,
			Version:        int(resp.Version),
			IsActive:       resp.IsActive,
			Content:        resp.Content,
			RequiredFields: resp.RequiredFields,
			CreatedAt:      resp.CreatedAt.Format(time.RFC3339),
			UpdatedAt:      resp.UpdatedAt.Format(time.RFC3339),
		},
	})
}

// UpdateTemplate handles template updates
func (h *TemplateHandler) UpdateTemplate(c *fiber.Ctx) error {
	var req dto.UpdateTemplateRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Success:      false,
			ErrorMessage: errors.GetAppErrorMessage(errors.TmpErrInvalidRequestBody),
			ErrorCode:    errors.TmpErrInvalidRequestBody,
		})
	}

	idParam := c.Params("id")
	uid, err := uuid.Parse(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Success:      false,
			ErrorMessage: errors.GetAppErrorMessage(errors.TmpErrUUIDParsing),
			ErrorCode:    errors.TmpErrUUIDParsing,
		})
	}

	// Fetch existing template
	existing, err := h.service.GetTemplateByID(c.Context(), uid)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{
			Success:      false,
			ErrorMessage: errors.GetAppErrorMessage(errors.TmpErrTemplateNotFound),
			ErrorCode:    errors.TmpErrTemplateNotFound,
		})
	}

	// Update the fields
	existing.IsActive = req.IsActive
	if req.RequiredFields != nil {
		existing.RequiredFields = req.RequiredFields
	}

	updatedTemplate, err := h.service.UpdateTemplate(c.Context(), existing)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Success:      false,
			ErrorMessage: errors.GetAppErrorMessage(errors.TmpErrTemplateUpdate),
			ErrorCode:    errors.TmpErrTemplateUpdate,
		})
	}

	return c.JSON(dto.UpdateTemplateResponse{
		Success: true,
		Message: "Template updated successfully",
		Data: dto.TemplateResponse{
			ID:             updatedTemplate.ID.String(),
			Name:           updatedTemplate.Name,
			Channel:        updatedTemplate.Channel,
			Language:       updatedTemplate.Language,
			Version:        int(updatedTemplate.Version),
			IsActive:       updatedTemplate.IsActive,
			Content:        updatedTemplate.Content,
			RequiredFields: updatedTemplate.RequiredFields,
			CreatedAt:      updatedTemplate.CreatedAt.Format(time.RFC3339),
			UpdatedAt:      updatedTemplate.UpdatedAt.Format(time.RFC3339),
		},
	})
}

// DeleteTemplate handles template deletion
func (h *TemplateHandler) DeleteTemplate(c *fiber.Ctx) error {
	req := dto.DeleteTemplateRequest{
		ID: c.Params("id"),
	}

	// Call service
	_, err := h.service.DeleteTemplate(c.Context(), req.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Success:      false,
			ErrorMessage: errors.GetAppErrorMessage(errors.TmpErrTemplateDelete),
			ErrorCode:    errors.TmpErrTemplateDelete,
		})
	}

	return c.JSON(dto.DeleteTemplateResponse{
		Success: true,
		Message: "Template deleted successfully",
	})
}
