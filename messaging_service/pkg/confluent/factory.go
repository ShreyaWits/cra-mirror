

package confluent

import (
	"context"
	"fmt"
	"messaging_service/internal/config"
	"messaging_service/pkg/observability"
)

// Factory provides methods for creating Confluent Kafka components
type Factory struct {
	obs *observability.ObservabilityStack
}

// NewFactory creates a new Factory
func NewFactory(obs *observability.ObservabilityStack) (KafkaFactory, error) {
	// Validate observability stack
	if obs == nil {
		return nil, fmt.Errorf("observability stack cannot be nil in Kafka factory")
	}

	// Validate required observability components
	if obs.LoggerService == nil {
		return nil, fmt.Errorf("logger service cannot be nil in Kafka factory")
	}

	return &Factory{obs: obs}, nil
}

// CreateProducer creates a new Confluent Producer
func (f *Factory) CreateProducer(ctx context.Context, cfg KafkaConfig) (Producer, error) {
	return NewProducer(ctx, cfg, f.obs)
}

// CreateConsumer creates a new Confluent Consumer
func (f *Factory) CreateConsumer(ctx context.Context, cfg KafkaConfig, groupID string, handler func([]byte) error) (Consumer, error) {
	return NewConsumer(cfg, groupID, handler, f.obs)
}

// CreateAdmin creates a new Confluent AdminClient
func (f *Factory) CreateAdmin(cfg *config.Config) (KafkaAdmin, error) {
	return NewAdmin(cfg, f.obs)
}

// CreateDLQProducer creates a new Confluent DLQ Producer
func (f *Factory) CreateDLQProducer(cfg KafkaConfig) (DLQProducer, error) {
	return NewDLQProducer(cfg, f.obs)
}
