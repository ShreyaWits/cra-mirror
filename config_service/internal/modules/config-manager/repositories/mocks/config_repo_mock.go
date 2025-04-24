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

func (m *MockRepository) Get(ctx context.Context, key string) (string, error) {
	args := m.Called(ctx, key)
	return args.String(0), args.Error(1)
}

func (m *MockRepository) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	args := m.Called(ctx, key, value, ttl)
	return args.Error(0)
}

func (m *MockRepository) Delete(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}
func (m *MockRepository) StoreConfig(serviceName, environment string, configData map[string]interface{}) (interface{}, error) {
	args := m.Called(serviceName, environment, configData)
	return args.Get(0), args.Error(1)
}

func (m *MockRepository) GetConfig(serviceName, environment string) (map[string]interface{}, error) {
	args := m.Called(serviceName, environment)
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func (m *MockRepository) GetAllKeys(prefix string) (map[string]string, error) {
	args := m.Called(prefix)
	return args.Get(0).(map[string]string), args.Error(1)
}

func (m *MockRepository) GetConfigMetadata(serviceName, environment string) (*models.ConfigMetadata, error) {
	args := m.Called(serviceName, environment)
	return args.Get(0).(*models.ConfigMetadata), args.Error(1)
}

func (m *MockRepository) GetConfigValue(serviceName, environment, key string) (interface{}, error) {
	args := m.Called(serviceName, environment, key)
	return args.Get(0), args.Error(1)
}
