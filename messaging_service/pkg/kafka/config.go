package kafka

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

// KafkaClusterMode defines the coordination mode for Kafka
type KafkaClusterMode string

const (
	// ZookeeperMode uses Zookeeper for broker coordination
	ZookeeperMode KafkaClusterMode = "zookeeper"
)

type ConsumerMode string

const (
	// Default replication factor for high availability
	DefaultReplicationFactor = 3

	// Default number of partitions for scalability
	DefaultPartitions = 3

	// Consumer group behaviors
	LatestOffset   ConsumerMode = "latest"   // Only read new messages after joining
	EarliestOffset ConsumerMode = "earliest" // Read all unprocessed messages
)

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

// DefaultExactlyOnceConfig returns a default configuration for exactly-once semantics
func DefaultExactlyOnceConfig() ExactlyOnceConfig {
	return ExactlyOnceConfig{
		EnableIdempotence:     true,
		EnableTransactions:    true,
		TransactionTimeoutMs:  60000, // 60 seconds
		TransactionalIDPrefix: "txn",
		IsolationLevel:        "read_committed",
		EnableDeduplication:   true,
		DeduplicationWindowMs: 30 * 60 * 1000, // 30 minutes
	}
}

// ConsumerConfig holds configuration for consumer behavior
type ConsumerConfig struct {
	// MaxWait is the maximum amount of time to wait for a batch of messages
	MaxWait time.Duration

	// ReadBackoffMin is the minimum amount of time to wait before retrying a read
	ReadBackoffMin time.Duration

	// ReadBackoffMax is the maximum amount of time to wait before retrying a read
	ReadBackoffMax time.Duration

	// CommitInterval is the interval at which offsets are committed to the broker
	CommitInterval time.Duration

	// HeartbeatInterval is the interval at which heartbeats are sent to the broker
	HeartbeatInterval time.Duration

	// SessionTimeout is the timeout used to detect consumer failures
	SessionTimeout time.Duration

	// RebalanceTimeout is the maximum time allowed for the group to rebalance
	RebalanceTimeout time.Duration

	// RetentionTime is the time to retain messages in the topic
	RetentionTime time.Duration

	// MaxAttempts is the maximum number of attempts to read a message
	MaxAttempts int

	// IsolationLevel determines whether to read committed, uncommitted, or both
	IsolationLevel string

	// AutoOffsetReset determines where to start reading when no offset is stored
	AutoOffsetReset string
}

// DefaultConsumerConfig returns a default configuration for consumer behavior
func DefaultConsumerConfig() ConsumerConfig {
	return ConsumerConfig{
		MaxWait:           500 * time.Millisecond,
		ReadBackoffMin:    50 * time.Millisecond,
		ReadBackoffMax:    200 * time.Millisecond,
		CommitInterval:    1 * time.Second,
		HeartbeatInterval: 3 * time.Second,
		SessionTimeout:    10 * time.Second,
		RebalanceTimeout:  60 * time.Second,
		RetentionTime:     3 * time.Second, // Changed to 3 seconds to match retention.ms
		MaxAttempts:       5,
		IsolationLevel:    "read_committed",
		AutoOffsetReset:   "latest",
	}
}

type KafkaConfig struct {
	Brokers           []string
	Topic             string
	GroupID           string
	Mode              ConsumerMode
	BalancerType      string
	MinBytes          int
	MaxBytes          int
	DeliverySemantics DeliverySemantics
	ReplicationFactor int
	NumPartitions     int
	ClusterMode       KafkaClusterMode
	ExactlyOnceConfig ExactlyOnceConfig
	ConsumerConfig    ConsumerConfig // New field for consumer configuration
}

// NewDefaultKafkaConfig creates a new KafkaConfig with sensible defaults for production use
func NewDefaultKafkaConfig(brokers []string, topic string) KafkaConfig {
	return KafkaConfig{
		Brokers:           brokers,
		Topic:             topic,
		Mode:              LatestOffset, // Default to latest offset - only receive new messages after consumer starts
		BalancerType:      "round_robin",
		MinBytes:          10 * 1024,        // 10KB
		MaxBytes:          10 * 1024 * 1024, // 10MB
		DeliverySemantics: AtLeastOnce,      // Default to at-least-once for safety
		ReplicationFactor: DefaultReplicationFactor,
		NumPartitions:     DefaultPartitions,
		ClusterMode:       ZookeeperMode, // Default to Zookeeper
		ExactlyOnceConfig: DefaultExactlyOnceConfig(),
		ConsumerConfig:    DefaultConsumerConfig(),
	}
}

