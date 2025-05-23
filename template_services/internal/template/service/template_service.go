package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	configEnv "template-services/internal/configs"
	"template-services/internal/constants"
	"template-services/internal/models"
	"template-services/internal/template/repository"
	cacheclient "template-services/pkg/client/cache_client"
	"template-services/pkg/observability"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Interface renamed for clarity
type TemplateServiceInterface interface {
	CreateTemplate(ctx context.Context, template *models.Template) (*models.Template, error)
	GetTemplate(ctx context.Context, id, name, channel, language string) (*models.Template, error)
	GetTemplateByID(ctx context.Context, id uuid.UUID) (*models.Template, error)
	UpdateTemplate(ctx context.Context, template *models.Template) (*models.Template, error)
	DeleteTemplate(ctx context.Context, id string) (*models.Template, error)
	ListTemplates(ctx context.Context) ([]models.Template, error)
}

// Struct renamed to avoid conflict
type TemplateServiceImpl struct {
	repo  repository.TemplateRepository
	cache *cacheclient.RedisClientStruct
	obs   *observability.ObservabilityStack
}

func NewTemplateService(repo repository.TemplateRepository, cache *cacheclient.RedisClientStruct, obs *observability.ObservabilityStack, config *configEnv.Config) *TemplateServiceImpl {
	return &TemplateServiceImpl{
		repo:  repo,
		cache: cache,
		obs:   obs,
	}
}

func (s *TemplateServiceImpl) CreateTemplate(ctx context.Context, template *models.Template) (*models.Template, error) {
	tCtx, span := s.obs.TracerService.StartTracer(ctx, "CreateTemplate")
	defer s.obs.TracerService.StopSpan(span)
	if err := s.repo.Create(tCtx, template); err != nil {
		return nil, fmt.Errorf("failed to create template: %w", err)
	}
	cacheKey := fmt.Sprintf("template:%s:%s:%s", template.Name, template.Channel, template.Language)
	bytes, err := json.Marshal(template)
	if err != nil {
		return nil, err
	}

	if errCache := s.cache.SetCache(tCtx, constants.ServiceName, cacheKey, string(bytes), 240*time.Hour); errCache != nil {
		log.Printf("Failed to cache template: %v", errCache)
	}
	return template, nil
}

func (s *TemplateServiceImpl) GetTemplate(ctx context.Context, id, name, channel, language string) (*models.Template, error) {
	tCtx, span := s.obs.TracerService.StartTracer(ctx, "GetTemplate")
	defer s.obs.TracerService.StopSpan(span)

	log.Printf("GetTemplate called with id=%s, name=%s, channel=%s, language=%s", id, name, channel, language)

	var idPtr *string
	if id != "" {
		idPtr = &id
	}

	cacheKey := fmt.Sprintf("template:%s:%s:%s", name, channel, language)
	log.Printf("Generated cache key: %s", cacheKey)

	var template *models.Template
	value, exist, err := s.cache.GetCache(tCtx, constants.ServiceName, cacheKey)
	if err != nil {
		log.Printf("Error retrieving from cache: %v", err)
	}

	if exist {
		log.Printf("Cache hit for key: %s", cacheKey)
		if err := json.Unmarshal([]byte(value), &template); err != nil {
			log.Printf("Failed to unmarshal template from cache: %v", err)
			return nil, fmt.Errorf("failed to unmarshal template from cache: %w", err)
		}
		log.Printf("Successfully returned template from cache")
		return template, nil
	}

	log.Printf("Cache miss. Fetching template from DB...")

	data, dbErr := s.repo.Get(tCtx, idPtr, &name, &channel, &language)
	log.Print("Fetching template from DB")
	if dbErr != nil {
		log.Printf("Failed to retrieve template from DB: %v", dbErr)
		return nil, fmt.Errorf("failed to get template: %w", dbErr)
	}

	bytes, err := json.Marshal(data)
	if err != nil {
		log.Printf("Failed to marshal template for caching: %v", err)
		return nil, err
	}

	if errCache := s.cache.SetCache(tCtx, constants.ServiceName, cacheKey, string(bytes), 240*time.Hour); errCache != nil {
		log.Printf("Failed to cache template: %v", errCache)
	} else {
		log.Printf("Template cached successfully with key: %s", cacheKey)
	}

	log.Printf("Returning template fetched from DB")
	return data, nil
}

func (s *TemplateServiceImpl) GetTemplateByID(ctx context.Context, id uuid.UUID) (*models.Template, error) {
	tCtx, span := s.obs.TracerService.StartTracer(ctx, "GetTemplateByID")
	defer s.obs.TracerService.StopSpan(span)
	idStr := id.String()

	data, err := s.repo.Get(tCtx, &idStr, nil, nil, nil)
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("template with ID %s not found", idStr)
	} else if err != nil {
		return nil, fmt.Errorf("failed to get template by ID")
	}

	return data, nil
}

func (s *TemplateServiceImpl) UpdateTemplate(ctx context.Context, template *models.Template) (*models.Template, error) {
	tCtx, span := s.obs.TracerService.StartTracer(ctx, "UpdateTemplate")
	defer s.obs.TracerService.StopSpan(span)
	if err := s.repo.Update(tCtx, template); err != nil {
		return template, fmt.Errorf("failed to update template: %w", err)
	}
	cacheKey := fmt.Sprintf("template:%s:%s:%s", template.Name, template.Channel, template.Language)
	if err := s.cache.InvalidateCache(tCtx, constants.ServiceName, cacheKey); err != nil {
		log.Printf("Failed to invalidate cache for updated template: %v", err)
	}
	return template, nil
}

func (s *TemplateServiceImpl) DeleteTemplate(ctx context.Context, id string) (*models.Template, error) {
	tCtx, span := s.obs.TracerService.StartTracer(ctx, "DeleteTemplate")
	defer s.obs.TracerService.StopSpan(span)
	delTemplate, err := s.repo.Get(tCtx, &id, nil, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve template before deletion: %w", err)
	}
	if err := s.repo.Delete(tCtx, id); err != nil {
		return nil, fmt.Errorf("failed to delete template: %w", err)
	}
	return delTemplate, nil
}

// ListTemplates returns all templates from the repository.
func (s *TemplateServiceImpl) ListTemplates(ctx context.Context) ([]models.Template, error) {
	tCtx, span := s.obs.TracerService.StartTracer(ctx, "DeleteTemplate")
	defer s.obs.TracerService.StopSpan(span)
	templates, err := s.repo.List(tCtx)
	if err != nil {
		return nil, fmt.Errorf("failed to list templates: %w", err)
	}
	return templates, nil
}
