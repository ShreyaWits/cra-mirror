package methods

import (
	"context"
	"thirdparty_service/internal/modules/execute/dtos"
	"thirdparty_service/internal/modules/execute/services"
	"thirdparty_service/internal/utils"
	"thirdparty_service/pkg/codes"
	protos "thirdparty_service/proto"
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

// Invoke Twilio SMS
func (h *ThirdPartyServer) InvokeTwilioSms(ctx context.Context, req *protos.InvokeTwilioRequest) (*protos.InvokeTwilioResponse, error) {

	// validate request
	payload := &dtos.TwilioSmsRequest{
		Phone:       req.Phone,
		Message:     req.Message,
		CountryCode: req.CountryCode,
	}

	validationErr := utils.Validate(payload)
	if validationErr != nil || len(validationErr) > 0 {
		err := validationErr[0]
		return &protos.InvokeTwilioResponse{
			Code:    err.Code,
			Message: codes.ErrorMessage(err.Code),
		}, nil
	}

	err := h.svc.SendTwilioSms(payload)
	if err != nil {
		return &protos.InvokeTwilioResponse{
			Code:    codes.TS1008,
			Message: codes.ErrorMessage(codes.TS1008),
		}, nil
	}

	return &protos.InvokeTwilioResponse{
		Code:    codes.TS0001,
		Message: codes.SuccessMessage(codes.TS0001),
	}, nil
}

// Invoke SendGrid Email
func (h *ThirdPartyServer) InvokeSendGridEmail(ctx context.Context, req *protos.InvokeSendGridRequest) (*protos.InvokeSendGridResponse, error) {

	// validate request
	payload := &dtos.SendGridEmailRequest{
		Subject: req.Subject,
		To:      req.To,
		Body:    req.Body,
	}
	validationErr := utils.Validate(payload)
	if validationErr != nil || len(validationErr) > 0 {
		err := validationErr[0]
		return &protos.InvokeSendGridResponse{
			Code:    err.Code,
			Message: codes.ErrorMessage(err.Code),
		}, nil
	}

	err := h.svc.SendEmailBySendGrid(payload)
	if err != nil {
		return &protos.InvokeSendGridResponse{
			Code:    codes.TS1007,
			Message: codes.ErrorMessage(codes.TS1007),
		}, nil
	}

	return &protos.InvokeSendGridResponse{
		Code:    codes.TS0001,
		Message: codes.SuccessMessage(codes.TS0001),
	}, nil
}
