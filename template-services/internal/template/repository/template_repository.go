package repository

import (
	"context"
	"strings"
	"template-services/internal/models"
	"template-services/internal/pkg/db"
)

type TemplateRepository interface {
	Create(ctx context.Context, template *models.Template) error
	Get(ctx context.Context, id, name, channel, language *string) (*models.Template, error)
	Update(ctx context.Context, template *models.Template) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context) ([]models.Template, error)
}

type templateRepository struct {
	db *db.YugabyteDB
}

func NewTemplateRepository(db *db.YugabyteDB) TemplateRepository {
	return &templateRepository{db: db}
}

func (r *templateRepository) Create(ctx context.Context, template *models.Template) error {
	return r.db.Model(&models.Template{}).Create(template).Error
}

func (r *templateRepository) Get(ctx context.Context, id, name, channel, language *string) (*models.Template, error) {
	var conditions []string
	var args []interface{}

	if id != nil && *id != "" {
		conditions = append(conditions, "id = ?")
		args = append(args, *id)
	}
	if name != nil && *name != "" {
		conditions = append(conditions, "name = ?")
		args = append(args, *name)
	}
	if channel != nil && *channel != "" {
		conditions = append(conditions, "channel = ?")
		args = append(args, *channel)
	}
	if language != nil && *language != "" {
		conditions = append(conditions, "language = ?")
		args = append(args, *language)
	}

	// Always include is_active = true
	// conditions = append(conditions, "is_active = true")

	query := strings.Join(conditions, " AND ")

	var template models.Template
	err := r.db.Model(&models.Template{}).Where(query, args...).First(&template).Error
	return &template, err
}

func (r *templateRepository) Update(ctx context.Context, template *models.Template) error {
	return r.db.Model(&models.Template{}).
		Where("id = ?", template.ID).
		Updates(template).Error
}

func (r *templateRepository) Delete(ctx context.Context, id string) error {
	return r.db.Model(&models.Template{}).Where("id = ?", id).Update("is_active", false).Error
}

func (r *templateRepository) List(ctx context.Context) ([]models.Template, error) {
	var templates []models.Template
	err := r.db.Model(&models.Template{}).Find(&templates).Error
	return templates, err
}
