// mocks/repository.go

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

func (m *MockRepository) GetEtcdKey(ctx context.Context, key string) (string, error) {
	args := m.Called(ctx, key)
	return args.String(0), args.Error(1)
}

func (m *MockRepository) SetEtcdKey(ctx context.Context, key string, value string, ttl time.Duration) error {
	args := m.Called(ctx, key, value, ttl)
	return args.Error(0)
}

func (m *MockRepository) DeleteEtcdKey(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

func (m *MockRepository) StoreConfig(ctx context.Context, serviceName, environment string, configData map[string]interface{}) (interface{}, error) {
	args := m.Called(ctx, serviceName, environment, configData)
	return args.Get(0), args.Error(1)
}

func (m *MockRepository) GetConfig(ctx context.Context, serviceName, environment string) (map[string]interface{}, error) {
	args := m.Called(ctx, serviceName, environment)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func (m *MockRepository) GetAllKeys(ctx context.Context, prefix string) (map[string]string, error) {
	args := m.Called(ctx, prefix)
	return args.Get(0).(map[string]string), args.Error(1)
}

func (m *MockRepository) GetConfigMetadata(ctx context.Context, serviceName, environment string) (*models.ConfigMetadata, error) {
	args := m.Called(ctx, serviceName, environment)
	return args.Get(0).(*models.ConfigMetadata), args.Error(1)
}

func (m *MockRepository) GetConfigValue(ctx context.Context, serviceName, environment, key string) (interface{}, error) {
	args := m.Called(ctx, serviceName, environment, key)
	return args.Get(0), args.Error(1)
}

// Mock implementation of CreateAdmin
func (m *MockRepository) CreateAdmin(ctx context.Context, admin *models.Admin) (*models.Admin, error) {
	args := m.Called(ctx, admin)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Admin), args.Error(1)
}

func (m *MockRepository) GetAdminByCredentials(ctx context.Context, username, password string) (*models.Admin, error) {
	args := m.Called(ctx, username, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Admin), args.Error(1)
}

func (m *MockRepository) NotifyWebhook(ctx context.Context, url, method string) (*models.Admin, error) {
	args := m.Called(ctx, url, method)
	return args.Get(0).(*models.Admin), args.Error(1)
}
