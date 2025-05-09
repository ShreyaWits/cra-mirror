package kafka

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
	// ZookeeperMode uses Zookeeper for broker coordination (legacy)
	ZookeeperMode KafkaClusterMode = "zookeeper"

	// KRaftMode uses Kafka Raft (KRaft) for broker coordination (modern, no Zookeeper)
	KRaftMode KafkaClusterMode = "kraft"
)

type ConsumerMode string

const (
	// Default replication factor for high availability
	DefaultReplicationFactor = 3

	// Default number of partitions for scalability
	DefaultPartitions = 3
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
}

// NewDefaultKafkaConfig creates a new KafkaConfig with sensible defaults for production use
func NewDefaultKafkaConfig(brokers []string, topic string) KafkaConfig {
	return KafkaConfig{
		Brokers:           brokers,
		Topic:             topic,
		Mode:              "", // Will be set based on usage
		BalancerType:      "round_robin",
		MinBytes:          10 * 1024,        // 10KB
		MaxBytes:          10 * 1024 * 1024, // 10MB
		DeliverySemantics: AtLeastOnce,      // Default to at-least-once for safety
		ReplicationFactor: DefaultReplicationFactor,
		NumPartitions:     DefaultPartitions,
		ClusterMode:       ZookeeperMode, // Default to Zookeeper for backward compatibility
		ExactlyOnceConfig: DefaultExactlyOnceConfig(),
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

// WithKRaftMode configures the Kafka client to use KRaft mode (no Zookeeper)
func (c KafkaConfig) WithKRaftMode() KafkaConfig {
	c.ClusterMode = KRaftMode
	return c
}
