package handler

import (
	"context"
	"fmt"
	"log"
	configEnv "template-services/internal/configs"
	"template-services/internal/models"
	"template-services/internal/template/dto"
	"template-services/internal/template/service"
	"template-services/pkg/errors"
	"template-services/pkg/observability"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

const (
	defaultTimeout = 5 * time.Second
)

type TemplateHandler struct {
	configService *service.ConfigServiceImpl
	service       service.TemplateServiceInterface
	obs           *observability.ObservabilityStack
}

func NewTemplateHandler(service service.TemplateServiceInterface, obs *observability.ObservabilityStack, configSrv *service.ConfigServiceImpl) *TemplateHandler {
	return &TemplateHandler{
		service:       service,
		obs:           obs,
		configService: configSrv,
	}

}

func (h *TemplateHandler) HandleConfigWebhookChange(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.Context(), defaultTimeout)
	defer cancel()

	functionName := "HandleConfigWebhookChange"
	functionFailed := "HandleConfigWebhookChange_Failed"

	tCtx, span := h.obs.TracerService.StartTracer(ctx, functionName)
	defer h.obs.TracerService.StopSpan(span)
	h.obs.MetricsService.IncrementCounter(tCtx, functionName, 1, map[string]string{})

	webhookConfigData := configEnv.ConfigServiceWebhookData{}
	if err := c.BodyParser(&webhookConfigData); err != nil {
		h.obs.LoggerService.Error(tCtx, "Failed to parse webhook config data", map[string]interface{}{
			"error": err.Error(),
		})
		h.obs.MetricsService.IncrementCounter(tCtx, functionFailed, 1, map[string]string{
			"error": errors.GetAppErrorMessage(errors.TmpErrInvalidRequestBody),
		})
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Success:      false,
			ErrorMessage: errors.GetAppErrorMessage(errors.TmpErrInvalidRequestBody),
			ErrorCode:    errors.TmpErrInvalidRequestBody,
			Data:         nil,
		})
	}

	err := h.configService.UpdateConfig(tCtx, &webhookConfigData.Values)

	if err != nil {
		fmt.Println(err)
		h.obs.LoggerService.Error(tCtx, "Failed to update config", map[string]interface{}{
			"error": err.Error(),
		})
		h.obs.MetricsService.IncrementCounter(tCtx, functionFailed, 1, map[string]string{
			"error": "config_update_failed",
		})
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Success:      false,
			ErrorMessage: "Failed to update config",
			ErrorCode:    "CFG100",
			Data:         nil,
		})
	}
	h.obs.LoggerService.Info(tCtx, "Webhook config data received", map[string]interface{}{
		"config_data": webhookConfigData,
	})
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"Success": true,
		"Message": "Webhook config data processed successfully",
		"Data":    nil,
	})
}

