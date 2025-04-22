package services

import (
	"fmt"

	commonDtos "protected_link/internal/common/api/dtos"
	"protected_link/internal/modules/authentication/api/utils"
	"protected_link/internal/modules/authentication/models"
	authRepository "protected_link/internal/modules/authentication/repositories"
	apiDtos "protected_link/internal/modules/link_generation/apis/dtos"
)

type AuthenticationService struct {
	otpRepo *authRepository.OTPRepository
}

func NewAuthenticationService(otpRepo *authRepository.OTPRepository) *AuthenticationService {
	return &AuthenticationService{otpRepo: otpRepo}
}

func (s *AuthenticationService) SendOtp(request apiDtos.GenerateUrlRequest) (*commonDtos.ApiResponseDto, error) {

	otpService := utils.NewOTPService()

	// Generate an OTP
	otp := otpService.GenerateOTP() // Generate a 6-digit OTP
	// Set the expiration time for the OTP

	println("Generated OTP Check ait:", otp)
	res, err := s.otpRepo.SendOtp(request, otp)
	if err != nil {
		return nil, fmt.Errorf("failed to save OTP: %w", err)
	}
	return res, nil
}

func (s *AuthenticationService) VerifyOTP(request *models.VerifyOTPRequest) (bool, error) {
	storedOTP, err := s.otpRepo.GetOTP(request.UserID)
	if err != nil {
		return false, fmt.Errorf("failed to retrieve OTP: %w", err)
	}

	if storedOTP == "" {
		return false, fmt.Errorf("OTP expired or not found")
	}

	return storedOTP == "inputOTP", nil
}
