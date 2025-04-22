package services

import (
	"fmt"
	commonDtos "protected_link/internal/common/api/dtos"
	srv "protected_link/internal/modules/authentication/services"
	apiDtos "protected_link/internal/modules/link_generation/apis/dtos"
	"protected_link/internal/modules/link_generation/repositories"
	"protected_link/internal/modules/link_generation/utils"
)

type GenerateLinkService struct {
	repo *repositories.GeneratedRepository
}

// NewGenerateLinkService initializes a new GenerateLinkService
func NewGenerateLinkService(repo *repositories.GeneratedRepository) *GenerateLinkService {
	return &GenerateLinkService{
		repo: repo,
	}
}

func (s *GenerateLinkService) SaveGeneratedLink(dto *apiDtos.GenerateUrlRequest) (*commonDtos.ApiResponseDto, error) {
	return s.repo.SaveGeneratedLink(dto)
}

func (s *GenerateLinkService) GetExtractData(link *string) (*commonDtos.ApiResponseDto, error) {

	result, _ := s.repo.GetTokenData(link)

	print("dadaad", result.Data)

	dto, err := utils.ConvertToGenerateUrlRequest(result.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to convert data: %w", err)
	}

	// Check if OTP is required
	if dto.OtpRequired {

		initNotification := srv.NewAuthenticationService(nil)

		req, err := initNotification.SendOtp(*dto)

		println("Generated OTP Check Request	:", req)
		if err != nil {
			return nil, fmt.Errorf("failed to generate OTP: %w", err)
		}

		return req, nil
	}

	return s.repo.GetTokenData(link)
}
