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

	if s.otpRepo == nil {
		return nil, fmt.Errorf("OTP repository is not initialized")
	}
	res, err := s.otpRepo.SendOtp(request, otp)
	if err != nil {
		return nil, fmt.Errorf("failed to save OTP: %w", err)
	}
	return res, nil
}

func (s *AuthenticationService) VerifyOTP(request *models.VerifyOTPRequest) (*commonDtos.ApiResponseDto, error) {
	result, err := s.otpRepo.VerifyOtp(request.UserID, request.OTP)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve OTP: %w", err)
	}

	if result.Message == "" {
		return nil, fmt.Errorf("OTP expired or not found")
	}

	return result, nil
}
