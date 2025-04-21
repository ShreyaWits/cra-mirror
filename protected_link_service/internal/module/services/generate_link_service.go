package services

import (
	commonDtos "protected_link/internal/common/api/dtos"

	apiDtos "protected_link/internal/module/apis/dtos"
	"protected_link/internal/module/repositories"
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
