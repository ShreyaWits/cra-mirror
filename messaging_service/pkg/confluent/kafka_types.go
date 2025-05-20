package confluent

import "github.com/confluentinc/confluent-kafka-go/v2/kafka"

// This file defines type aliases for commonly used kafka types to simplify imports
// and avoid undefined type errors when using the Confluent Kafka Go library.
//
// Usage:
// 1. For fields in structs: use the type aliases directly
//    Example: producer *KafkaProducer
//
// 2. For parameters/return values in functions that call the kafka library directly:
//    continue using the kafka package directly
//    Example: producer, err := kafka.NewProducer(...)
//
// 3. For function signatures in your own interfaces: use the type aliases
//    Example: func SendMessage(headers []Header) error

// Define type aliases for Kafka types used across the codebase
type (
	// Header is an alias for kafka.Header
	Header = kafka.Header

	// KafkaProducer is an alias for kafka.Producer
	KafkaProducer = kafka.Producer

	// KafkaConsumer is an alias for kafka.Consumer
	KafkaConsumer = kafka.Consumer

	// ConfigMap is an alias for kafka.ConfigMap
	ConfigMap = kafka.ConfigMap

	// AdminClient is an alias for kafka.AdminClient
	AdminClient = kafka.AdminClient

	// ConfigEntry is an alias for kafka.ConfigEntry
	ConfigEntry = kafka.ConfigEntry

	// Message is an alias for kafka.Message
	Message = kafka.Message

	// TopicPartition is an alias for kafka.TopicPartition
	TopicPartition = kafka.TopicPartition

	// Error is an alias for kafka.Error
	Error = kafka.Error

	// TopicSpecification is an alias for kafka.TopicSpecification
	TopicSpecification = kafka.TopicSpecification
)

// Constants from the kafka package
const (
	// PartitionAny means to use any partition
	PartitionAny = kafka.PartitionAny

	// ErrTimedOut is a timeout error
	ErrTimedOut = kafka.ErrTimedOut

	// ErrNoError indicates no error
	ErrNoError = kafka.ErrNoError
)
