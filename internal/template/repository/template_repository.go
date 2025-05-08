package repository

import (
	"context"
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

	var query string
	var args []interface{}
	if id != nil {
		query = "id = ?"
		args = append(args, *id)
	} else if name != nil {
		query = "AND name = ?"
		args = append(args, *name)
	} else if channel != nil {
		query = "AND channel = ?"
		args = append(args, *channel)
	} else if language != nil {
		query = "AND language = ?"
		args = append(args, *language)
	}

	query += " AND is_active = true"

	var template models.Template
	err := r.db.Model(&models.Template{}).Where(query,
		args...).First(&template).Error
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
