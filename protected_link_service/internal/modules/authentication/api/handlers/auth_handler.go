package handlers

import (
	"context"
	"fmt"
	commonDtos "protected_link/internal/common/api/dtos"
	"protected_link/internal/modules/authentication/models"
	authService "protected_link/internal/modules/authentication/services"
	pb "protected_link/pkg/grpc/proto"
	"protected_link/pkg/validation"
	"reflect"
)

// AuthHandler handles authentication-related gRPC requests
type AuthHandler struct {
	authService *authService.AuthenticationService
	pb.UnimplementedAuthServiceServer
}

// NewAuthHandler creates a new instance of AuthHandler
func NewAuthHandler(authService *authService.AuthenticationService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// VerifyOTPV1 handles OTP verification requests
func (h *AuthHandler) VerifyOTPV1(ctx context.Context, req *pb.VerifyOTPRequestV1) (*pb.VerifyOTPResponseV1, error) {
	if err := h.validateRequest(req); err != nil {
		return h.createErrorResponse(err), nil
	}

	internalReq := h.mapToInternalRequest(req)
	if err := h.validateInternalRequest(internalReq); err != nil {
		return h.createErrorResponse(err), nil
	}

	result, err := h.authService.VerifyOTP(internalReq)
	if err != nil {
		return h.createErrorResponse(err), nil
	}

	return h.createSuccessResponse(result), nil
}

// validateRequest checks if the request is valid
func (h *AuthHandler) validateRequest(req *pb.VerifyOTPRequestV1) error {
	if req == nil {
		return fmt.Errorf("request cannot be nil")
	}
	return nil
}

// mapToInternalRequest converts gRPC request to internal model
func (h *AuthHandler) mapToInternalRequest(req *pb.VerifyOTPRequestV1) *models.VerifyOTPRequest {
	return &models.VerifyOTPRequest{
		UserID:         req.UserId,
		OTP:            req.Otp,
		VerificationID: req.VerificationId,
	}
}

// validateInternalRequest validates the internal request model
func (h *AuthHandler) validateInternalRequest(req *models.VerifyOTPRequest) error {
	fieldErrors, err := validation.ValidateVerifyOtpRequest(*req)
	if err != nil || len(fieldErrors) > 0 {
		return fmt.Errorf("validation failed: %v", fieldErrors)
	}
	return nil
}

// createErrorResponse creates an error response
func (h *AuthHandler) createErrorResponse(err error) *pb.VerifyOTPResponseV1 {
	return &pb.VerifyOTPResponseV1{
		Success: false,
		Message: "Validation failed",
		Error:   map[string]string{"error": err.Error()},
	}
}

// createSuccessResponse creates a success response
func (h *AuthHandler) createSuccessResponse(result *commonDtos.ApiResponseDto) *pb.VerifyOTPResponseV1 {
	response := &models.VerifyOTPResponse{
		Success: result.Success,
		Message: result.Message,
		Data:    result.Data,
		Error:   result.Error,
	}

	return &pb.VerifyOTPResponseV1{
		Success: response.Success,
		Message: response.Message,
		Data:    h.convertToMapString(response.Data),
		Error:   h.convertErrorToMap(response.Error),
	}
}

// convertToMapString converts interface{} to map[string]string
func (h *AuthHandler) convertToMapString(data interface{}) map[string]string {
	if data == nil {
		return map[string]string{}
	}

	switch v := data.(type) {
	case map[string]string:
		return v
	case map[string]interface{}:
		return h.convertMapInterfaceToString(v)
	default:
		return h.structToMapString(data)
	}
}

// convertMapInterfaceToString converts map[string]interface{} to map[string]string
func (h *AuthHandler) convertMapInterfaceToString(input map[string]interface{}) map[string]string {
	output := make(map[string]string, len(input))
	for key, value := range input {
		output[key] = fmt.Sprintf("%v", value)
	}
	return output
}

// structToMapString converts a struct to map[string]string
func (h *AuthHandler) structToMapString(data interface{}) map[string]string {
	result := make(map[string]string)
	val := reflect.ValueOf(data)

	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return result
	}

	typ := val.Type()
	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		if !val.Field(i).CanInterface() {
			continue
		}

		key := field.Tag.Get("json")
		if key == "" || key == "-" {
			key = field.Name
		}

		value := val.Field(i).Interface()
		result[key] = fmt.Sprintf("%v", value)
	}
	return result
}

// convertErrorToMap converts error interface to map[string]string
func (h *AuthHandler) convertErrorToMap(err interface{}) map[string]string {
	if err == nil {
		return map[string]string{}
	}

	if errMap, ok := err.(map[string]string); ok {
		return errMap
	}
	return map[string]string{}
}
