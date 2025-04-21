package services

import (
	commonDtos "protected_link/internal/common/api/dtos"
	apiDtos "protected_link/internal/modules/protected_link_generation/apis/dtos"
	"protected_link/internal/modules/protected_link_generation/repositories"
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
	return s.repo.GetTokenData(link)
}
