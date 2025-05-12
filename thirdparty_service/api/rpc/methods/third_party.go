package methods

import (
	"context"
	"thirdparty_service/internal/services"
	protos "thirdparty_service/protos"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ThirdPartyServer struct {
	protos.UnimplementedThirdPartyServiceServer
	svc services.Service
}

type AadhaarVerificationResult struct {
	Verified bool
	Name     string
	Dob      string
}

func (h *ThirdPartyServer) VerifyAadhaar(ctx context.Context, req *protos.VerifyAadhaarRequest) (*protos.VerifyAadhaarResponse, error) {
	result, name, dob, err := h.svc.VerifyAadhaar(req.AadhaarNumber)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "VerifyAadhaar failed: %v", err)
	}
	return &protos.VerifyAadhaarResponse{
		Verified: result,
		Name:     name,
		Dob:      dob,
	}, nil
}

// Other gRPC methods follow the same structure...
