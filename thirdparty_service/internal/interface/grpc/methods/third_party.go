package methods

import (
	"context"
	"thirdparty_service/internal/modules/execute/services"
	protos "thirdparty_service/proto"

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

func InitMethods(svc services.Service) *ThirdPartyServer {
	return &ThirdPartyServer{
		svc: svc,
	}
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

// Invoke Twilio SMS
func (h *ThirdPartyServer) InvokeTwilioSms(ctx context.Context, req *protos.SendSMSRequest) (*protos.SendSMSResponse, error) {

	err := h.svc.SendTwilioSms(req.Phone, req.Message)
	if err != nil {
		return &protos.SendSMSResponse{
			Status: "FAILED",
		}, nil
	}

	return &protos.SendSMSResponse{
		Status: "SUCCESS",
	}, nil
}
