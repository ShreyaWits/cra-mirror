package mocks

import (
	"context"
	"nps-config-service/internal/modules/config-manager/models"
	"time"

	"github.com/stretchr/testify/mock"
)

type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) StoreConfig(ctx context.Context, serviceName, environment string, configData map[string]interface{}) (interface{}, error) {
	args := m.Called(ctx, serviceName, environment, configData)
	return args.Get(0), args.Error(1)
}

func (m *MockRepository) GetConfig(ctx context.Context, serviceName, environment string) (map[string]interface{}, error) {
	args := m.Called(ctx, serviceName, environment)
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func (m *MockRepository) GetConfigValue(ctx context.Context, serviceName, environment, key string) (interface{}, error) {
	args := m.Called(ctx, serviceName, environment, key)
	return args.Get(0), args.Error(1)
}

func (m *MockRepository) SetEtcdKey(ctx context.Context, key string, data string, ttl time.Duration) error {
	args := m.Called(ctx, key, data, ttl)
	return args.Error(0)
}

func (m *MockRepository) GetEtcdKey(ctx context.Context, key string) (string, error) {
	args := m.Called(ctx, key)
	return args.String(0), args.Error(1)
}

func (m *MockRepository) DeleteEtcdKey(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

func (m *MockRepository) CreateAdmin(ctx context.Context, admin *models.Admin) (*models.Admin, error) {
	args := m.Called(ctx, admin)
	return args.Get(0).(*models.Admin), args.Error(1)
}

func (m *MockRepository) GetAdminByCredentials(ctx context.Context, username, password string) (*models.Admin, error) {
	args := m.Called(ctx, username, password)
	return args.Get(0).(*models.Admin), args.Error(1)
}

func (m *MockRepository) ListAdmins(ctx context.Context) ([]*models.Admin, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*models.Admin), args.Error(1)
}

func (m *MockRepository) DeleteAdmin(ctx context.Context, username string) error {
	args := m.Called(ctx, username)
	return args.Error(0)
}
