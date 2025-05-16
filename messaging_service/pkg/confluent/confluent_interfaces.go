package confluent

import (
	"context"
	"messaging_service/internal/config"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

type KafkaProducerInterface interface {
	Events() chan kafka.Event
	Produce(msg *kafka.Message, deliveryChan chan kafka.Event) error
	Flush(timeoutMs int) int
	InitTransactions(ctx context.Context) error
	BeginTransaction() error
	CommitTransaction(ctx context.Context) error
	AbortTransaction(ctx context.Context) error
	Close()
}

type Producer interface {
	// WriteWithRetry sends a message to Kafka with retry logic
	WriteWithRetry(ctx context.Context, key, value []byte, headers []Header) error

	// BeginTransaction starts a new transaction for exactly-once semantics
	BeginTransaction() error

	// CommitTransaction commits the current transaction
	CommitTransaction(ctx context.Context) error

	// AbortTransaction aborts the current transaction
	AbortTransaction(ctx context.Context) error

	// Close closes the producer and frees resources
	Close() error
}
type KafkaAdmin interface {
	// CreateTopic creates a new topic with the given configuration
	CreateTopic(ctx context.Context, topic string, numPartitions, replicationFactor int, configs map[string]string) error

	// DeleteTopic deletes an existing topic
	DeleteTopic(ctx context.Context, topic string) error

	// ListTopics lists all available topics
	ListTopics(ctx context.Context) ([]string, error)

	// HasActiveConsumers checks if a topic has any active consumer groups
	HasActiveConsumers(ctx context.Context, topic string) (bool, error)

	// Close releases any resources held by the admin client
	Close() error
}

// KafkaFactory defines the factory interface for creating Confluent Kafka components
type KafkaFactory interface {
	// CreateProducer creates a new producer
	CreateProducer(cfg KafkaConfig) (Producer, error)

	// CreateConsumer creates a new consumer
	CreateConsumer(cfg KafkaConfig, groupID string, handler func([]byte) error) (Consumer, error)

	// CreateAdmin creates a new admin client
	CreateAdmin(*config.Config) (KafkaAdmin, error)

	// CreateDLQProducer creates a new DLQ producer
	CreateDLQProducer(cfg KafkaConfig) (DLQProducer, error)
}
