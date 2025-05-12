package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"template-services/internal/models"
	"template-services/internal/pkg/cache"
	"template-services/internal/template/repository"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TemplateServiceInterface interface {
	CreateTemplate(ctx context.Context, template *models.Template) (*models.Template, error)
	GetTemplate(ctx context.Context, id, name, channel, language string) (*models.Template, error)
	GetTemplateByID(ctx context.Context, id uuid.UUID) (*models.Template, error)
	UpdateTemplate(ctx context.Context, template *models.Template) (*models.Template, error)
	DeleteTemplate(ctx context.Context, id string) (*models.Template, error)
}

type TemplateService struct {
	repo  repository.TemplateRepository
	cache *cache.RedisCache
}

func NewTemplateService(repo repository.TemplateRepository, cache *cache.RedisCache) *TemplateService {
	return &TemplateService{
		repo:  repo,
		cache: cache,
	}
}

func (s *TemplateService) CreateTemplate(ctx context.Context, template *models.Template) (*models.Template, error) {
	if err := s.repo.Create(ctx, template); err != nil {
		return nil, fmt.Errorf("failed to create template: %w", err)
	}

	cacheKey := fmt.Sprintf("template:%s:%s:%s", template.Name, template.Channel, template.Language)
	if err := s.cache.Set(cacheKey, template); err != nil {
		log.Printf("Failed to cache template: %v", err)
	}

	return template, nil
}

func (s *TemplateService) GetTemplate(ctx context.Context, id, name, channel, language string) (*models.Template, error) {
	var idPtr *string
	if id != "" {
		idPtr = &id
	}
	cacheKey := fmt.Sprintf("template:%s:%s:%s", name, channel, language)

	// Try to get from cache first
	var template *models.Template
	if err := s.cache.Get(cacheKey, &template); err == nil {
		return template, nil
	}

	// If not in cache, get from database
	template, err := s.repo.Get(ctx, idPtr, &name, &channel, &language)
	if err != nil {
		return nil, fmt.Errorf("failed to get template: %w", err)
	}

	// Cache the result
	if err := s.cache.Set(cacheKey, template); err != nil {
		log.Printf("Failed to cache template: %v", err)
	}

	return template, nil
}

func (s *TemplateService) GetTemplateByID(ctx context.Context, id uuid.UUID) (*models.Template, error) {
	idStr := id.String()

	data, err := s.repo.Get(ctx, &idStr, nil, nil, nil)
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("template with ID %s not found", idStr)
	} else if err != nil {
		return nil, fmt.Errorf("failed to get template by ID")
	}

	return data, nil

}

func (s *TemplateService) UpdateTemplate(ctx context.Context, template *models.Template) (*models.Template, error) {
	// Update the template in the repository
	if err := s.repo.Update(ctx, template); err != nil {
		return template, fmt.Errorf("failed to update template: %w", err)
	}

	// Invalidate the cache
	cacheKey := fmt.Sprintf("template:%s:%s:%s", template.Name, template.Channel, template.Language)
	if err := s.cache.Delete(cacheKey); err != nil {
		log.Printf("Failed to invalidate cache for updated template: %v", err)
	}

	return template, nil
}

func (s *TemplateService) DeleteTemplate(ctx context.Context, id string) (*models.Template, error) {
	// Retrieve the template before deleting
	delTemplate, err := s.repo.Get(ctx, &id, nil, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve template before deletion: %w", err)
	}

	// Soft delete in the repository
	if err := s.repo.Delete(ctx, id); err != nil {
		return nil, fmt.Errorf("failed to delete template: %w", err)
	}

	return delTemplate, nil
}
