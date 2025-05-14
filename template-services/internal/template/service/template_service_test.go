package service

import (
	"context"
	"errors"
	"fmt"
	"template-services/internal/models"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// --- Mocks ---

type MockRepo struct {
	mock.Mock
}

// List implements repository.TemplateRepository.
func (m *MockRepo) List(ctx context.Context) ([]models.Template, error) {
	panic("unimplemented")
}

func (m *MockRepo) Create(ctx context.Context, template *models.Template) error {
	args := m.Called(ctx, template)
	return args.Error(0)
}

func (m *MockRepo) Get(ctx context.Context, id *string, name, channel, language *string) (*models.Template, error) {
	args := m.Called(ctx, id, name, channel, language)
	return args.Get(0).(*models.Template), args.Error(1)
}

func (m *MockRepo) Update(ctx context.Context, template *models.Template) error {
	args := m.Called(ctx, template)
	return args.Error(0)
}

func (m *MockRepo) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

type MockCache struct {
	mock.Mock
}

func (m *MockCache) Set(key string, value interface{}, ttl time.Duration) error {
	args := m.Called(key, value, ttl)
	return args.Error(0)
}

func (m *MockCache) Get(key string, dest interface{}) error {
	args := m.Called(key, dest)
	return args.Error(0)
}

func (m *MockCache) Delete(key string) error {
	args := m.Called(key)
	return args.Error(0)
}

// --- Tests ---

func TestCreateTemplate_Success(t *testing.T) {
	mockRepo := new(MockRepo)
	mockCache := new(MockCache)
	svc := NewTemplateService(mockRepo, mockCache)

	template := &models.Template{
		ID:       uuid.New(),
		Name:     "test",
		Channel:  "email",
		Language: "en",
	}

	mockRepo.On("Create", mock.Anything, template).Return(nil)
	mockCache.On("Set", mock.Anything, template, time.Duration(0)).Return(nil)

	result, err := svc.CreateTemplate(context.Background(), template)
	assert.NoError(t, err)
	assert.Equal(t, template, result)
	mockRepo.AssertExpectations(t)
	mockCache.AssertExpectations(t)
}

func TestCreateTemplate_RepoError(t *testing.T) {
	mockRepo := new(MockRepo)
	mockCache := new(MockCache)
	svc := NewTemplateService(mockRepo, mockCache)

	template := &models.Template{
		ID:       uuid.New(),
		Name:     "fail",
		Channel:  "sms",
		Language: "fr",
	}

	mockRepo.On("Create", mock.Anything, template).Return(errors.New("repo error"))

	result, err := svc.CreateTemplate(context.Background(), template)
	assert.Nil(t, result)
	assert.Error(t, err)
	mockRepo.AssertExpectations(t)
}

func TestGetTemplate_CacheHit(t *testing.T) {
	mockRepo := new(MockRepo)
	mockCache := new(MockCache)
	svc := NewTemplateService(mockRepo, mockCache)

	template := &models.Template{
		ID:       uuid.New(),
		Name:     "cached",
		Channel:  "push",
		Language: "es",
	}

	cacheKey := fmt.Sprintf("template:%s:%s:%s", template.Name, template.Channel, template.Language)
	mockCache.On("Get", cacheKey, mock.AnythingOfType("**models.Template")).Run(func(args mock.Arguments) {
		arg := args.Get(1).(**models.Template)
		*arg = template
	}).Return(nil)

	result, err := svc.GetTemplate(context.Background(), "", template.Name, template.Channel, template.Language)
	assert.NoError(t, err)
	assert.Equal(t, template, result)
	mockCache.AssertExpectations(t)
}

func TestGetTemplate_CacheMiss_RepoHit(t *testing.T) {
	mockRepo := new(MockRepo)
	mockCache := new(MockCache)
	svc := NewTemplateService(mockRepo, mockCache)

	template := &models.Template{
		ID:       uuid.New(),
		Name:     "notcached",
		Channel:  "web",
		Language: "de",
	}

	cacheKey := fmt.Sprintf("template:%s:%s:%s", template.Name, template.Channel, template.Language)
	mockCache.On("Get", cacheKey, mock.AnythingOfType("**models.Template")).Return(errors.New("cache miss"))
	mockRepo.On("Get", mock.Anything, (*string)(nil), &template.Name, &template.Channel, &template.Language).Return(template, nil)
	mockCache.On("Set", cacheKey, template, time.Duration(0)).Return(nil)

	result, err := svc.GetTemplate(context.Background(), "", template.Name, template.Channel, template.Language)
	assert.NoError(t, err)
	assert.Equal(t, template, result)
	mockRepo.AssertExpectations(t)
	mockCache.AssertExpectations(t)
}

func TestGetTemplateByID_Success(t *testing.T) {
	mockRepo := new(MockRepo)
	mockCache := new(MockCache)
	svc := NewTemplateService(mockRepo, mockCache)

	id := uuid.New()
	idStr := id.String()
	template := &models.Template{
		ID: id,
	}

	mockRepo.On("Get", mock.Anything, &idStr, (*string)(nil), (*string)(nil), (*string)(nil)).Return(template, nil)

	result, err := svc.GetTemplateByID(context.Background(), id)
	assert.NoError(t, err)
	assert.Equal(t, template, result)
	mockRepo.AssertExpectations(t)
}

func TestGetTemplateByID_NotFound(t *testing.T) {
	mockRepo := new(MockRepo)
	mockCache := new(MockCache)
	svc := NewTemplateService(mockRepo, mockCache)

	id := uuid.New()
	idStr := id.String()

	mockRepo.On("Get", mock.Anything, &idStr, (*string)(nil), (*string)(nil), (*string)(nil)).Return(&models.Template{}, gorm.ErrRecordNotFound)

	result, err := svc.GetTemplateByID(context.Background(), id)
	assert.Nil(t, result)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
	mockRepo.AssertExpectations(t)
}

func TestUpdateTemplate_Success(t *testing.T) {
	mockRepo := new(MockRepo)
	mockCache := new(MockCache)
	svc := NewTemplateService(mockRepo, mockCache)

	template := &models.Template{
		ID:       uuid.New(),
		Name:     "update",
		Channel:  "email",
		Language: "en",
	}

	cacheKey := fmt.Sprintf("template:%s:%s:%s", template.Name, template.Channel, template.Language)
	mockRepo.On("Update", mock.Anything, template).Return(nil)
	mockCache.On("Delete", cacheKey).Return(nil)

	result, err := svc.UpdateTemplate(context.Background(), template)
	assert.NoError(t, err)
	assert.Equal(t, template, result)
	mockRepo.AssertExpectations(t)
	mockCache.AssertExpectations(t)
}

func TestUpdateTemplate_RepoError(t *testing.T) {
	mockRepo := new(MockRepo)
	mockCache := new(MockCache)
	svc := NewTemplateService(mockRepo, mockCache)

	template := &models.Template{
		ID:       uuid.New(),
		Name:     "failupdate",
		Channel:  "sms",
		Language: "fr",
	}

	mockRepo.On("Update", mock.Anything, template).Return(errors.New("update error"))

	result, err := svc.UpdateTemplate(context.Background(), template)
	assert.Error(t, err)
	assert.Equal(t, template, result)
	mockRepo.AssertExpectations(t)
}

func TestDeleteTemplate_Success(t *testing.T) {
	mockRepo := new(MockRepo)
	mockCache := new(MockCache)
	svc := NewTemplateService(mockRepo, mockCache)

	id := uuid.New().String()
	template := &models.Template{
		ID: uuid.MustParse(id),
	}

	mockRepo.On("Get", mock.Anything, &id, (*string)(nil), (*string)(nil), (*string)(nil)).Return(template, nil)
	mockRepo.On("Delete", mock.Anything, id).Return(nil)

	result, err := svc.DeleteTemplate(context.Background(), id)
	assert.NoError(t, err)
	assert.Equal(t, template, result)
	mockRepo.AssertExpectations(t)
}

func TestDeleteTemplate_RepoError(t *testing.T) {
	mockRepo := new(MockRepo)
	mockCache := new(MockCache)
	svc := NewTemplateService(mockRepo, mockCache)

	id := uuid.New().String()
	template := &models.Template{
		ID: uuid.MustParse(id),
	}

	mockRepo.On("Get", mock.Anything, &id, (*string)(nil), (*string)(nil), (*string)(nil)).Return(template, nil)
	mockRepo.On("Delete", mock.Anything, id).Return(errors.New("delete error"))

	result, err := svc.DeleteTemplate(context.Background(), id)
	assert.Nil(t, result)
	assert.Error(t, err)
	mockRepo.AssertExpectations(t)
}
