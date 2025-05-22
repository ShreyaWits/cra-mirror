package repository

import (
	"context"
	// "errors"
	"testing"

	"template-services/internal/models"
	"template-services/pkg/db"
	"template-services/pkg/observability"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockDBModeler is a testify mock for db.DBModeler
type MockDBModeler struct {
	mock.Mock

}

func (m *MockDBModeler) Model(value interface{}) db.DBModeler {
	m.Called(value)
	return m
}
func (m *MockDBModeler) Create(value interface{}) db.DBModeler {
	m.Called(value)
	return m
}
func (m *MockDBModeler) Where(query interface{}, args ...interface{}) db.DBModeler {
	m.Called(query, args)
	return m
}
func (m *MockDBModeler) Updates(values interface{}) db.DBModeler {
	m.Called(values)
	return m
}
func (m *MockDBModeler) Update(column string, value interface{}) db.DBModeler {
	m.Called(column, value)
	return m
}
func (m *MockDBModeler) First(dest interface{}, conds ...interface{}) db.DBModeler {
	m.Called(dest, conds)
	return m
}
func (m *MockDBModeler) Find(dest interface{}, conds ...interface{}) db.DBModeler {
	m.Called(dest, conds)
	return m
}
func (m *MockDBModeler) Delete(value interface{}, conds ...interface{}) db.DBModeler {
	m.Called(value, conds)
	return m
}
func (m *MockDBModeler) Error() error {
	args := m.Called()
	return args.Error(0)
}

func TestCreate(t *testing.T) {
	mockDB := new(MockDBModeler)
	repo := NewTemplateRepository(mockDB,&observability.ObservabilityStack{})
	id := uuid.New()
	template := &models.Template{ID: id, Name: "Test"}

	mockDB.On("Model", mock.AnythingOfType("*models.Template")).Return(mockDB)
	mockDB.On("Create", template).Return(mockDB)
	mockDB.On("Error").Return(nil)

	err := repo.Create(context.Background(), template)
	assert.NoError(t, err)
	mockDB.AssertExpectations(t)
	mockDB.AssertExpectations(t)
}

func TestGet(t *testing.T) {
	mockDB := new(MockDBModeler)
	repo := NewTemplateRepository(mockDB,&observability.ObservabilityStack{})
	id := uuid.New()
	name := "Test"
	channel := "email"
	language := "en"
	template := &models.Template{
		ID:       id,
		Name:     name,
		Channel:  channel,
		Language: language,
	}

	mockDB.On("Model", mock.AnythingOfType("*models.Template")).Return(mockDB)
	mockDB.On("Where", mock.Anything, mock.Anything).Return(mockDB)
	mockDB.On("First", mock.AnythingOfType("*models.Template"), mock.Anything).Run(func(args mock.Arguments) {
		arg := args.Get(0).(*models.Template)
		*arg = *template
	}).Return(mockDB)
	mockDB.On("Error").Return(nil)

	idStr := id.String()
	result, err := repo.Get(context.Background(), &idStr, &name, &channel, &language)
	assert.NoError(t, err)
	assert.Equal(t, template.ID, result.ID)
	assert.Equal(t, template.Name, result.Name)
	assert.Equal(t, template.Channel, result.Channel)
	assert.Equal(t, template.Language, result.Language)
	mockDB.AssertExpectations(t)
	mockDB.AssertExpectations(t)
}

func TestUpdate(t *testing.T) {
	mockDB := new(MockDBModeler)
	repo := NewTemplateRepository(mockDB,&observability.ObservabilityStack{})
	id := uuid.New()
	template := &models.Template{ID: id, Name: "Updated"}

	mockDB.On("Model", mock.AnythingOfType("*models.Template")).Return(mockDB)
	mockDB.On("Where", "id = ?", []interface{}{template.ID}).Return(mockDB)
	mockDB.On("Updates", template).Return(mockDB)
	mockDB.On("Error").Return(nil)

	err := repo.Update(context.Background(), template)
	assert.NoError(t, err)
	mockDB.AssertExpectations(t)
	mockDB.AssertExpectations(t)
}

func TestDelete(t *testing.T) {
	mockDB := new(MockDBModeler)
	repo := NewTemplateRepository(mockDB,&observability.ObservabilityStack{})
	id := uuid.New()

	mockDB.On("Model", mock.AnythingOfType("*models.Template")).Return(mockDB)
	mockDB.On("Where", "id = ?", []interface{}{id.String()}).Return(mockDB)
	mockDB.On("Update", "is_active", false).Return(mockDB)
	mockDB.On("Error").Return(nil)

	err := repo.Delete(context.Background(), id.String())
	assert.NoError(t, err)
	mockDB.AssertExpectations(t)
	mockDB.AssertExpectations(t)
}

func TestList(t *testing.T) {
	mockDB := new(MockDBModeler)
	repo := NewTemplateRepository(mockDB,&observability.ObservabilityStack{})
	templates := []models.Template{
		{ID: uuid.New(), Name: "A"},
		{ID: uuid.New(), Name: "B"},
	}

	mockDB.On("Model", mock.AnythingOfType("*models.Template")).Return(mockDB)
	mockDB.On("Find", mock.AnythingOfType("*[]models.Template"), mock.Anything).Run(func(args mock.Arguments) {
		arg := args.Get(0).(*[]models.Template)
		*arg = templates
	}).Return(mockDB)
	mockDB.On("Error").Return(nil)

	result, err := repo.List(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, templates, result)
	mockDB.AssertExpectations(t)
	mockDB.AssertExpectations(t)
}

// func TestGet_Error(t *testing.T) {
// 	mockDB := new(MockDBModeler)
// 	repo := NewTemplateRepository(mockDB)
// 	id := uuid.New()
// 	idStr := id.String()

// 	mockDB.On("Model", mock.AnythingOfType("*models.Template")).Return(mockDB)
// 	mockDB.On("Where", mock.Anything, mock.Anything).Return(mockDB)
// 	mockDB.On("First", mock.AnythingOfType("*models.Template"), mock.Anything).Return(mockDB)
// 	mockDB.On("Error").Return(errors.New("record not found"))

// 	result, err := repo.Get(context.Background(), &idStr, nil, nil, nil)
// 	assert.Error(t, err)
// 	assert.Equal(t, "record not found", err.Error())
// 	assert.Nil(t, result)
// 	mockDB.AssertExpectations(t)
// 	mockDB.AssertExpectations(t)
// }
