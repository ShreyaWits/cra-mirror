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
func (h *ThirdPartyServer) InvokeSendgridEmail(ctx context.Context, req *protos.InvokeSendGridRequest) (*protos.InvokeSendGridResponse, error) {

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

func (h *ThirdPartyServer) InvokeWhatsAppMessage(ctx context.Context, req *protos.InvokeWhatsAppRequest) (*protos.InvokeWhatsAppResponse, error) {
	// validate request
	payload := &dtos.SendWhatsAppMessageRequest{
		Phone:       req.Phone,
		Message:     req.Message,
		CountryCode: req.CountryCode,
	}
	validationErr := utils.Validate(payload)
	if len(validationErr) > 0 {
		err := validationErr[0]
		return &protos.InvokeWhatsAppResponse{
			Code:    err.Code,
			Message: codes.ErrorMessage(err.Code),
		}, nil
	}

	// send WhatsApp message
	err := h.svc.SendWhatsAppMessage(payload)
	if err != nil {
		return &protos.InvokeWhatsAppResponse{
			Code:    codes.TS1012, // define this code in `pkg/codes`
			Message: codes.ErrorMessage(codes.TS1012),
		}, nil
	}

	return &protos.InvokeWhatsAppResponse{
		Code:    codes.TS0001,
		Message: codes.SuccessMessage(codes.TS0001),
	}, nil
}

func (h *ThirdPartyServer) InvokePushNotification(ctx context.Context, req *protos.PushNotificationRequest) (*protos.PushNotificationResponse, error) {
	payload := &dtos.PushNotificationRequest{
		ToToken: req.ToToken,
		Title:   req.Title,
		Body:    req.Body,
	}

	validationErr := utils.Validate(payload)
	if validationErr != nil || len(validationErr) > 0 {
		err := validationErr[0]
		return &protos.PushNotificationResponse{
			Status:  err.Code,
			Message: codes.ErrorMessage(err.Code),
		}, nil
	}

	err := h.svc.SendPushNotification(payload)
	if err != nil {
		return &protos.PushNotificationResponse{
			Status:  codes.TS1013,
			Message: codes.ErrorMessage(codes.TS1013),
		}, nil
	}

	return &protos.PushNotificationResponse{
		Status:  codes.TS0001,
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

	ok, name, dob, err := h.svc.VerifyAadhaar(payload.AadharNumber)

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
			Name: name,
			Dob:  dob,
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

	ok, name, panType, err := h.svc.VerifyPAN(payload.PANNumber)

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
			Name:    name,
			Dob:     "", // The service layer doesn't return Dob for PAN, only Name and PanType
			PanType: panType,
		},
	}, nil
}
