package kafka

import (
	pb "cra-protos/messaging_service"
	"time"
)

// Configurator handles all Kafka-specific configuration logic
type Configurator struct {
	brokers []string
}

// NewConfigurator creates a new Kafka Configurator
func NewConfigurator(brokers []string) *Configurator {
	return &Configurator{
		brokers: brokers,
	}
}

// CreatePublishConfig creates Kafka configuration for publishing messages
func (k *Configurator) CreatePublishConfig(topic string, req *pb.PublishRequest) KafkaConfig {
	cfg := NewDefaultKafkaConfig(k.brokers, topic)

	if req.ProducerConfig != nil {
		cfg.ExactlyOnceConfig.EnableIdempotence = req.ProducerConfig.EnableIdempotence
		cfg.ExactlyOnceConfig.EnableTransactions = req.ProducerConfig.EnableIdempotence
		if req.ProducerConfig.DeliverySemantics == pb.DeliverySemantics_DELIVERY_SEMANTICS_EXACTLY_ONCE {
			cfg.ExactlyOnceConfig.EnableTransactions = true
		}
	}

	return cfg
}

// CreateSubscribeConfig creates Kafka configuration for subscribing to messages
func (k *Configurator) CreateSubscribeConfig(topic string, groupID string, req *pb.SubscribeRequest) KafkaConfig {
	cfg := NewDefaultKafkaConfig(k.brokers, topic)
	cfg.GroupID = groupID

	if req.ConsumerConfig != nil {
		cfg.ConsumerConfig.MaxWait = time.Duration(req.ConsumerConfig.MaxWaitMs) * time.Millisecond
		cfg.ConsumerConfig.CommitInterval = time.Duration(req.ConsumerConfig.CommitIntervalMs) * time.Millisecond

		// Set isolation level based on enum
		switch req.ConsumerConfig.IsolationLevel {
		case pb.IsolationLevel_ISOLATION_LEVEL_READ_COMMITTED:
			cfg.ConsumerConfig.IsolationLevel = "read_committed"
		case pb.IsolationLevel_ISOLATION_LEVEL_READ_UNCOMMITTED:
			cfg.ConsumerConfig.IsolationLevel = "read_uncommitted"
		default:
			cfg.ConsumerConfig.IsolationLevel = "read_committed"
		}

		// Set auto offset reset based on enum
		switch req.ConsumerConfig.AutoOffsetReset {
		case pb.AutoOffsetReset_AUTO_OFFSET_RESET_LATEST:
			cfg.ConsumerConfig.AutoOffsetReset = "latest"
		case pb.AutoOffsetReset_AUTO_OFFSET_RESET_EARLIEST:
			cfg.ConsumerConfig.AutoOffsetReset = "earliest"
		default:
			cfg.ConsumerConfig.AutoOffsetReset = "latest"
		}
	}

	return cfg
}

// CreateTopicConfig creates Kafka configuration for topic creation
func (k *Configurator) CreateTopicConfig(topic string, req *pb.CreateTopicRequest) KafkaConfig {
	cfg := NewDefaultKafkaConfig(k.brokers, topic)
	cfg.NumPartitions = int(req.NumPartitions)
	cfg.ReplicationFactor = int(req.ReplicationFactor)
	return cfg
}
