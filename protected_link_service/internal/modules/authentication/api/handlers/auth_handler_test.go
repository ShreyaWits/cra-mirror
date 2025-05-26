package handlers

import (
	"context"
	"errors"
	"testing"

	commonDtos "protected_link/internal/common/api/dtos"
	"protected_link/internal/modules/authentication/models"

	pb "protected_link/pkg/grpc/proto"
	"protected_link/pkg/validation"

	"github.com/stretchr/testify/assert"
)

// mockAuthService implements a mock of IAuthenticationService
type mockAuthService struct {
	VerifyOTPFunc    func(req *models.VerifyOTPRequest) (*commonDtos.ApiResponseDto, error)
	GetAuthTokenFunc func(userID string, token string) (*apiDtos.GenerateUrlRequest, error)
	SendOtpFunc      func(req *apiDtos.GenerateUrlRequest, phone string) (*commonDtos.ApiResponseDto, error)
}

func (m *mockAuthService) VerifyOTP(req *models.VerifyOTPRequest) (*commonDtos.ApiResponseDto, error) {
	if m.VerifyOTPFunc != nil {
		return m.VerifyOTPFunc(req)
	}
	return nil, nil
}

func (m *mockAuthService) GetAuthToken(userID string, token string) (*apiDtos.GenerateUrlRequest, error) {
	if m.GetAuthTokenFunc != nil {
		return m.GetAuthTokenFunc(userID, token)
	}
	return &apiDtos.GenerateUrlRequest{
		UserID:      userID,
		Name:        "Test User",
		RequestType: "access",
		ModelType:   "jwt",
		Email:       "test@example.com",
		ExpireIn:    "10m",
		OtpRequired: false,
		Phone:       "1234567890",
		ChannelType: "email",
		Data:        map[string]interface{}{"key": "value"},
	}, nil
}

func (m *mockAuthService) SendOtp(req *apiDtos.GenerateUrlRequest, phone string) (*commonDtos.ApiResponseDto, error) {
	if m.SendOtpFunc != nil {
		return m.SendOtpFunc(req, phone)
	}
	return &commonDtos.ApiResponseDto{
		Success: true,
		Message: "OTP sent successfully",
	}, nil
}

// ------------------------
// TEST CASES START HERE
// ------------------------

func TestVerifyOTPV1_NilRequest(t *testing.T) {
	handler := NewAuthHandler(&mockAuthService{})

	resp, err := handler.VerifyOTPV1(context.Background(), nil)

	assert.NoError(t, err)
	assert.False(t, resp.Success)
	assert.Equal(t, "Validation failed", resp.Message)
	assert.Contains(t, resp.Error["error"], "request cannot be nil")
}

func TestVerifyOTPV1_InvalidInternalRequest(t *testing.T) {
	validation.InitValidator() // ✅ Initialize the validator here
	handler := NewAuthHandler(&mockAuthService{})

	req := &pb.VerifyOTPRequestV1{
		UserId:         "",
		Otp:            "",
		VerificationId: "",
	}

	resp, err := handler.VerifyOTPV1(context.Background(), req)

	assert.NoError(t, err)
	assert.False(t, resp.Success)
	assert.Equal(t, "Validation failed", resp.Message)
	assert.Contains(t, resp.Error["error"], "validation failed")
}

func TestVerifyOTPV1_ServiceReturnsError(t *testing.T) {
	validation.InitValidator() // ✅ Initialize the validator here
	mockService := &mockAuthService{
		VerifyOTPFunc: func(req *models.VerifyOTPRequest) (*commonDtos.ApiResponseDto, error) {
			return nil, errors.New("OTP verification failed")
		},
	}

	handler := NewAuthHandler(mockService)

	req := &pb.VerifyOTPRequestV1{
		UserId:         "123",
		Otp:            "456789",
		VerificationId: "abc-123",
	}

	resp, err := handler.VerifyOTPV1(context.Background(), req)

	assert.NoError(t, err)
	assert.False(t, resp.Success)
	assert.Equal(t, "Validation failed", resp.Message)
	assert.Contains(t, resp.Error["error"], "OTP verification failed")
}

func TestVerifyOTPV1_Success(t *testing.T) {
	validation.InitValidator() // ✅ Initialize the validator here
	mockService := &mockAuthService{
		VerifyOTPFunc: func(req *models.VerifyOTPRequest) (*commonDtos.ApiResponseDto, error) {
			return &commonDtos.ApiResponseDto{
				Success: true,
				Message: "OTP verified successfully",
				Data:    map[string]string{"status": "verified"},
			}, nil
		},
	}

	handler := NewAuthHandler(mockService)

	req := &pb.VerifyOTPRequestV1{
		UserId:         "123",
		Otp:            "456789",
		VerificationId: "abc-123",
	}

	resp, err := handler.VerifyOTPV1(context.Background(), req)

	assert.NoError(t, err)
	assert.True(t, resp.Success)
	assert.Equal(t, "OTP verified successfully", resp.Message)
	assert.Equal(t, "verified", resp.Data["status"])
}

