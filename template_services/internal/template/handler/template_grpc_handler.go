package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"template-services/internal/models"
	"template-services/internal/template/middleware"
	"template-services/internal/template/service"
	cacheclient "template-services/pkg/client/cache_client"
	"template-services/pkg/errors"
	appErrors "template-services/pkg/errors"
	"template-services/pkg/observability"
	pb "template-services/proto"
	"time"
)

type TemplateGRPCHandler struct {
	service service.TemplateServiceInterface
	obs     *observability.ObservabilityStack
	cache   *cacheclient.RedisClientStruct
	pb.UnimplementedTemplateServiceServer
}

func NewTemplateGRPCHandler(service service.TemplateServiceInterface, obs *observability.ObservabilityStack, cache *cacheclient.RedisClientStruct) *TemplateGRPCHandler {
	return &TemplateGRPCHandler{
		service: service,
		obs:     obs,
		cache:   cache,
	}
}

func (h *TemplateGRPCHandler) GetTemplateV1(ctx context.Context, req *pb.GetTemplateRequest) (*pb.TemplateResponse, error) {
	functionName := "GetTemplateV1"
	functionFailed := "GetTemplateV1_Failed"

	if h == nil || h.obs.TracerService == nil {
		log.Println("[PANIC GUARD] TemplateGRPCHandler or its dependencies are nil")
		return buildErrorResponse(nil, "appErrors.TmpErrInternalServer"), nil
	}

	tCtx, span := h.obs.TracerService.StartTracer(ctx, functionName)
	defer func() {
		if r := recover(); r != nil {
			h.obs.LoggerService.Error(tCtx, fmt.Errorf("panic recovered in GetTemplateV1: %v", r))
		}
		h.obs.TracerService.StopSpan(span)
	}()

	h.obs.LoggerService.Info(tCtx, "Starting GetTemplateV1 request", map[string]interface{}{
		"name":     req.GetName(),
		"channel":  req.GetChannel(),
		"language": req.GetLanguage(),
	})

	if h.obs.MetricsService != nil {
		h.obs.MetricsService.IncrementCounter(tCtx, functionName, 1, map[string]string{})
	}

	// Input DTO validation
	dtoReq := struct {
		Name     string `validate:"required" error_code:"TMP009"`
		Channel  string `validate:"required,oneof=email sms push" error_code:"TMP0010"`
		Language string `validate:"required,len=2" error_code:"TMP0011"`
	}{
		Name:     req.GetName(),
		Channel:  req.GetChannel(),
		Language: req.GetLanguage(),
	}

	if errMap, isValid := middleware.ValidateStruct(dtoReq); !isValid {
		h.obs.LoggerService.Warn(tCtx, "Validation failed for GetTemplateV1", map[string]interface{}{
			"errors": errMap,
		})

		sanitized := make(map[string]string)
		for k, v := range errMap {
			sanitized[k] = sanitize(v)
		}
		return buildErrorResponse(sanitized, appErrors.TmpErrInvalidRequestBody), nil
	}

	cacheKey := fmt.Sprintf("template:%s:%s:%s", req.Name, req.Channel, req.Language)
	h.obs.LoggerService.Info(tCtx, "Attempting to fetch from cache", map[string]interface{}{
		"cacheKey": cacheKey,
	})

	var cachedValue string
	var found bool
	var err error

	if found {
		var cachedTemplate models.Template
		if err := json.Unmarshal([]byte(cachedValue), &cachedTemplate); err == nil {
			h.obs.LoggerService.Info(tCtx, "Template retrieved from cache", map[string]interface{}{
				"name":     cachedTemplate.Name,
				"channel":  cachedTemplate.Channel,
				"language": cachedTemplate.Language,
			})
			return buildSuccessResponse(&cachedTemplate), nil
		} else {
			h.obs.LoggerService.Warn(tCtx, "Failed to unmarshal cached template", map[string]interface{}{
				"cacheKey": cacheKey,
				"error":    err.Error(),
			})
		}
	}

	h.obs.LoggerService.Info(tCtx, "Cache miss. Fetching template from service", map[string]interface{}{
		"name":     req.Name,
		"channel":  req.Channel,
		"language": req.Language,
	})

	if h.service == nil {
		h.obs.LoggerService.Error(tCtx, fmt.Errorf("template service is nil"))
		return buildErrorResponse(nil, "appErrors.TmpErrInternalServer"), nil
	}

	resp, err := h.service.GetTemplate(ctx, "", req.Name, req.Channel, req.Language)
	if err != nil {
		h.obs.LoggerService.Error(tCtx, err)
		if h.obs.MetricsService != nil {
			h.obs.MetricsService.IncrementCounter(tCtx, functionFailed, 1, map[string]string{
				"error": errors.GetAppErrorMessage(errors.TmpErrTemplateNotFound),
			})
		}
		return buildErrorResponse(
			map[string]string{
				"template": appErrors.GetAppErrorMessage(appErrors.TmpErrTemplateNotFound),
			},
			appErrors.TmpErrTemplateNotFound,
		), nil
	}

	if resp != nil {
		if respBytes, err := json.Marshal(resp); err == nil {
			if h.cache != nil {
				if err := h.cache.SetCache(tCtx, "templates", cacheKey, string(respBytes), 240*time.Hour); err != nil {
					h.obs.LoggerService.Error(tCtx, fmt.Errorf("cache set error: %w", err))
				} else {
					h.obs.LoggerService.Info(tCtx, "Template cached successfully", map[string]interface{}{
						"cacheKey": cacheKey,
					})
				}
			}
		} else {
			h.obs.LoggerService.Warn(tCtx, "Failed to marshal template response for caching", map[string]interface{}{
				"error": err.Error(),
			})
		}
	}

	h.obs.LoggerService.Info(tCtx, "Template retrieved successfully", map[string]interface{}{
		"name":     resp.Name,
		"channel":  resp.Channel,
		"language": resp.Language,
	})

	return buildSuccessResponse(resp), nil
}

// Helper function to build success response
func buildSuccessResponse(resp *models.Template) *pb.TemplateResponse {
	requiredFieldsStr := ""
	if len(resp.RequiredFields) > 0 {
		requiredFieldsStr = strings.Join(resp.RequiredFields, ",")
	}
	return &pb.TemplateResponse{
		Success: true,
		Message: map[string]string{
			"message": "Template retrieved successfully",
		},
		Data: map[string]string{
			"id":              resp.ID.String(),
			"name":            resp.Name,
			"channel":         resp.Channel,
			"language":        resp.Language,
			"content":         resp.Content,
			"required_fields": requiredFieldsStr,
			"version":         fmt.Sprintf("%d", resp.Version),
			"is_active":       fmt.Sprintf("%v", resp.IsActive),
			"created_at":      resp.CreatedAt.Format(time.RFC3339),
			"updated_at":      resp.UpdatedAt.Format(time.RFC3339),
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
