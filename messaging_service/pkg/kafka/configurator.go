package kafka

import (
	pb "cra-protos/messaging_service"
	"messaging_service/internal/config"
	"time"
)

// Configurator handles all Kafka-specific configuration logic
type Configurator struct {
	brokers []string
	config  *config.Config
}

// NewConfigurator creates a new Kafka Configurator
func NewConfigurator(brokers []string, config *config.Config) *Configurator {
	return &Configurator{
		brokers: brokers,
		config:  config,
	}
}

// CreatePublishConfig creates Kafka configuration for publishing messages
func (k *Configurator) CreatePublishConfig(topic string, req *pb.PublishRequest) KafkaConfig {
	cfg := NewDefaultKafkaConfig(k.brokers, topic)

	// Set exactly-once delivery semantics by default for publishing
	cfg.DeliverySemantics = ExactlyOnce
	cfg.ExactlyOnceConfig.EnableIdempotence = true
	cfg.ExactlyOnceConfig.EnableTransactions = true

	// Apply environment-based configuration if available
	if k.config != nil {
		cfg.MaxAttempts = k.config.KafkaMaxAttempts
		cfg.RetryBackoffMs = k.config.KafkaRetryBackoffMs
		cfg.BatchSize = k.config.KafkaBatchSize
		cfg.BatchBytes = k.config.KafkaBatchBytes
		cfg.BatchTimeout = time.Duration(k.config.KafkaBatchTimeoutMs) * time.Millisecond
		cfg.ReadTimeout = time.Duration(k.config.KafkaReadTimeoutMs) * time.Millisecond
		cfg.WriteTimeout = time.Duration(k.config.KafkaWriteTimeoutMs) * time.Millisecond
		cfg.CompressionCodec = k.config.KafkaCompressionCodec
	} else {
		// Fallback to default values if config is not available
		cfg.MaxAttempts = 3
		cfg.RetryBackoffMs = 100
		cfg.BatchSize = 100
		cfg.BatchBytes = 1 * 1024 * 1024 // 1MB
		cfg.BatchTimeout = 500 * time.Millisecond
		cfg.ReadTimeout = 5 * time.Second
		cfg.WriteTimeout = 5 * time.Second
		cfg.CompressionCodec = "snappy"
	}

	return cfg
}

// CreateSubscribeConfig creates Kafka configuration for subscribing to messages
func (k *Configurator) CreateSubscribeConfig(topic string, groupID string, req *pb.SubscribeRequest) KafkaConfig {
	cfg := NewDefaultKafkaConfig(k.brokers, topic)
	cfg.GroupID = groupID

	// Apply environment-based configuration if available
	if k.config != nil {
		// Use environment values for consumer settings
		cfg.BatchSize = k.config.KafkaBatchSize
		cfg.BatchBytes = k.config.KafkaBatchBytes
		cfg.BatchTimeout = time.Duration(k.config.KafkaBatchTimeoutMs) * time.Millisecond
		cfg.CompressionCodec = k.config.KafkaCompressionCodec
		cfg.MaxAttempts = k.config.KafkaMaxAttempts
		cfg.RetryBackoffMs = k.config.KafkaRetryBackoffMs
		cfg.ReadTimeout = time.Duration(k.config.KafkaReadTimeoutMs) * time.Millisecond
		cfg.WriteTimeout = time.Duration(k.config.KafkaWriteTimeoutMs) * time.Millisecond

		// Set consumer specific settings from environment
		cfg.ConsumerConfig.MaxWait = time.Duration(k.config.KafkaConsumerMaxWaitMs) * time.Millisecond
		cfg.ConsumerConfig.CommitInterval = time.Duration(k.config.KafkaConsumerCommitIntervalMs) * time.Millisecond
		cfg.ConsumerConfig.SessionTimeout = time.Duration(k.config.KafkaConsumerSessionTimeoutMs) * time.Millisecond
		cfg.ConsumerConfig.HeartbeatInterval = time.Duration(k.config.KafkaConsumerHeartbeatMs) * time.Millisecond

		// Set auto offset reset from environment
		cfg.ConsumerConfig.AutoOffsetReset = k.config.KafkaConsumerAutoOffsetReset

		// Set isolation level for exactly-once delivery
		cfg.ConsumerConfig.IsolationLevel = k.config.KafkaIsolationLevel

		// Configure for exactly-once delivery semantics to prevent duplicates
		if k.config.KafkaIsolationLevel == "read_committed" {
			cfg.DeliverySemantics = ExactlyOnce
			cfg.ExactlyOnceConfig.EnableDeduplication = true
			cfg.ExactlyOnceConfig.IsolationLevel = "read_committed"
		}
	} else {
		// Default consumer settings if environment is not available
		cfg.ConsumerConfig.MaxWait = 100 * time.Millisecond
		cfg.ConsumerConfig.CommitInterval = 1000 * time.Millisecond
	}

	// Default to read_committed isolation level if not set in environment
	if cfg.ConsumerConfig.IsolationLevel == "" {
		cfg.ConsumerConfig.IsolationLevel = mapIsolationLevel(pb.IsolationLevel_ISOLATION_LEVEL_READ_COMMITTED)
	}

	// Default to latest offset reset if not set in environment
	if cfg.ConsumerConfig.AutoOffsetReset == "" {
		cfg.ConsumerConfig.AutoOffsetReset = mapAutoOffsetReset(pb.AutoOffsetReset_AUTO_OFFSET_RESET_LATEST)
	}

	// If provided in the request, use the specified isolation level and auto offset reset
	if req != nil {
		// Other configuration options would be set here based on the request
	}

	return cfg
}

