package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"template-services/internal/models"
	"template-services/internal/template/middleware"
	"template-services/internal/template/service"
	"template-services/pkg/errors"
	appErrors "template-services/pkg/errors"
	"template-services/pkg/observability"
	pb "template-services/proto"
	"time"
)

type TemplateGRPCHandler struct {
	service service.TemplateServiceInterface
	obs     *observability.ObservabilityStack
	pb.UnimplementedTemplateServiceServer
}

func NewTemplateGRPCHandler(service service.TemplateServiceInterface, obs *observability.ObservabilityStack) *TemplateGRPCHandler {
	return &TemplateGRPCHandler{
		service: service,
		obs:     obs,
	}
}

func (h *TemplateGRPCHandler) GetTemplateV1(ctx context.Context, req *pb.GetTemplateRequest) (*pb.TemplateResponse, error) {
	functionName := "GetTemplateV1"
	functionFailed := "GetTemplateV1_Failed"

	tCtx, span := h.obs.TracerService.StartTracer(ctx, functionName)
	defer h.obs.TracerService.StopSpan(span)
	h.obs.MetricsService.IncrementCounter(tCtx, functionName, 1, map[string]string{})

	// Map gRPC request to DTO
	dtoReq := struct {
		Name     string `validate:"required" error_code:"TMP009"`
		Channel  string `validate:"required,oneof=email sms push" error_code:"TMP0010"`
		Language string `validate:"required,len=2" error_code:"TMP0011"`
	}{
		Name:     req.Name,
		Channel:  req.Channel,
		Language: req.Language,
	}

	// Validate input using middleware
	if errMap, isValid := middleware.ValidateStruct(dtoReq); !isValid {
		sanitized := make(map[string]string)
		for k, v := range errMap {
			sanitized[k] = sanitize(v)
		}
		return buildErrorResponse(sanitized, appErrors.TmpErrInvalidRequestBody), nil
	}

	// Try to get from cache first
	cacheKey := fmt.Sprintf("template:%s:%s:%s", req.Name, req.Channel, req.Language)
	cachedValue, found, err := h.obs.CacheClient.GetCache(tCtx, "templates", cacheKey, "")
	if err != nil {
		h.obs.LoggerService.Error(tCtx, fmt.Errorf("cache get error: %w", err))
	}

	if found {
		var cachedTemplate models.Template
		if err := json.Unmarshal([]byte(cachedValue), &cachedTemplate); err == nil {
			h.obs.LoggerService.Info(tCtx, "Template retrieved from cache")
			return buildSuccessResponse(&cachedTemplate), nil
		}
	}

	// Call the service if not in cache
	resp, err := h.service.GetTemplate(ctx, "", req.Name, req.Channel, req.Language)
	if err != nil {
		h.obs.LoggerService.Error(tCtx, err)
		h.obs.MetricsService.IncrementCounter(tCtx, functionFailed, 1, map[string]string{"error": errors.GetAppErrorMessage(errors.TmpErrTemplateNotFound)})
		return buildErrorResponse(
			map[string]string{
				"template": appErrors.GetAppErrorMessage(appErrors.TmpErrTemplateNotFound),
			},
			appErrors.TmpErrTemplateNotFound,
		), nil
	}

	// Cache the result
	if resp != nil {
		if respBytes, err := json.Marshal(resp); err == nil {
			if err := h.obs.CacheClient.SetCache(tCtx, "templates", cacheKey, string(respBytes), 240*time.Hour, ""); err != nil {
				h.obs.LoggerService.Error(tCtx, fmt.Errorf("cache set error: %w", err))
			}
		}
	}

	h.obs.LoggerService.Info(tCtx, "Template retrieved successfully")
	return buildSuccessResponse(resp), nil
}

// Helper function to build success response
func buildSuccessResponse(resp *models.Template) *pb.TemplateResponse {
	return &pb.TemplateResponse{
		Success: true,
		Message: map[string]string{
			"info": "Template retrieved successfully",
		},
		Error: nil,
		Data: map[string]string{
			"id":              resp.ID.String(),
			"name":            resp.Name,
			"channel":         resp.Channel,
			"language":        resp.Language,
			"content":         resp.Content,
			"required_fields": strings.Join(resp.RequiredFields, ","),
			"version":         fmt.Sprintf("%d", resp.Version),
			"is_active":       fmt.Sprintf("%t", resp.IsActive),
			"created_at":      resp.CreatedAt.String(),
			"updated_at":      resp.UpdatedAt.String(),
		},
	}
}

// Helper function to build error response
func buildErrorResponse(message map[string]string, errorCode string) *pb.TemplateResponse {
	return &pb.TemplateResponse{
		Success: false,
		Message: message,
		Error: map[string]string{
			"code": errorCode,
		},
		Data: map[string]string{},
	}
}

// Sanitize message to remove unwanted characters
func sanitize(input string) string {
	return strings.ReplaceAll(strings.TrimSpace(input), "\n", "")
}
