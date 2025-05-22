package confluent

import (
	pb "cra-protos/messaging_service"
	"fmt"
	"messaging_service/internal/config"
	"time"
)

// Configurator handles all Kafka-specific configuration logic for Confluent
type Configurator struct {
	brokers []string
	config  *config.Config
}

// NewConfigurator creates a new Confluent Kafka Configurator
func NewConfigurator(brokers []string, config *config.Config) *Configurator {
	// Note: We'll keep this for backward compatibility but update callers to use SafeNewConfigurator
	// Validate input arguments
	if brokers == nil || len(brokers) == 0 {
		panic("Kafka brokers cannot be nil or empty in configurator")
	}
	if config == nil {
		panic("config cannot be nil in configurator")
	}

	return &Configurator{
		brokers: brokers,
		config:  config,
	}
}

// SafeNewConfigurator creates a new Confluent Kafka Configurator with error return instead of panic
func SafeNewConfigurator(brokers []string, config *config.Config) (*Configurator, error) {
	// Validate input arguments
	if brokers == nil || len(brokers) == 0 {
		return nil, fmt.Errorf("kafka brokers cannot be nil or empty in configurator")
	}
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil in configurator")
	}

	return &Configurator{
		brokers: brokers,
		config:  config,
	}, nil
}

// CreatePublishConfig creates Confluent Kafka configuration for publishing messages
func (k *Configurator) CreatePublishConfig(topic string, req *pb.PublishRequest) KafkaConfig {
	cfg := NewDefaultKafkaConfig(k.brokers, topic)

	// Set exactly-once delivery semantics by default for publishing
	cfg.DeliverySemantics = ExactlyOnce
	cfg.ExactlyOnceConfig.EnableIdempotence = true
	cfg.ExactlyOnceConfig.EnableTransactions = true

	cfg.MaxAttempts = k.config.KafkaMaxAttempts
	cfg.RetryBackoffMs = k.config.KafkaRetryBackoffMs
	cfg.BatchSize = k.config.KafkaBatchSize
	cfg.BatchBytes = k.config.KafkaBatchBytes
	cfg.BatchTimeout = time.Duration(k.config.KafkaBatchTimeoutMs) * time.Millisecond
	cfg.ReadTimeout = time.Duration(k.config.KafkaReadTimeoutMs) * time.Millisecond
	cfg.WriteTimeout = time.Duration(k.config.KafkaWriteTimeoutMs) * time.Millisecond
	cfg.CompressionType = k.config.KafkaCompressionCodec

	return cfg
}

