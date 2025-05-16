package services

import (
	commonDtos "protected_link/internal/common/api/dtos"
	apiDtos "protected_link/internal/modules/link_generation/apis/dtos"
)

type GenerateLinkServiceInterface interface {
	SaveGeneratedLink(req *apiDtos.GenerateUrlRequest) (*commonDtos.ApiResponseDto, error)
	DeleteGeneratedLink(link string) (*commonDtos.ApiResponseDto, error)
	GetExtractData(token *string) (*commonDtos.ApiResponseDto, error)
}
