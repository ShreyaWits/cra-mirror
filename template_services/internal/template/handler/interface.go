package handler

import (
	"context"
	"template-services/internal/models"

	"github.com/google/uuid"
)

type TemplateServiceInterface interface {
	CreateTemplate(ctx context.Context, template *models.Template) (*models.Template, error)
	GetTemplate(ctx context.Context, id, name, channel, language string) (*models.Template, error)
	GetTemplateByID(ctx context.Context, id uuid.UUID) (*models.Template, error)
	UpdateTemplate(ctx context.Context, template *models.Template) (*models.Template, error)
	DeleteTemplate(ctx context.Context, id string) (*models.Template, error)
}
