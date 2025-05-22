package mock

import (
	"context"
	"messaging_service/pkg/logger"

	"github.com/stretchr/testify/mock"
)

// MockLogger implements the logger interface for testing
type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) Info(ctx context.Context, args ...interface{}) {
	m.Called(append([]interface{}{ctx}, args...)...)
}

func (m *MockLogger) Error(ctx context.Context, args ...interface{}) {
	m.Called(append([]interface{}{ctx}, args...)...)
}

func (m *MockLogger) Debug(ctx context.Context, args ...interface{}) {
	m.Called(append([]interface{}{ctx}, args...)...)
}

func (m *MockLogger) Warn(ctx context.Context, args ...interface{}) {
	m.Called(append([]interface{}{ctx}, args...)...)
}

func (m *MockLogger) WithFields(fields map[string]interface{}) logger.Logger {
	return m
}

func (m *MockLogger) Sync() error {
	return nil
}
