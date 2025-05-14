package methods

import (
	"context"
	"thirdparty_service/internal/interface/grpc/mock"
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
		Type:    req.Type,
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

func (h *ThirdPartyServer) VerifyAadhaar(ctx context.Context, req *protos.AadharVerifyRequest) (*protos.AadharVerifyResponse, error) {

	// validate request
	payload := &dtos.VerifyAadharRequest{
		AadharNumber: req.Number,
	}
	validationErr := utils.Validate(payload)
	if validationErr != nil || len(validationErr) > 0 {
		err := validationErr[0]
		return &protos.AadharVerifyResponse{
			Code:     err.Code,
			Message:  codes.ErrorMessage(err.Code),
			Verified: false,
		}, nil
	}

	ok, result, err := mock.VerifyAadhar(payload.AadharNumber)

	if err != nil || !ok {
		return &protos.AadharVerifyResponse{
			Code:     codes.TS1010,
			Message:  codes.ErrorMessage(codes.TS1010),
			Verified: false,
		}, nil
	}

	return &protos.AadharVerifyResponse{
		Code:     codes.TS0001,
		Message:  codes.SuccessMessage(codes.TS0001),
		Verified: true,
		Data: &protos.AadharResult{
			Name: result["name"],
			Dob:  result["dob"],
		},
	}, nil
}

func (h *ThirdPartyServer) VerifyPAN(ctx context.Context, req *protos.PANVerifyRequest) (*protos.PANVerifyResponse, error) {

	// validate request
	payload := &dtos.VerifyPANRequest{
		PANNumber: req.Number,
	}
	validationErr := utils.Validate(payload)
	if validationErr != nil || len(validationErr) > 0 {
		err := validationErr[0]
		return &protos.PANVerifyResponse{
			Code:     err.Code,
			Message:  codes.ErrorMessage(err.Code),
			Verified: false,
		}, nil
	}

	ok, result, err := mock.VerifyAadhar(payload.PANNumber)

	if err != nil || !ok {
		return &protos.PANVerifyResponse{
			Code:     codes.TS1011,
			Message:  codes.ErrorMessage(codes.TS1011),
			Verified: false,
		}, nil
	}

	return &protos.PANVerifyResponse{
		Code:     codes.TS0001,
		Message:  codes.SuccessMessage(codes.TS0001),
		Verified: true,
		Data: &protos.PanResult{
			Name:    result["name"],
			Dob:     result["dob"],
			PanType: result["pan_type"],
		},
	}, nil
}
