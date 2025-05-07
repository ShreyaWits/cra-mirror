package handlers

import (
	"context"

	pb "Document-Processing/proto"
)

type HealthHandler struct {
	pb.UnimplementedHealthServiceServer
}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

func (h *HealthHandler) Check(ctx context.Context, req *pb.HealthCheckRequest) (*pb.HealthCheckResponse, error) {
	return &pb.HealthCheckResponse{
		Status: pb.HealthCheckResponse_SERVING,
	}, nil
}
