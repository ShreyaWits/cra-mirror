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

// KafkaConfig holds all configuration related to Kafka connections
type KafkaConfig struct {
	// Core connection settings
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

	// Retry and timeout settings
	MaxAttempts    int           // Maximum number of retry attempts
	RetryBackoffMs int           // Base backoff time between retries in milliseconds
	ReadTimeout    time.Duration // Read timeout for Kafka operations
	WriteTimeout   time.Duration // Write timeout for Kafka operations

	// Producer batch settings
	BatchSize    int           // Number of messages to batch together
	BatchBytes   int64         // Maximum size of a batch in bytes
	BatchTimeout time.Duration // Maximum time to wait for a batch to fill

	// Message format and compression
	CompressionCodec string // Compression codec: gzip, snappy, lz4, zstd

	// Additional config objects
	ExactlyOnceConfig ExactlyOnceConfig
	ConsumerConfig    ConsumerConfig
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
		MaxAttempts:       3,                      // Default retry attempts
		RetryBackoffMs:    100,                    // Default backoff in ms
		ReadTimeout:       10 * time.Second,       // Default read timeout
		WriteTimeout:      10 * time.Second,       // Default write timeout
		BatchSize:         100,                    // Default batch size
		BatchBytes:        int64(1 * 1024 * 1024), // 1MB default batch size
		BatchTimeout:      1 * time.Second,        // Default batch timeout
		CompressionCodec:  "snappy",               // Default compression
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

// WithMaxAttempts sets the maximum number of retry attempts
func (c KafkaConfig) WithMaxAttempts(attempts int) KafkaConfig {
	c.MaxAttempts = attempts
	return c
}

// WithIsolationLevel sets the isolation level for reading
func (c KafkaConfig) WithIsolationLevel(level string) KafkaConfig {
	c.ConsumerConfig.IsolationLevel = level
	return c
}

// WithAutoOffsetReset sets the auto offset reset behavior
func (c KafkaConfig) WithAutoOffsetReset(reset string) KafkaConfig {
	c.ConsumerConfig.AutoOffsetReset = reset
	return c
}

// WithRetryBackoff sets the retry backoff in milliseconds
func (c KafkaConfig) WithRetryBackoff(backoffMs int) KafkaConfig {
	c.RetryBackoffMs = backoffMs
	return c
}

// WithReadTimeout sets the read timeout for Kafka operations
func (c KafkaConfig) WithReadTimeout(timeout time.Duration) KafkaConfig {
	c.ReadTimeout = timeout
	return c
}

// WithWriteTimeout sets the write timeout for Kafka operations
func (c KafkaConfig) WithWriteTimeout(timeout time.Duration) KafkaConfig {
	c.WriteTimeout = timeout
	return c
}

// WithBatchSettings sets the batch processing configuration
func (c KafkaConfig) WithBatchSettings(size int, bytes int, timeout time.Duration) KafkaConfig {
	c.BatchSize = size
	c.BatchBytes = int64(bytes)
	c.BatchTimeout = timeout
	return c
}

// WithCompression sets the compression codec
func (c KafkaConfig) WithCompression(codec string) KafkaConfig {
	c.CompressionCodec = codec
	return c
}
