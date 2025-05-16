package handlers

import (
	"context"
	pb "encryption_microservice/internal/common/proto_gen"
)

// HandleHealth handles the health check request
func (h *EncryptionHandlerImpl) HealthCheck(c context.Context, req *pb.HealthCheckRequest) (*pb.HealthCheckResponse, error) {
	return &pb.HealthCheckResponse{
		HealthStatus: "OK",
	}, nil
}
