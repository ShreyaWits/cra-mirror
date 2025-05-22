package repository

import (
	"context"
	"fmt"
	"log"
	"strings"
	"template-services/internal/models"
	"template-services/pkg/db"
	"template-services/pkg/observability"
)

type TemplateRepository interface {
	Create(ctx context.Context, template *models.Template) error
	Get(ctx context.Context, id, name, channel, language *string) (*models.Template, error)
	Update(ctx context.Context, template *models.Template) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context) ([]models.Template, error)
}

type templateRepository struct {
	db  db.DBModeler
	obs *observability.ObservabilityStack
}

func NewTemplateRepository(db db.DBModeler, obs *observability.ObservabilityStack) TemplateRepository {
	return &templateRepository{db: db, obs: obs}
}

func (r *templateRepository) Create(ctx context.Context, template *models.Template) error {
	_, span := r.obs.TracerService.StartTracer(ctx, "Create")
	defer r.obs.TracerService.StopSpan(span)
	return r.db.Model(&models.Template{}).Create(template).Error()
}

func (r *templateRepository) Get(ctx context.Context, id, name, channel, language *string) (*models.Template, error) {
	log.Print("GetTemplate called with id=", id, ", name=", name, ", channel=", channel, ", language=", language)
	fmt.Println(r.obs)
	_, span := r.obs.TracerService.StartTracer(ctx, "Get")
	defer r.obs.TracerService.StopSpan(span)
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

	err := r.db.Model(&models.Template{}).Where(query, args...).First(&template).Error()
	log.Print("Query executed: ", query, " with args: ", args)
	return &template, err
}

func (r *templateRepository) Update(ctx context.Context, template *models.Template) error {
	_, span := r.obs.TracerService.StartTracer(ctx, "Update")
	defer r.obs.TracerService.StopSpan(span)
	result := r.db.Model(&models.Template{}).Where("id = ?", template.ID).Updates(template).Error()
	return result
}

func (r *templateRepository) Delete(ctx context.Context, id string) error {
	_, span := r.obs.TracerService.StartTracer(ctx, "Delete")
	defer r.obs.TracerService.StopSpan(span)
	return r.db.Model(&models.Template{}).Where("id = ?", id).Update("is_active", false).Error()
}

func (r *templateRepository) List(ctx context.Context) ([]models.Template, error) {
	_, span := r.obs.TracerService.StartTracer(ctx, "List")
	defer r.obs.TracerService.StopSpan(span)
	var templates []models.Template
	err := r.db.Model(&models.Template{}).Find(&templates).Error()
	return templates, err
}
