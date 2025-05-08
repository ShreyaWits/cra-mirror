package handler

import (
	"context"
	"log"
	"template-services/internal/template/service"
	pb "template-services/proto"
)

type TemplateGRPCHandler struct {
	service *service.TemplateService
	pb.UnimplementedTemplateServiceServer
}

func NewTemplateGRPCHandler(service *service.TemplateService) *TemplateGRPCHandler {
	return &TemplateGRPCHandler{
		service: service,
	}
}

func (h *TemplateGRPCHandler) GetTemplateV1(ctx context.Context, req *pb.GetTemplateRequest) (*pb.TemplateResponse, error) {
	// {
	// 	"name": "welcome_sms",
	// 	"channel": "sms",
	// 	"language": "en"
	// }

	resp, err := h.service.GetTemplate(ctx, req.Id, req.Name, req.Channel, req.Language)
	log.Println("🚀 ~ func ~ resp:", resp)
	if err != nil {
		return &pb.TemplateResponse{
			Success: false,
			Message: "Validation failed",
			Data:    nil,
			Error:   err.Error(),
		}, nil
	}

	return &pb.TemplateResponse{
		Success: true,
		Message: "Template retrieved successfully",
		Data: map[string]string{
			"ID":       resp.ID.String(),
			"Name":     resp.Name,
			"Channel":  resp.Channel,
			"Language": resp.Language,
		},
	}, nil
}