// CreateTemplate handles template creation
func (h *TemplateHandler) CreateTemplate(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.Context(), defaultTimeout)
	defer cancel()
	// Get request ID from context or generate new one
	// requestID := getRequestID(ctx)
	functionName := "CreateTemplate"
	functionFailed := "CreateTemplate_Failed"

	// Create timeout context

	tCtx, span := h.obs.TracerService.StartTracer(ctx, functionName)
	defer h.obs.TracerService.StopSpan(span)
	h.obs.MetricsService.IncrementCounter(tCtx, functionName, 1, map[string]string{})

	// Validate request
	var req dto.CreateTemplateRequest
	if err := c.BodyParser(&req); err != nil {
		h.obs.LoggerService.Error(tCtx, err)
		h.obs.MetricsService.IncrementCounter(tCtx, functionFailed, 1, map[string]string{"error": errors.GetAppErrorMessage(errors.TmpErrInvalidRequestBody)})
		return c.Status(fiber.StatusAccepted).JSON(dto.ErrorResponse{
			Success:      false,
			ErrorMessage: errors.GetAppErrorMessage(errors.TmpErrInvalidRequestBody),
			ErrorCode:    errors.TmpErrInvalidRequestBody,
			Data:         nil,
		})
	}

	// Validate required fields
	if req.Name == "" || req.Channel == "" || req.Language == "" || req.Content == "" {
		h.obs.LoggerService.Error(tCtx, errors.GetAppErrorMessage(errors.TmpErrInvalidRequestBody))
		h.obs.MetricsService.IncrementCounter(tCtx, functionFailed, 1, map[string]string{"error": errors.GetAppErrorMessage(errors.TmpErrInvalidRequestBody)})
		return c.Status(fiber.StatusAccepted).JSON(dto.ErrorResponse{
			Success:      false,
			ErrorMessage: "Missing required fields",
			ErrorCode:    errors.TmpErrInvalidRequestBody,
			Data:         nil,
		})
	}

	// Check if template with same name already exists
	existingTemplate, _ := h.service.GetTemplate(tCtx, "", req.Name, req.Channel, req.Language)
	if existingTemplate != nil {
		h.obs.LoggerService.Error(tCtx, errors.GetAppErrorMessage(errors.TmpErrTemplateNameAlreadyExists))
		h.obs.MetricsService.IncrementCounter(tCtx, functionFailed, 1, map[string]string{"error": errors.GetAppErrorMessage(errors.TmpErrTemplateNameAlreadyExists)})
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Success:      false,
			ErrorMessage: "Template name already exists",
			ErrorCode:    errors.TmpErrTemplateNameAlreadyExists,
			Data:         nil,
		})

	}

	// Convert to proto request
	protoReq := &models.Template{
		Name:           req.Name,
		Channel:        req.Channel,
		Language:       req.Language,
		Content:        req.Content,
		IsActive:       req.IsActive,
		RequiredFields: pq.StringArray(req.RequiredFields),
	}

	// Call service
	resp, err := h.service.CreateTemplate(tCtx, protoReq)
	if err != nil {
		h.obs.LoggerService.Error(tCtx, errors.GetAppErrorMessage(errors.TmpErrTemplateCreate))
		h.obs.MetricsService.IncrementCounter(tCtx, functionFailed, 1, map[string]string{"error": errors.GetAppErrorMessage(errors.TmpErrTemplateCreate)})
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Success:      false,
			ErrorMessage: errors.GetAppErrorMessage(errors.TmpErrTemplateCreate),
			ErrorCode:    errors.TmpErrTemplateCreate,
			Data:         nil,
		})
	}
	h.obs.LoggerService.Info(tCtx, "Template created successfully")

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
	ctx, cancel := context.WithTimeout(c.Context(), defaultTimeout)
	defer cancel()

	functionName := "GetTemplate"
	functionFailed := "GetTemplate_Failed"

	tCtx, span := h.obs.TracerService.StartTracer(ctx, functionName)
	defer h.obs.TracerService.StopSpan(span)
	h.obs.MetricsService.IncrementCounter(tCtx, functionName, 1, map[string]string{})

	var req dto.GetTemplateRequest
	if err := c.BodyParser(&req); err != nil {
		h.obs.LoggerService.Error(tCtx, "Failed to parse request body", map[string]interface{}{
			"error": err.Error(),
		})
		h.obs.MetricsService.IncrementCounter(tCtx, functionFailed, 1, map[string]string{
			"error": errors.GetAppErrorMessage(errors.TmpErrInvalidRequestBody),
		})
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Success:      false,
			ErrorMessage: errors.GetAppErrorMessage(errors.TmpErrInvalidRequestBody),
			ErrorCode:    errors.TmpErrInvalidRequestBody,
			Data:         nil,
		})
	}

	h.obs.LoggerService.Info(tCtx, "Parsed GetTemplate request", map[string]interface{}{
		"name":     req.Name,
		"channel":  req.Channel,
		"language": req.Language,
	})

	resp, err := h.service.GetTemplate(c.Context(), "", req.Name, req.Channel, req.Language)
	log.Printf("GetTemplate response: %+v", resp)
	if err != nil {
		h.obs.LoggerService.Error(tCtx, "Failed to get template from service", map[string]interface{}{
			"error":    err.Error(),
			"name":     req.Name,
			"channel":  req.Channel,
			"language": req.Language,
		})
		h.obs.MetricsService.IncrementCounter(tCtx, functionFailed, 1, map[string]string{
			"error": errors.GetAppErrorMessage(errors.TmpErrTemplateNotFound),
		})
		return c.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{
			Success:      false,
			ErrorMessage: errors.GetAppErrorMessage(errors.TmpErrTemplateNotFound),
			ErrorCode:    errors.TmpErrTemplateNotFound,
			Data:         nil,
		})
	}

	h.obs.LoggerService.Info(tCtx, "Template retrieved successfully", map[string]interface{}{
		"id":        resp.ID.String(),
		"name":      resp.Name,
		"channel":   resp.Channel,
		"language":  resp.Language,
		"version":   resp.Version,
		"is_active": resp.IsActive,
	})

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
			RequiredFields: pq.StringArray(resp.RequiredFields),
			CreatedAt:      resp.CreatedAt,
			UpdatedAt:      resp.UpdatedAt,
		},
	})
}

