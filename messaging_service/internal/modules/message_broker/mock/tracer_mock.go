package mock

import (
	"context"

	"github.com/stretchr/testify/mock"
)

// MockTracerService implements a mock tracer
type MockTracerService struct {
	mock.Mock
}

func (m *MockTracerService) StartTracer(ctx context.Context, name string) (context.Context, interface{}) {
	args := m.Called(ctx, name)
	return args.Get(0).(context.Context), args.Get(1)
}

func (m *MockTracerService) StopSpan(span interface{}) {
	m.Called(span)
}

func (m *MockTracerService) SetAttributes(span interface{}, attrs map[string]string) {
	m.Called(span, attrs)
}

func (m *MockTracerService) SetStatus(span interface{}, code int, message string) {
	m.Called(span, code, message)
}

func (m *MockTracerService) RecordError(span interface{}, err error) {
	m.Called(span, err)
}
