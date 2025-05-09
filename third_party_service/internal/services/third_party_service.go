package services

import "third_party_service/internal/repositories"

type ThirdPartyService struct {
	repo repositories.ThirdPartyRepository
}

func NewThirdPartyService(repo repositories.ThirdPartyRepository) *ThirdPartyService {
	return &ThirdPartyService{repo: repo}
}

func (s *ThirdPartyService) CreateSendEmail(req interface{}) (interface{}, error) {
	return nil, nil
}
