package mocks

import (
	"nps-config-service/internal/modules/config-manager/apis/dtos"

	"github.com/stretchr/testify/mock"
)

type MockWebhookService struct {
	mock.Mock
}

func (m *MockWebhookService) RegisterWebhookService(configData dtos.RegisterWebhookRequest) (interface{}, error) {
	args := m.Called(configData)
	return args.Get(0), args.Error(1)
}

func (m *MockWebhookService) GetWebhooks(env, service string) ([]dtos.RegisterWebhookRequest, error) {
	args := m.Called(env, service)
	return args.Get(0).([]dtos.RegisterWebhookRequest), args.Error(1)
}

func (m *MockWebhookService) DeleteWebhook(env, service, url, method string) (string, error) {
	args := m.Called(env, service, url, method)
	return args.String(0), args.Error(1)
}

func (m *MockWebhookService) DeleteAllWebhooks(env, service string) error {
	args := m.Called(env, service)
	return args.Error(0)
}

func (m *MockWebhookService) NotifyWebhook(hook dtos.RegisterWebhookRequest, data map[string]interface{}) {
	m.Called(hook, data)
}
