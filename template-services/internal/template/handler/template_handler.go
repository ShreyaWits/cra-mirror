package handler

import (
	"log"
	"template-services/internal/models"
	errors "template-services/internal/pkg/errors"
	"template-services/internal/template/dto"
	"template-services/internal/template/service" // Ensure this import is correct

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
			Data:         nil,
		})
	}

	// Check if template with same name already exists
	existingTemplate, err := h.service.GetTemplate(c.Context(), "", req.Name, req.Channel, req.Language)
	if existingTemplate != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Success:      false,
			ErrorMessage: "Template name already exists",
			ErrorCode:    errors.TmpErrTemplateNameAlreadyExists,
			Data:         nil,
		})
	}

	// Convert to proto request
	protoReq := &models.Template{
		Name:     req.Name,
		Channel:  req.Channel,
		Language: req.Language,
		Content:  req.Content,
		IsActive: req.IsActive,
	}

	// Call service
	resp, err := h.service.CreateTemplate(c.Context(), protoReq)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Success:      false,
			ErrorMessage: errors.GetAppErrorMessage(errors.TmpErrTemplateCreate),
			ErrorCode:    errors.TmpErrTemplateCreate,
			Data:         nil,
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
	}

	// Call service
	resp, err := h.service.GetTemplate(c.Context(), "", req.Name, req.Channel, req.Language)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Success:      false,
			ErrorMessage: errors.GetAppErrorMessage(errors.TmpErrTemplateFetch),
			ErrorCode:    errors.TmpErrTemplateFetch,
			Data:         nil,
		})
	}

	return c.JSON(dto.GetTemplateResponse{
		Success: true,
		Message: "Template retrieved successfully",
		Data: dto.TemplateResponse{
			ID:        resp.ID.String(),
			Name:      resp.Name,
			Channel:   resp.Channel,
			Language:  resp.Language,
			Version:   int(resp.Version),
			IsActive:  resp.IsActive,
			Content:   resp.Content,
			CreatedAt: resp.CreatedAt,
			UpdatedAt: resp.UpdatedAt,
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
			Data:         nil,
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
			Data:         nil,
		})
	}

	// Update the fields
	existing.IsActive = *req.IsActive

	updatedTemplate, err := h.service.UpdateTemplate(c.Context(), existing)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Success:      false,
			ErrorMessage: errors.GetAppErrorMessage(errors.TmpErrTemplateUpdate),
			ErrorCode:    errors.TmpErrTemplateUpdate,
			Data:         nil,
		})
	}
	log.Printf("Updated template: %+v", updatedTemplate)
	return c.JSON(dto.UpdateTemplateResponse{
		Success: true,
		Message: "Template updated successfully",
		Data: dto.TemplateResponse{
			ID:        updatedTemplate.ID.String(),
			Name:      updatedTemplate.Name,
			Channel:   updatedTemplate.Channel,
			Language:  updatedTemplate.Language,
			Version:   int(updatedTemplate.Version),
			IsActive:  updatedTemplate.IsActive,
			Content:   updatedTemplate.Content,
			CreatedAt: updatedTemplate.CreatedAt,
			UpdatedAt: updatedTemplate.UpdatedAt,
		},
	})
}

// DeleteTemplate handles template deletion
func (h *TemplateHandler) DeleteTemplate(c *fiber.Ctx) error {
	var req dto.DeleteTemplateRequest
	log.Printf("Request body: %+v", req)
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Success:      false,
			ErrorMessage: errors.GetAppErrorMessage(errors.TmpErrInvalidRequestBody),
			ErrorCode:    errors.TmpErrInvalidRequestBody,
			Data:         nil,
		})
	}

	// Parse and validate UUID
	uid, err := uuid.Parse(req.TemplateID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Success:      false,
			ErrorMessage: errors.GetAppErrorMessage(errors.TmpErrUUIDParsing),
			ErrorCode:    errors.TmpErrUUIDParsing,
			Data:         nil,
		})
	}

	// Check if template exists
	existing, err := h.service.GetTemplateByID(c.Context(), uid)
	log.Printf("Existing template: %+v", existing)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{
			Success:      false,
			ErrorMessage: errors.GetAppErrorMessage(errors.TmpErrTemplateNotFound),
			ErrorCode:    errors.TmpErrTemplateNotFound,
			Data:         nil,
		})
	}

	// Call service to delete
	_, err = h.service.DeleteTemplate(c.Context(), existing.ID.String())

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Success:      false,
			ErrorMessage: errors.GetAppErrorMessage(errors.TmpErrTemplateDelete),
			ErrorCode:    errors.TmpErrTemplateDelete,
			Data:         nil,
		})
	}

	return c.JSON(dto.DeleteTemplateResponse{
		Success: true,
		Message: "Template deleted successfully",
	})
}
