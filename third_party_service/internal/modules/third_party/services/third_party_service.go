package services

import (
	"third_party_service/internal/modules/third_party/api/schemas"
	"third_party_service/internal/modules/third_party/repositories"
)

type ThirdPartyService struct {
	repo repositories.ThirdPartyRepository
}

func NewThirdPartyService(repo repositories.ThirdPartyRepository) *ThirdPartyService {
	return &ThirdPartyService{repo: repo}
}

func (s *ThirdPartyService) CreateSendEmail(req schemas.CreateEmailRequest) (schemas.CreateEmailResponse, error) {
	return schemas.CreateEmailResponse{}, nil
}
