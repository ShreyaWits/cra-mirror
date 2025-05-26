package metrics

import (
	"context"
	"sync"
)

// CounterCall represents a call to IncrementCounter
type CounterCall struct {
	Name  string
	Value int64
	Attrs map[string]string
	Ctx   context.Context
}

// HistogramCall represents a call to RecordHistogram
type HistogramCall struct {
	Name  string
	Value float64
	Attrs map[string]string
	Ctx   context.Context
}

// MockMetricsService is a mock implementation of MetricsService for testing
type MockMetricsService struct {
	mu             sync.Mutex
	CounterCalls   []CounterCall
	HistogramCalls []HistogramCall
}

// NewMockMetricsService creates a new mock metrics service
func NewMockMetricsService() *MockMetricsService {
	return &MockMetricsService{
		CounterCalls:   []CounterCall{},
		HistogramCalls: []HistogramCall{},
	}
}

// IncrementCounter records a counter call
func (m *MockMetricsService) IncrementCounter(ctx context.Context, name string, value int64, attrs map[string]string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Make a copy of the attrs map to avoid test interference
	attrsCopy := make(map[string]string)
	if attrs != nil {
		for k, v := range attrs {
			attrsCopy[k] = v
		}
	}

	m.CounterCalls = append(m.CounterCalls, CounterCall{
		Name:  name,
		Value: value,
		Attrs: attrsCopy,
		Ctx:   ctx,
	})
}

// RecordHistogram records a histogram call
func (m *MockMetricsService) RecordHistogram(ctx context.Context, name string, value float64, attrs map[string]string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Make a copy of the attrs map to avoid test interference
	attrsCopy := make(map[string]string)
	if attrs != nil {
		for k, v := range attrs {
			attrsCopy[k] = v
		}
	}

	m.HistogramCalls = append(m.HistogramCalls, HistogramCall{
		Name:  name,
		Value: value,
		Attrs: attrsCopy,
		Ctx:   ctx,
	})
}
