package services

import (
	"fmt"
	"log"

	commonDtos "protected_link/internal/common/api/dtos"

	"protected_link/internal/modules/authentication/api/utils"
	"protected_link/internal/modules/authentication/models"
	authRepository "protected_link/internal/modules/authentication/repositories"
	repository "protected_link/internal/modules/cassandra/repository"
	apiDtos "protected_link/internal/modules/link_generation/apis/dtos"
	"protected_link/pkg/jwt"
)

// AuthenticationService handles authentication-related operations
type AuthenticationService struct {
	otpRepo    *authRepository.OTPRepository
	casendra   repository.ICassandraRepository
	jwtService *jwt.JwtCreation
}

// NewAuthenticationService creates a new instance of AuthenticationService
func NewAuthenticationService(otpRepo *authRepository.OTPRepository, casendra repository.ICassandraRepository) *AuthenticationService {
	jwtService, err := jwt.NewJwtCreation()
	if err != nil {
		log.Printf("Failed to initialize JWT service: %v", err)
		return nil
	}

	return &AuthenticationService{
		otpRepo:    otpRepo,
		casendra:   casendra,
		jwtService: jwtService,
	}
}

// GetAuthToken retrieves an authentication token for a user
func (s *AuthenticationService) GetAuthToken(userID, tokenID string) (*apiDtos.GenerateUrlRequest, error) {
	if s.casendra == nil {
		return nil, fmt.Errorf("cassandra repository not initialized")
	}

	// For hybrid model, get from Cassandra
	data, err := s.casendra.GetDataByID(tokenID)
	if err != nil {
		return nil, fmt.Errorf("failed to get auth token: %w", err)
	}

	return data, nil
}

// SendOtp sends an OTP to the user
func (s *AuthenticationService) SendOtp(request *apiDtos.GenerateUrlRequest, dbId string) (*commonDtos.ApiResponseDto, error) {
	if s.otpRepo == nil {
		return nil, fmt.Errorf("OTP repository not initialized")
	}

	if request == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	// Generate OTP
	otp := utils.GenerateOTP()
	log.Printf("Generated OTP :[%s]", otp)

	// Send OTP via repository
	res, err := s.otpRepo.SendOtp(*request, otp, dbId)
	if err != nil {
		return nil, fmt.Errorf("failed to send OTP: %w", err)
	}

	return res, nil
}

// VerifyOTP verifies an OTP for a user
func (s *AuthenticationService) VerifyOTP(request *models.VerifyOTPRequest) (*commonDtos.ApiResponseDto, error) {
	if s.otpRepo == nil {
		return nil, fmt.Errorf("OTP repository not initialized")
	}

	if request == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	result, err := s.otpRepo.VerifyOtp(request)
	if err != nil {
		return nil, fmt.Errorf("failed to verify OTP: %w", err)
	}

	return result, nil
}
