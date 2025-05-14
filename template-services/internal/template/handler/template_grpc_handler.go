package handler

import (
	"context"
	"strings"

	appErrors "template-services/internal/pkg/errors"
	"template-services/internal/template/dto"
	"template-services/internal/template/middleware"
	"template-services/internal/template/service"
	pb "template-services/proto"
)

type TemplateGRPCHandler struct {
	service service.TemplateServiceInterface
	pb.UnimplementedTemplateServiceServer
}

func NewTemplateGRPCHandler(service service.TemplateServiceInterface) *TemplateGRPCHandler {
	return &TemplateGRPCHandler{
		service: service,
	}
}


func (h *TemplateGRPCHandler) GetTemplateV1(ctx context.Context, req *pb.GetTemplateRequest) (*pb.TemplateResponse, error) {
	// Map gRPC request to DTO
	dtoReq := dto.GetTemplateRequestV1{
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

	// Call the service
	resp, err := h.service.GetTemplate(ctx, "", req.Name, req.Channel, req.Language)
	if err != nil {
		return buildErrorResponse(
			map[string]string{
				"template": appErrors.GetAppErrorMessage(appErrors.TmpErrTemplateNotFound),
			},
			appErrors.TmpErrTemplateNotFound,
		), nil
	}

	// Return success
	return &pb.TemplateResponse{
		Success: true,
		Message: map[string]string{
			"info": "Template retrieved successfully",
		},
		Error: nil,
		Data: map[string]string{
			"id":       resp.ID.String(),
			"name":     resp.Name,
			"channel":  resp.Channel,
			"language": resp.Language,
		},
	}, nil
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