// UpdateTemplate handles template updates
func (h *TemplateHandler) UpdateTemplate(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.Context(), defaultTimeout)
	defer cancel()

	functionName := "UpdateTemplate"
	functionFailed := "UpdateTemplate_Failed"

	tCtx, span := h.obs.TracerService.StartTracer(ctx, functionName)
	defer h.obs.TracerService.StopSpan(span)
	h.obs.MetricsService.IncrementCounter(tCtx, functionName, 1, map[string]string{})

	var req dto.UpdateTemplateRequest
	if err := c.BodyParser(&req); err != nil {
		h.obs.LoggerService.Error(tCtx, "Failed to parse update template request", map[string]interface{}{
			"error": err.Error(),
		})
		h.obs.MetricsService.IncrementCounter(tCtx, functionFailed, 1, map[string]string{
			"error": errors.GetAppErrorMessage(errors.TmpErrInvalidRequestBody),
		})
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Success:      false,
			ErrorMessage: errors.GetAppErrorMessage(errors.TmpErrInvalidRequestBody),
			ErrorCode:    errors.TmpErrInvalidRequestBody,
			Data:         nil,
		})
	}

	h.obs.LoggerService.Info(tCtx, "Received update request", map[string]interface{}{
		"template_id": req.TemplateID,
		"is_active":   req.IsActive,
	})

	uid, err := uuid.Parse(req.TemplateID)
	if err != nil {
		h.obs.LoggerService.Error(tCtx, "Failed to parse UUID", map[string]interface{}{
			"template_id": req.TemplateID,
			"error":       err.Error(),
		})
		h.obs.MetricsService.IncrementCounter(tCtx, functionFailed, 1, map[string]string{
			"error": errors.GetAppErrorMessage(errors.TmpErrUUIDParsing),
		})
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Success:      false,
			ErrorMessage: errors.GetAppErrorMessage(errors.TmpErrUUIDParsing),
			ErrorCode:    errors.TmpErrUUIDParsing,
		})
	}

	existing, err := h.service.GetTemplateByID(c.Context(), uid)
	if err != nil {
		h.obs.LoggerService.Error(tCtx, "Template not found for update", map[string]interface{}{
			"template_id": uid.String(),
			"error":       err.Error(),
		})
		h.obs.MetricsService.IncrementCounter(tCtx, functionFailed, 1, map[string]string{
			"error": errors.GetAppErrorMessage(errors.TmpErrTemplateNotFound),
		})
		return c.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{
			Success:      false,
			ErrorMessage: errors.GetAppErrorMessage(errors.TmpErrTemplateNotFound),
			ErrorCode:    errors.TmpErrTemplateNotFound,
			Data:         nil,
		})
	}

	h.obs.LoggerService.Info(tCtx, "Fetched template to update", map[string]interface{}{
		"template_id":           existing.ID.String(),
		"current_active_status": existing.IsActive,
	})

	existing.IsActive = *req.IsActive

	updatedTemplate, err := h.service.UpdateTemplate(c.Context(), existing)
	if err != nil {
		h.obs.LoggerService.Error(tCtx, "Failed to update template", map[string]interface{}{
			"template_id": uid.String(),
			"error":       err.Error(),
		})
		h.obs.MetricsService.IncrementCounter(tCtx, functionFailed, 1, map[string]string{
			"error": errors.GetAppErrorMessage(errors.TmpErrTemplateUpdate),
		})
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Success:      false,
			ErrorMessage: errors.GetAppErrorMessage(errors.TmpErrTemplateUpdate),
			ErrorCode:    errors.TmpErrTemplateUpdate,
			Data:         nil,
		})
	}

	h.obs.LoggerService.Info(tCtx, "Template updated successfully", map[string]interface{}{
		"template_id": updatedTemplate.ID.String(),
		"is_active":   updatedTemplate.IsActive,
		"version":     updatedTemplate.Version,
	})

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
	ctx, cancel := context.WithTimeout(c.Context(), defaultTimeout)
	defer cancel()

	functionName := "DeleteTemplate"
	functionFailed := "DeleteTemplate_Failed"

	tCtx, span := h.obs.TracerService.StartTracer(ctx, functionName)
	defer h.obs.TracerService.StopSpan(span)

	h.obs.MetricsService.IncrementCounter(tCtx, functionName, 1, map[string]string{})
	h.obs.LoggerService.Info(tCtx, "Received delete template request")

	var req dto.DeleteTemplateRequest
	if err := c.BodyParser(&req); err != nil {
		h.obs.LoggerService.Error(tCtx, "Failed to parse delete template request", map[string]interface{}{
			"error": err.Error(),
		})
		h.obs.MetricsService.IncrementCounter(tCtx, functionFailed, 1, map[string]string{
			"error": errors.GetAppErrorMessage(errors.TmpErrInvalidRequestBody),
		})
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Success:      false,
			ErrorMessage: errors.GetAppErrorMessage(errors.TmpErrInvalidRequestBody),
			ErrorCode:    errors.TmpErrInvalidRequestBody,
			Data:         nil,
		})
	}

	h.obs.LoggerService.Info(tCtx, "Delete request parsed", map[string]interface{}{
		"template_id": req.TemplateID,
	})

	uid, err := uuid.Parse(req.TemplateID)
	if err != nil {
		h.obs.LoggerService.Error(tCtx, "Invalid UUID format", map[string]interface{}{
			"template_id": req.TemplateID,
			"error":       err.Error(),
		})
		h.obs.MetricsService.IncrementCounter(tCtx, functionFailed, 1, map[string]string{
			"error": errors.GetAppErrorMessage(errors.TmpErrUUIDParsing),
		})
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Success:      false,
			ErrorMessage: errors.GetAppErrorMessage(errors.TmpErrUUIDParsing),
			ErrorCode:    errors.TmpErrUUIDParsing,
			Data:         nil,
		})
	}

	existing, err := h.service.GetTemplateByID(c.Context(), uid)
	if err != nil {
		h.obs.LoggerService.Error(tCtx, "Template not found for deletion", map[string]interface{}{
			"template_id": uid.String(),
			"error":       err.Error(),
		})
		h.obs.MetricsService.IncrementCounter(tCtx, functionFailed, 1, map[string]string{
			"error": errors.GetAppErrorMessage(errors.TmpErrTemplateNotFound),
		})
		return c.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{
			Success:      false,
			ErrorMessage: errors.GetAppErrorMessage(errors.TmpErrTemplateNotFound),
			ErrorCode:    errors.TmpErrTemplateNotFound,
			Data:         nil,
		})
	}

	h.obs.LoggerService.Info(tCtx, "Template found for deletion", map[string]interface{}{
		"template_id": existing.ID.String(),
	})

	_, err = h.service.DeleteTemplate(c.Context(), existing.ID.String())
	if err != nil {
		h.obs.LoggerService.Error(tCtx, "Failed to delete template", map[string]interface{}{
			"template_id": existing.ID.String(),
			"error":       err.Error(),
		})
		h.obs.MetricsService.IncrementCounter(tCtx, functionFailed, 1, map[string]string{
			"error": errors.GetAppErrorMessage(errors.TmpErrTemplateDelete),
		})
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Success:      false,
			ErrorMessage: errors.GetAppErrorMessage(errors.TmpErrTemplateDelete),
			ErrorCode:    errors.TmpErrTemplateDelete,
			Data:         nil,
		})
	}

	h.obs.LoggerService.Info(tCtx, "Template deleted successfully", map[string]interface{}{
		"template_id": existing.ID.String(),
	})

	return c.JSON(dto.DeleteTemplateResponse{
		Success: true,
		Message: "Template deleted successfully",
	})
}

