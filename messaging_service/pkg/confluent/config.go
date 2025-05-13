package confluent

import (
	"time"
)

// DeliverySemantics defines the message delivery guarantee
type DeliverySemantics string

const (
	// AtLeastOnce guarantees that messages are never lost but may be delivered more than once
	// This is the default with DLQ and retry policies for resilience
	AtLeastOnce DeliverySemantics = "at-least-once"

	// ExactlyOnce guarantees that messages are delivered exactly once
	// Should only be used where duplicate side effects are catastrophic (e.g., payments, audit trails)
	ExactlyOnce DeliverySemantics = "exactly-once"
)

// KafkaConfig holds the configuration for Kafka producers and consumers
type KafkaConfig struct {
	// Connection settings
	Brokers []string
	Topic   string

	// Topic configuration
	NumPartitions     int
	ReplicationFactor int

	// Producer configuration
	BatchSize       int
	BatchBytes      int64
	BatchTimeout    time.Duration
	CompressionType string
	MaxAttempts     int
	RetryBackoffMs  int
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	BalancerType    string

	// Delivery semantics
	DeliverySemantics DeliverySemantics

	// Exactly-once semantics configuration
	ExactlyOnceConfig ExactlyOnceConfig

	// Consumer configuration
	ConsumerConfig ConsumerConfig
}

// ExactlyOnceConfig holds configuration for exactly-once semantics
type ExactlyOnceConfig struct {
	// Enable producer idempotence (enable.idempotence=true)
	EnableIdempotence bool

	// Enable transactions for exactly-once processing
	EnableTransactions bool

	// Transaction timeout in milliseconds
	TransactionTimeoutMs int

	// Transactional ID prefix (will be combined with a unique identifier)
	TransactionalIDPrefix string

	// Consumer isolation level (read_committed or read_uncommitted)
	IsolationLevel string

	// Enable consumer deduplication for exactly-once semantics
	EnableDeduplication bool

	// Deduplication window in milliseconds
	DeduplicationWindowMs int
}

// ConsumerConfig holds configuration for Kafka consumers
type ConsumerConfig struct {
	// Group ID for consumer groups
	GroupID string

	// Minimum number of bytes to fetch in a request
	MinBytes int

	// Maximum number of bytes to fetch in a request
	MaxBytes int

	// Maximum time to wait for a response from the broker
	MaxWait time.Duration

	// Minimum delay between reads
	ReadBackoffMin time.Duration

	// Maximum delay between reads
	ReadBackoffMax time.Duration

	// Auto offset reset policy (earliest, latest, or none)
	AutoOffsetReset string

	// Retention time for messages
	RetentionTime time.Duration

	// Maximum number of retry attempts
	MaxAttempts int

	// Whether to automatically commit offsets
	AutoCommit bool

	// Interval between auto commits
	CommitInterval time.Duration

	// Consumer session timeout
	SessionTimeout time.Duration

	// Consumer heartbeat interval
	HeartbeatInterval time.Duration

	// Maximum number of records to poll
	MaxPollRecords int
}

// NewDefaultKafkaConfig creates a new KafkaConfig with sensible defaults
func NewDefaultKafkaConfig(brokers []string, topic string) KafkaConfig {
	return KafkaConfig{
		Brokers:           brokers,
		Topic:             topic,
		NumPartitions:     3,
		ReplicationFactor: 3,
		BatchSize:         100,
		BatchBytes:        1024 * 1024, // 1MB
		BatchTimeout:      500 * time.Millisecond,
		CompressionType:   "snappy",
		MaxAttempts:       3,
		RetryBackoffMs:    100,
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      5 * time.Second,
		BalancerType:      "roundrobin",
		DeliverySemantics: AtLeastOnce,
		ExactlyOnceConfig: DefaultExactlyOnceConfig(),
		ConsumerConfig:    DefaultConsumerConfig(),
	}
}

// DefaultExactlyOnceConfig returns a default configuration for exactly-once semantics
func DefaultExactlyOnceConfig() ExactlyOnceConfig {
	return ExactlyOnceConfig{
		EnableIdempotence:     true,
		EnableTransactions:    true,
		TransactionTimeoutMs:  10000, // 10 seconds
		TransactionalIDPrefix: "txn",
		IsolationLevel:        "read_committed",
		EnableDeduplication:   false,
		DeduplicationWindowMs: 60000, // 1 minute
	}
}