func TestMapToInternalRequest(t *testing.T) {
	handler := NewAuthHandler(&mockAuthService{})

	req := &pb.VerifyOTPRequestV1{
		UserId:         "user123",
		Otp:            "otp456",
		VerificationId: "verif789",
	}

	internalReq := handler.mapToInternalRequest(req)

	assert.Equal(t, "user123", internalReq.UserID)
	assert.Equal(t, "otp456", internalReq.OTP)
	assert.Equal(t, "verif789", internalReq.VerificationID)
}

func TestConvertToMapString_WithMapString(t *testing.T) {
	handler := NewAuthHandler(&mockAuthService{})

	input := map[string]string{"foo": "bar", "baz": "qux"}
	output := handler.convertToMapString(input)

	assert.Equal(t, input, output)
}

func TestConvertToMapString_WithMapInterface(t *testing.T) {
	handler := NewAuthHandler(&mockAuthService{})

	input := map[string]interface{}{"foo": "bar", "number": 123}
	expected := map[string]string{"foo": "bar", "number": "123"}

	output := handler.convertToMapString(input)
	assert.Equal(t, expected, output)
}

type customStruct struct {
	Name  string `json:"name"`
	Value int    `json:"value"`
}

func TestConvertToMapString_WithStruct(t *testing.T) {
	handler := NewAuthHandler(&mockAuthService{})

	data := customStruct{Name: "Alice", Value: 42}
	expected := map[string]string{"name": "Alice", "value": "42"}

	output := handler.convertToMapString(data)
	assert.Equal(t, expected, output)
}

func TestConvertToMapString_WithUnsupportedType(t *testing.T) {
	handler := NewAuthHandler(&mockAuthService{})

	input := "just a string"
	output := handler.convertToMapString(input)

	assert.Empty(t, output)
}

func TestConvertToMapString_WithNil(t *testing.T) {
	handler := NewAuthHandler(&mockAuthService{})
	output := handler.convertToMapString(nil)
	assert.Empty(t, output)
}

type complexStruct struct {
	Name    string                 `json:"name"`
	Age     int                    `json:"age"`
	Nested  *nestedStruct          `json:"nested"`
	MapData map[string]interface{} `json:"map_data"`
}

type nestedStruct struct {
	Value string `json:"value"`
}

func TestConvertToMapString_WithComplexStruct(t *testing.T) {
	handler := NewAuthHandler(&mockAuthService{})

	data := complexStruct{
		Name: "John",
		Age:  30,
		Nested: &nestedStruct{
			Value: "nested value",
		},
		MapData: map[string]interface{}{
			"key1": "value1",
			"key2": 123,
		},
	}

	expected := map[string]string{
		"name":     "John",
		"age":      "30",
		"nested":   "&{nested value}",
		"map_data": "map[key1:value1 key2:123]",
	}

	output := handler.convertToMapString(data)
	assert.Equal(t, expected, output)
}

type structWithUnexported struct {
	Name    string `json:"name"`
	age     int    `json:"age"` // unexported field
	private string // no json tag
}

func TestStructToMapString_WithUnexportedFields(t *testing.T) {
	handler := NewAuthHandler(&mockAuthService{})

	data := structWithUnexported{
		Name:    "John",
		age:     30,
		private: "private",
	}

	expected := map[string]string{
		"name": "John",
	}

	output := handler.structToMapString(data)
	assert.Equal(t, expected, output)
}

type customError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *customError) Error() string {
	return e.Message
}

func TestConvertErrorToMap_WithCustomError(t *testing.T) {
	handler := NewAuthHandler(&mockAuthService{})

	err := &customError{
		Code:    "E001",
		Message: "Custom error message",
	}

	expected := map[string]string{
		"code":    "E001",
		"message": "Custom error message",
	}

	output := handler.convertErrorToMap(err)
	assert.Equal(t, expected, output)
}

func TestConvertErrorToMap_WithMapString(t *testing.T) {
	handler := NewAuthHandler(&mockAuthService{})

	errInput := map[string]string{"code": "400", "message": "Bad request"}
	output := handler.convertErrorToMap(errInput)

	assert.Equal(t, errInput, output)
}

func TestConvertErrorToMap_WithNil(t *testing.T) {
	handler := NewAuthHandler(&mockAuthService{})

	output := handler.convertErrorToMap(nil)
	assert.Empty(t, output)
}
