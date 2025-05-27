package mock

import (
	"messaging_service/pkg/logger"
	metrics "messaging_service/pkg/matrics"
	"messaging_service/pkg/tracer"
)

// MockObservabilityStack is a test implementation of the ObservabilityStack
// that uses mocks for its components
type MockObservabilityStack struct {
	MockTracer  *tracer.MockTracerService
	MockMetrics *metrics.MockMetricsService
	MockLogger  *logger.MockLogger
}

// NewMockObservabilityStack creates a new mock observability stack for testing
func NewMockObservabilityStack() *MockObservabilityStack {
	return &MockObservabilityStack{
		MockTracer:  tracer.NewMockTracerService(),
		MockMetrics: metrics.NewMockMetricsService(),
		MockLogger:  logger.NewMockLogger(),
	}
}

// GetObservabilityComponents returns the components of the mock observability stack
// as interfaces that match the actual ObservabilityStack
func (m *MockObservabilityStack) GetObservabilityComponents() (tracer.TracerService, metrics.MetricsService, logger.Logger) {
	return m.MockTracer, m.MockMetrics, m.MockLogger
}