// DefaultConsumerConfig returns a default configuration for consumers
func DefaultConsumerConfig() ConsumerConfig {
	return ConsumerConfig{
		GroupID:           "default-group",
		MinBytes:          10e3, // 10KB
		MaxBytes:          10e6, // 10MB
		MaxWait:           1 * time.Second,
		ReadBackoffMin:    100 * time.Millisecond,
		ReadBackoffMax:    1 * time.Second,
		AutoOffsetReset:   "earliest",
		RetentionTime:     7 * 24 * time.Hour, // 7 days
		MaxAttempts:       3,
		AutoCommit:        false,
		CommitInterval:    5 * time.Second,
		SessionTimeout:    30 * time.Second,
		HeartbeatInterval: 1 * time.Second,
		MaxPollRecords:    500,
	}
}

// ToConfluent converts the KafkaConfig to a Confluent ConfigMap
func (cfg *KafkaConfig) ToConfluentProducerConfig() *ConfigMap {
	configMap := &ConfigMap{
		"bootstrap.servers": buildBrokerString(cfg.Brokers),
		"retry.backoff.ms":  cfg.RetryBackoffMs,
		"message.max.bytes": int(cfg.BatchBytes),
	}

	// Set timeout values
	if cfg.ReadTimeout > 0 {
		_ = configMap.SetKey("socket.timeout.ms", int(cfg.ReadTimeout.Milliseconds()))
	}
	if cfg.WriteTimeout > 0 {
		_ = configMap.SetKey("message.timeout.ms", int(cfg.WriteTimeout.Milliseconds()))
	}

	// Set compression
	if cfg.CompressionType != "" {
		_ = configMap.SetKey("compression.type", cfg.CompressionType)
	}

	// Configure batch settings
	if cfg.BatchSize > 0 {
		_ = configMap.SetKey("batch.size", cfg.BatchSize)
	}
	if cfg.BatchTimeout > 0 {
		_ = configMap.SetKey("linger.ms", int(cfg.BatchTimeout.Milliseconds()))
	}

	// Configure delivery semantics
	if cfg.DeliverySemantics == ExactlyOnce {
		// Enable idempotence for exactly-once delivery
		if cfg.ExactlyOnceConfig.EnableIdempotence {
			_ = configMap.SetKey("enable.idempotence", true)
		}

		// Always require all acks for exactly-once semantics
		_ = configMap.SetKey("acks", "all")

		// Enable transactions if needed
		if cfg.ExactlyOnceConfig.EnableTransactions {
			if cfg.ExactlyOnceConfig.TransactionalIDPrefix != "" {
				_ = configMap.SetKey("transactional.id", cfg.ExactlyOnceConfig.TransactionalIDPrefix+"-"+cfg.Topic)
			}
			if cfg.ExactlyOnceConfig.TransactionTimeoutMs > 0 {
				_ = configMap.SetKey("transaction.timeout.ms", cfg.ExactlyOnceConfig.TransactionTimeoutMs)
			}
		}
	} else {
		// Default to at-least-once semantics
		_ = configMap.SetKey("acks", "1")
	}

	return configMap
}

// ToConfluentConsumerConfig converts the KafkaConfig to a Confluent consumer ConfigMap
func (cfg *KafkaConfig) ToConfluentConsumerConfig(groupID string) *ConfigMap {
	configMap := &ConfigMap{
		"bootstrap.servers":  buildBrokerString(cfg.Brokers),
		"group.id":           groupID,
		"auto.offset.reset":  cfg.ConsumerConfig.AutoOffsetReset,
		"enable.auto.commit": cfg.ConsumerConfig.AutoCommit,
	}

	// Set read timeouts
	if cfg.ReadTimeout > 0 {
		_ = configMap.SetKey("socket.timeout.ms", int(cfg.ReadTimeout.Milliseconds()))
	}

	// Set consumer group settings
	if cfg.ConsumerConfig.SessionTimeout > 0 {
		_ = configMap.SetKey("session.timeout.ms", int(cfg.ConsumerConfig.SessionTimeout.Milliseconds()))
	}
	if cfg.ConsumerConfig.HeartbeatInterval > 0 {
		_ = configMap.SetKey("heartbeat.interval.ms", int(cfg.ConsumerConfig.HeartbeatInterval.Milliseconds()))
	}

	// Set message consumption settings
	if cfg.ConsumerConfig.CommitInterval > 0 {
		_ = configMap.SetKey("auto.commit.interval.ms", int(cfg.ConsumerConfig.CommitInterval.Milliseconds()))
	}

	// Configure isolation level for exactly-once semantics
	if cfg.DeliverySemantics == ExactlyOnce && cfg.ExactlyOnceConfig.IsolationLevel != "" {
		_ = configMap.SetKey("isolation.level", cfg.ExactlyOnceConfig.IsolationLevel)
	}

	return configMap
}

// Helper function to build broker string
func buildBrokerString(brokers []string) string {
	if len(brokers) == 0 {
		return "localhost:9092"
	}

	result := ""
	for i, broker := range brokers {
		if i > 0 {
			result += ","
		}
		result += broker
	}
	return result
}