// CreateSubscribeConfig creates Confluent Kafka configuration for subscribing to messages
func (k *Configurator) CreateSubscribeConfig(topic string, groupID string, req *pb.SubscribeRequest) KafkaConfig {
	cfg := NewDefaultKafkaConfig(k.brokers, topic)
	cfg.ConsumerConfig.GroupID = groupID

	// Set exactly-once delivery semantics by default for consumers
	cfg.DeliverySemantics = ExactlyOnce
	cfg.ExactlyOnceConfig.EnableDeduplication = true
	cfg.ExactlyOnceConfig.IsolationLevel = "read_committed"

	// Apply environment-based configuration if available
	if k.config != nil {
		// Use environment values for consumer settings
		cfg.BatchSize = k.config.KafkaBatchSize
		cfg.BatchBytes = k.config.KafkaBatchBytes
		cfg.BatchTimeout = time.Duration(k.config.KafkaBatchTimeoutMs) * time.Millisecond
		cfg.CompressionType = k.config.KafkaCompressionCodec
		cfg.MaxAttempts = k.config.KafkaMaxAttempts
		cfg.RetryBackoffMs = k.config.KafkaRetryBackoffMs
		cfg.ReadTimeout = time.Duration(k.config.KafkaReadTimeoutMs) * time.Millisecond
		cfg.WriteTimeout = time.Duration(k.config.KafkaWriteTimeoutMs) * time.Millisecond

		// Set consumer specific settings from environment
		cfg.ConsumerConfig.MaxWait = time.Duration(k.config.KafkaConsumerMaxWaitMs) * time.Millisecond
		cfg.ConsumerConfig.CommitInterval = time.Duration(k.config.KafkaConsumerCommitIntervalMs) * time.Millisecond
		cfg.ConsumerConfig.SessionTimeout = time.Duration(k.config.KafkaConsumerSessionTimeoutMs) * time.Millisecond
		cfg.ConsumerConfig.HeartbeatInterval = time.Duration(k.config.KafkaConsumerHeartbeatMs) * time.Millisecond

		// Set max poll records if configured
		if k.config.KafkaConsumerMaxPollRecords > 0 {
			cfg.ConsumerConfig.MaxPollRecords = k.config.KafkaConsumerMaxPollRecords
		}

		// Set auto offset reset from environment
		cfg.ConsumerConfig.AutoOffsetReset = k.config.KafkaConsumerAutoOffsetReset

		// Set auto commit from environment
		cfg.ConsumerConfig.AutoCommit = k.config.KafkaEnableAutoCommit

		// Override isolation level if specified in the config
		if k.config.KafkaIsolationLevel != "" {
			cfg.ExactlyOnceConfig.IsolationLevel = k.config.KafkaIsolationLevel
		}
	} else {
		// Default consumer settings if environment is not available
		cfg.ConsumerConfig.MaxWait = 100 * time.Millisecond
		cfg.ConsumerConfig.CommitInterval = 1000 * time.Millisecond
		cfg.ConsumerConfig.MaxPollRecords = 500
		cfg.ConsumerConfig.AutoCommit = false
	}

	// Default to latest offset reset if not set in environment
	if cfg.ConsumerConfig.AutoOffsetReset == "" {
		cfg.ConsumerConfig.AutoOffsetReset = "latest"
	}

	return cfg
}

// CreateTopicConfig creates Confluent Kafka configuration for topic creation
func (k *Configurator) CreateTopicConfig(topic string, req *pb.CreateTopicRequest) KafkaConfig {
	cfg := NewDefaultKafkaConfig(k.brokers, topic)

	// Apply environment-based configuration if available
	if k.config != nil {
		cfg.NumPartitions = k.config.KafkaNumPartitions
		cfg.ReplicationFactor = k.config.KafkaReplicationFactor
		cfg.BatchSize = k.config.KafkaBatchSize
		cfg.BatchBytes = k.config.KafkaBatchBytes
		cfg.BatchTimeout = time.Duration(k.config.KafkaBatchTimeoutMs) * time.Millisecond
		cfg.CompressionType = k.config.KafkaCompressionCodec

		// Set reasonable minimums for partitions and replication
		if cfg.NumPartitions < 1 {
			cfg.NumPartitions = 1
		}

		if cfg.ReplicationFactor < 1 {
			cfg.ReplicationFactor = 1
		}
	} else {
		// Default configuration as fallback
		cfg.NumPartitions = 3
		cfg.ReplicationFactor = 3
		cfg.BatchSize = 100
		cfg.BatchBytes = 1 * 1024 * 1024 // 1MB
		cfg.BatchTimeout = 500 * time.Millisecond
		cfg.CompressionType = "snappy"
	}

	return cfg
}

// CreateAdminConfig creates Confluent Kafka configuration for admin operations
func (k *Configurator) CreateAdminConfig() KafkaConfig {
	cfg := NewDefaultKafkaConfig(k.brokers, "")

	// Apply environment-based configuration if available
	if k.config != nil {
		cfg.ReadTimeout = time.Duration(k.config.KafkaReadTimeoutMs) * time.Millisecond
		cfg.WriteTimeout = time.Duration(k.config.KafkaWriteTimeoutMs) * time.Millisecond
	} else {
		cfg.ReadTimeout = 30 * time.Second
		cfg.WriteTimeout = 30 * time.Second
	}

	return cfg
}
