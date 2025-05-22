package handlers

import (
	"context"
	"fmt"
	"log"

	pb "Document-Processing/proto"
)

type HealthHandler struct {
	pb.UnimplementedHealthServiceServer
}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

func (h *HealthHandler) Check(ctx context.Context, req *pb.HealthCheckRequest) (*pb.HealthCheckResponse, error) {
	// Log the health check request
	log.Printf("Health check requested for service: %s", req.Service)

	// Check if the requested service is valid
	if req.Service == "" {
		return &pb.HealthCheckResponse{
			Status: pb.HealthCheckResponse_UNKNOWN,
		}, fmt.Errorf("service name is required")
	}

	// For now, we only support the document processing service
	if req.Service != "documentprocessing.DocumentProcessingService" {
		return &pb.HealthCheckResponse{
			Status: pb.HealthCheckResponse_NOT_SERVING,
		}, fmt.Errorf("unsupported service: %s", req.Service)
	}

	// Return serving status
	return &pb.HealthCheckResponse{
		Status: pb.HealthCheckResponse_SERVING,
	}, nil
}
