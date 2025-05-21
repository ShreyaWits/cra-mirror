package authRepository_test

import (
	"errors"
	"testing"

	commonDtos "protected_link/internal/common/api/dtos"
	"protected_link/internal/modules/authentication/models"
	apiDtos "protected_link/internal/modules/link_generation/apis/dtos"

	"github.com/stretchr/testify/assert"
)

// Mock implementation of IOTPRepository
type MockOTPRepository struct{}

func (m *MockOTPRepository) SendOtp(request apiDtos.GenerateUrlRequest, otp string, dbId string) (*commonDtos.ApiResponseDto, error) {
	if otp == "123456" {
		return &commonDtos.ApiResponseDto{Message: "OTP Sent"}, nil
	}
	return nil, errors.New("Failed to send OTP")
}

func (m *MockOTPRepository) GetOTP(userID string) (string, error) {
	if userID == "valid-user" {
		return "123456", nil
	}
	return "", errors.New("User not found")
}

func (m *MockOTPRepository) VerifyOtp(request *models.VerifyOTPRequest) (*commonDtos.ApiResponseDto, error) {
	if request.OTP == "123456" {
		return &commonDtos.ApiResponseDto{Message: "OTP Verified"}, nil
	}
	return nil, errors.New("Invalid OTP")
}

// Success test cases
func TestSendOtp_Success(t *testing.T) {
	repo := &MockOTPRepository{}
	resp, err := repo.SendOtp(apiDtos.GenerateUrlRequest{}, "123456", "db1")

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "OTP Sent", resp.Message)
}

func TestGetOTP_Success(t *testing.T) {
	repo := &MockOTPRepository{}
	otp, err := repo.GetOTP("valid-user")

	assert.NoError(t, err)
	assert.Equal(t, "123456", otp)
}

func TestVerifyOtp_Success(t *testing.T) {
	repo := &MockOTPRepository{}
	req := &models.VerifyOTPRequest{
		OTP: "123456",
	}
	resp, err := repo.VerifyOtp(req)

	assert.NoError(t, err)
	assert.Equal(t, "OTP Verified", resp.Message)
}

// Failure test cases
func TestSendOtp_Failure(t *testing.T) {
	repo := &MockOTPRepository{}
	resp, err := repo.SendOtp(apiDtos.GenerateUrlRequest{}, "invalid", "db1")

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, "Failed to send OTP", err.Error())
}

func TestGetOTP_Failure(t *testing.T) {
	repo := &MockOTPRepository{}
	otp, err := repo.GetOTP("invalid-user")

	assert.Error(t, err)
	assert.Empty(t, otp)
	assert.Equal(t, "User not found", err.Error())
}

func TestVerifyOtp_Failure(t *testing.T) {
	repo := &MockOTPRepository{}
	req := &models.VerifyOTPRequest{
		OTP: "invalid",
	}
	resp, err := repo.VerifyOtp(req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, "Invalid OTP", err.Error())
}