// WithAtLeastOnce configures the Kafka client for at-least-once delivery semantics
func (c KafkaConfig) WithAtLeastOnce() KafkaConfig {
	c.DeliverySemantics = AtLeastOnce
	return c
}

// WithExactlyOnce configures the Kafka client for exactly-once delivery semantics
func (c KafkaConfig) WithExactlyOnce() KafkaConfig {
	c.DeliverySemantics = ExactlyOnce
	c.ExactlyOnceConfig = DefaultExactlyOnceConfig()
	return c
}

// WithCustomExactlyOnce configures the Kafka client with custom exactly-once settings
func (c KafkaConfig) WithCustomExactlyOnce(config ExactlyOnceConfig) KafkaConfig {
	c.DeliverySemantics = ExactlyOnce
	c.ExactlyOnceConfig = config
	return c
}

// WithHighAvailability configures the Kafka topic for high availability
func (c KafkaConfig) WithHighAvailability(replicationFactor int) KafkaConfig {
	c.ReplicationFactor = replicationFactor
	return c
}

// WithLatestOffset configures the consumer to only read new messages after joining
func (c KafkaConfig) WithLatestOffset() KafkaConfig {
	c.Mode = LatestOffset
	return c
}

// WithEarliestOffset configures the consumer to read all unprocessed messages
func (c KafkaConfig) WithEarliestOffset() KafkaConfig {
	c.Mode = EarliestOffset
	return c
}

// WithConsumerConfig sets custom consumer configuration
func (c KafkaConfig) WithConsumerConfig(config ConsumerConfig) KafkaConfig {
	c.ConsumerConfig = config
	return c
}

// WithMaxWait sets the maximum wait time for a batch of messages
func (c KafkaConfig) WithMaxWait(duration time.Duration) KafkaConfig {
	c.ConsumerConfig.MaxWait = duration
	return c
}

// WithReadBackoff sets the read backoff configuration
func (c KafkaConfig) WithReadBackoff(min, max time.Duration) KafkaConfig {
	c.ConsumerConfig.ReadBackoffMin = min
	c.ConsumerConfig.ReadBackoffMax = max
	return c
}

// WithCommitInterval sets the commit interval
func (c KafkaConfig) WithCommitInterval(duration time.Duration) KafkaConfig {
	c.ConsumerConfig.CommitInterval = duration
	return c
}

// WithHeartbeatInterval sets the heartbeat interval
func (c KafkaConfig) WithHeartbeatInterval(duration time.Duration) KafkaConfig {
	c.ConsumerConfig.HeartbeatInterval = duration
	return c
}

// WithSessionTimeout sets the session timeout
func (c KafkaConfig) WithSessionTimeout(duration time.Duration) KafkaConfig {
	c.ConsumerConfig.SessionTimeout = duration
	return c
}

// WithRebalanceTimeout sets the rebalance timeout
func (c KafkaConfig) WithRebalanceTimeout(duration time.Duration) KafkaConfig {
	c.ConsumerConfig.RebalanceTimeout = duration
	return c
}

// WithRetentionTime sets the message retention time
func (c KafkaConfig) WithRetentionTime(duration time.Duration) KafkaConfig {
	c.ConsumerConfig.RetentionTime = duration
	return c
}

// WithMaxAttempts sets the maximum number of read attempts
func (c KafkaConfig) WithMaxAttempts(attempts int) KafkaConfig {
	c.ConsumerConfig.MaxAttempts = attempts
	return c
}

// WithIsolationLevel sets the isolation level
func (c KafkaConfig) WithIsolationLevel(level string) KafkaConfig {
	c.ConsumerConfig.IsolationLevel = level
	return c
}

// WithAutoOffsetReset sets the auto offset reset behavior
func (c KafkaConfig) WithAutoOffsetReset(reset string) KafkaConfig {
	c.ConsumerConfig.AutoOffsetReset = reset
	return c
}
