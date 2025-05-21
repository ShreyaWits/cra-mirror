package mock

import (
	"messaging_service/pkg/logger"
)

// MockObservabilityStack is a mock implementation of the observability stack
type MockObservabilityStack struct {
	LoggerService  *MockLogger
	TracerService  *MockTracerService
	MetricsService interface{}
}

// NewMockObservabilityStack creates a new mock observability stack
func NewMockObservabilityStack() *MockObservabilityStack {
	return &MockObservabilityStack{
		LoggerService: new(MockLogger),
		TracerService: new(MockTracerService),
	}
}

// GetLoggerService returns the logger service
func (m *MockObservabilityStack) GetLoggerService() logger.Logger {
	return m.LoggerService
}