// CreateTopicConfig creates Kafka configuration for topic creation
func (k *Configurator) CreateTopicConfig(topic string, req *pb.CreateTopicRequest) KafkaConfig {
	cfg := NewDefaultKafkaConfig(k.brokers, topic)

	// Apply environment-based configuration if available
	if k.config != nil {
		cfg.NumPartitions = k.config.KafkaNumPartitions
		cfg.ReplicationFactor = k.config.KafkaReplicationFactor
		cfg.BatchSize = k.config.KafkaBatchSize
		cfg.BatchBytes = k.config.KafkaBatchBytes
		cfg.BatchTimeout = time.Duration(k.config.KafkaBatchTimeoutMs) * time.Millisecond
		cfg.CompressionCodec = k.config.KafkaCompressionCodec
		cfg.ConsumerConfig.RetentionTime = time.Duration(k.config.KafkaRetentionMs) * time.Millisecond
	} else {
		// Default configuration as fallback
		cfg.NumPartitions = 3
		cfg.ReplicationFactor = 3
		cfg.BatchSize = 100
		cfg.BatchBytes = 1 * 1024 * 1024 // 1MB
		cfg.BatchTimeout = 500 * time.Millisecond
		cfg.CompressionCodec = "snappy"
		cfg.ConsumerConfig.RetentionTime = 3 * time.Second
	}

	// Set cleanup policy based on protobuf enum
	// Just mapping here, actual usage would be in admin.CreateTopic
	_ = mapCleanupPolicy(pb.CleanupPolicy_CLEANUP_POLICY_DELETE)

	return cfg
}

// CreateProducer creates a new Producer for a topic with appropriate configuration
func (k *Configurator) CreateProducer(topic string) Producer {
	cfg := k.CreatePublishConfig(topic, nil)
	return NewProducer(cfg)
}

// Helper functions to map protobuf enums to Kafka config values

// mapDeliverySemantics maps the protobuf DeliverySemantics enum to Kafka config values
func mapDeliverySemantics(semantics pb.DeliverySemantics) DeliverySemantics {
	switch semantics {
	case pb.DeliverySemantics_DELIVERY_SEMANTICS_AT_LEAST_ONCE:
		return AtLeastOnce
	case pb.DeliverySemantics_DELIVERY_SEMANTICS_EXACTLY_ONCE:
		return ExactlyOnce
	default:
		return AtLeastOnce // Default to at-least-once for safety
	}
}

// mapIsolationLevel maps the protobuf IsolationLevel enum to Kafka config values
func mapIsolationLevel(level pb.IsolationLevel) string {
	switch level {
	case pb.IsolationLevel_ISOLATION_LEVEL_READ_COMMITTED:
		return "read_committed"
	case pb.IsolationLevel_ISOLATION_LEVEL_READ_UNCOMMITTED:
		return "read_uncommitted"
	default:
		return "read_committed" // Default to read_committed for safety
	}
}

// mapAutoOffsetReset maps the protobuf AutoOffsetReset enum to Kafka config values
func mapAutoOffsetReset(reset pb.AutoOffsetReset) string {
	switch reset {
	case pb.AutoOffsetReset_AUTO_OFFSET_RESET_LATEST:
		return "latest"
	case pb.AutoOffsetReset_AUTO_OFFSET_RESET_EARLIEST:
		return "earliest"
	default:
		return "latest" // Default to latest for most common use case
	}
}

// mapCleanupPolicy maps the protobuf CleanupPolicy enum to Kafka config values
func mapCleanupPolicy(policy pb.CleanupPolicy) string {
	switch policy {
	case pb.CleanupPolicy_CLEANUP_POLICY_DELETE:
		return "delete"
	case pb.CleanupPolicy_CLEANUP_POLICY_COMPACT:
		return "compact"
	default:
		return "delete" // Default to delete for most common use case
	}
}
