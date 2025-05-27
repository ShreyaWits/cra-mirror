package mocks

import (
	"github.com/stretchr/testify/mock"
)

type MockConfigProvider struct {
	mock.Mock
	adminSecret string
	jwtSecret   string
}

func NewMockConfigProvider(adminSecret, jwtSecret string) *MockConfigProvider {
	return &MockConfigProvider{
		adminSecret: adminSecret,
		jwtSecret:   jwtSecret,
	}
}

func (m *MockConfigProvider) GetAdminSecret() string {
	return m.adminSecret
}

func (m *MockConfigProvider) GetJWTSecret() string {
	return m.jwtSecret
}