func (h *TemplateHandler) ListTemplates(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.Context(), defaultTimeout)
	defer cancel()

	functionName := "ListTemplate"
	functionFailed := "ListTemplate_Failed"

	tCtx, span := h.obs.TracerService.StartTracer(ctx, functionName)
	defer h.obs.TracerService.StopSpan(span)

	h.obs.MetricsService.IncrementCounter(tCtx, functionName, 1, map[string]string{})
	h.obs.LoggerService.Info(tCtx, "ListTemplates request received")

	// Call service
	templates, err := h.service.ListTemplates(tCtx)
	if err != nil {
		h.obs.LoggerService.Error(tCtx, "Failed to retrieve templates from service", map[string]interface{}{
			"error": err.Error(),
		})
		h.obs.MetricsService.IncrementCounter(tCtx, functionFailed, 1, map[string]string{
			"error": errors.GetAppErrorMessage(errors.TmpErrTemplateListEmpty),
		})
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Success:      false,
			ErrorMessage: errors.GetAppErrorMessage(errors.TmpErrTemplateListEmpty),
			ErrorCode:    errors.TmpErrTemplateListEmpty,
			Data:         nil,
		})
	}

	// Check if templates list is empty
	if len(templates) == 0 {
		h.obs.LoggerService.Warn(tCtx, "Template list is empty")
		h.obs.MetricsService.IncrementCounter(tCtx, functionFailed, 1, map[string]string{
			"error": "template_list_empty",
		})
		return c.Status(fiber.StatusOK).JSON(dto.ErrorResponse{
			Success:      true,
			ErrorMessage: errors.GetAppErrorMessage(errors.TmpErrTemplateListEmpty),
			ErrorCode:    errors.TmpErrTemplateListEmpty,
			Data:         nil,
		})
	}

	// Log the count of retrieved templates
	h.obs.LoggerService.Info(tCtx, "Templates retrieved from service", map[string]interface{}{
		"template_count": len(templates),
	})

	// Convert templates to response format
	templateResponses := make([]dto.TemplateResponse, len(templates))
	for i, template := range templates {
		templateResponses[i] = dto.TemplateResponse{
			ID:             template.ID.String(),
			Name:           template.Name,
			Channel:        template.Channel,
			Language:       template.Language,
			Version:        int(template.Version),
			IsActive:       template.IsActive,
			Content:        template.Content,
			RequiredFields: template.RequiredFields, // Add RequiredFields to the response
			CreatedAt:      template.CreatedAt,
			UpdatedAt:      template.UpdatedAt,
		}
	}

	h.obs.LoggerService.Info(tCtx, "Templates processed for response")
	return c.JSON(dto.ListTemplatesResponse{
		Success: true,
		Message: "Templates retrieved successfully",
		Data:    templateResponses,
	})
}
