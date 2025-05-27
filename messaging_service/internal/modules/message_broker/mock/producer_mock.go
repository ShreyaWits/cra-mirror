package mock

import (
	"context"

	"github.com/golang/mock/gomock"
)

// MockProducerWithPublish is a mock implementation that adds a Publish method
// for backward compatibility with tests
type MockProducerWithPublish struct {
	*MockProducer
}

// NewMockProducerWithPublish creates a new mock producer instance with Publish method
func NewMockProducerWithPublish(ctrl *gomock.Controller) *MockProducerWithPublish {
	return &MockProducerWithPublish{
		MockProducer: NewMockProducer(ctrl),
	}
}

// Publish is a convenience method that delegates to WriteWithRetry
// This is used for backward compatibility with tests that expect a Publish method
func (m *MockProducerWithPublish) Publish(ctx context.Context, key, value []byte) error {
	return m.WriteWithRetry(ctx, key, value, nil)
}
