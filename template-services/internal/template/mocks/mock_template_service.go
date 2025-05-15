package mocks

import (
	"context"
	"template-services/internal/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// MockTemplateService implements the TemplateServiceInterface
type MockTemplateService struct {
	mock.Mock
}

// ListTemplates implements service.TemplateServiceInterface.
func (m *MockTemplateService) ListTemplates(ctx context.Context) ([]models.Template, error) {
	panic("unimplemented")
}

func (m *MockTemplateService) CreateTemplate(ctx context.Context, template *models.Template) (*models.Template, error) {
	args := m.Called(ctx, template)
	if tmpl, ok := args.Get(0).(*models.Template); ok {
		return tmpl, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockTemplateService) GetTemplate(ctx context.Context, id, name, channel, language string) (*models.Template, error) {
	args := m.Called(ctx, id, name, channel, language)
	if tmpl, ok := args.Get(0).(*models.Template); ok {
		return tmpl, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockTemplateService) GetTemplateByID(ctx context.Context, id uuid.UUID) (*models.Template, error) {
	args := m.Called(ctx, id)
	if tmpl, ok := args.Get(0).(*models.Template); ok {
		return tmpl, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockTemplateService) UpdateTemplate(ctx context.Context, template *models.Template) (*models.Template, error) {
	args := m.Called(ctx, template)
	if tmpl, ok := args.Get(0).(*models.Template); ok {
		return tmpl, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockTemplateService) DeleteTemplate(ctx context.Context, id string) (*models.Template, error) {
	args := m.Called(ctx, id)
	if tmpl, ok := args.Get(0).(*models.Template); ok {
		return tmpl, args.Error(1)
	}
	return nil, args.Error(1)
}
