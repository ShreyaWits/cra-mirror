package confluent

// KafkaFactory defines the factory interface for creating Confluent Kafka components
type KafkaFactory interface {
	// CreateProducer creates a new producer
	CreateProducer(cfg KafkaConfig) (Producer, error)

	// CreateConsumer creates a new consumer
	CreateConsumer(cfg KafkaConfig, groupID string, handler func([]byte) error) (Consumer, error)

	// CreateAdmin creates a new admin client
	CreateAdmin(cfg KafkaConfig) (KafkaAdmin, error)

	// CreateDLQProducer creates a new DLQ producer
	CreateDLQProducer(cfg KafkaConfig) (DLQProducer, error)
}

// Factory provides methods for creating Confluent Kafka components
type Factory struct{}

// NewFactory creates a new Factory
func NewFactory() KafkaFactory {
	return &Factory{}
}

// CreateProducer creates a new Confluent Producer
func (f *Factory) CreateProducer(cfg KafkaConfig) (Producer, error) {
	return NewProducer(cfg)
}

// CreateConsumer creates a new Confluent Consumer
func (f *Factory) CreateConsumer(cfg KafkaConfig, groupID string, handler func([]byte) error) (Consumer, error) {
	return NewConsumer(cfg, groupID, handler)
}

// CreateAdmin creates a new Confluent AdminClient
func (f *Factory) CreateAdmin(cfg KafkaConfig) (KafkaAdmin, error) {
	return NewAdmin(cfg)
}

// CreateDLQProducer creates a new Confluent DLQ Producer
func (f *Factory) CreateDLQProducer(cfg KafkaConfig) (DLQProducer, error) {
	return NewDLQProducer(cfg)
}
