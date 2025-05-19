package authRepository

import (
	commonDtos "protected_link/internal/common/api/dtos"
	"protected_link/internal/modules/authentication/models"
	apiDtos "protected_link/internal/modules/link_generation/apis/dtos"
)

type IOTPRepository interface {
	SendOtp(request apiDtos.GenerateUrlRequest, otp string, dbId string) (*commonDtos.ApiResponseDto, error)
	GetOTP(userID string) (string, error)
	VerifyOtp(request *models.VerifyOTPRequest) (*commonDtos.ApiResponseDto, error)
}
