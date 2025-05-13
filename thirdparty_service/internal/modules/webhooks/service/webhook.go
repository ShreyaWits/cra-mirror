package services

import (
	"thirdparty_service/internal/models"
	repositories "thirdparty_service/internal/modules/webhooks/repository"
)

type WebhookService interface {
	ProcessWebhook(key string, webhook *models.Webhook) error
}

type webhookService struct {
	repo repositories.WebhookRepository
}

func NewWebhookService(repo repositories.WebhookRepository) WebhookService {
	return &webhookService{repo: repo}
}

func (s *webhookService) ProcessWebhook(key string, webhook *models.Webhook) error {
	return s.repo.SaveWebhook(key, webhook)
}
