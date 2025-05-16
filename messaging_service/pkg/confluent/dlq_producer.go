package confluent

import (
	"context"
	"fmt"
	"time"

	"messaging_service/pkg/logger"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

// DLQConfig holds configuration for a dead-letter queue
type DLQConfig struct {
	// MaxRetries before sending to DLQ
	MaxRetries int

	// Suffix to append to topic name to create DLQ topic name
	TopicSuffix string
}

// DefaultDLQConfig returns default configuration for a DLQ
func DefaultDLQConfig() DLQConfig {
	return DLQConfig{
		MaxRetries:  3,
		TopicSuffix: "-dlq",
	}
}

// DLQProducerImpl implements the DLQProducer interface
type DLQProducerImpl struct {
	Producer KafkaProducerInterface
	Topic    string
	Config   KafkaConfig
}

// NewDLQProducer creates a new DLQ producer
func NewDLQProducer(cfg KafkaConfig) (DLQProducer, error) {
	// Create DLQ topic name
	dlqTopic := GetDLQTopicName(cfg.Topic)

	// Create a copy of the config with the DLQ topic
	dlqConfig := cfg
	dlqConfig.Topic = dlqTopic

	// Get producer configuration
	producerConfig := dlqConfig.ToConfluentProducerConfig()

	// Create the producer
	producer, err := kafka.NewProducer(producerConfig)
	if err != nil {
		logger.LogErrorEvent("", "kafka_dlq_producer_creation", dlqTopic, "error",
			fmt.Sprintf("Failed to create DLQ producer: %v", err))
		return nil, fmt.Errorf("failed to create DLQ producer: %w", err)
	}

	// Start monitoring delivery reports
	go KafkaMonitorDeliveryReports(producer, dlqTopic)

	return &DLQProducerImpl{
		Producer: producer,
		Topic:    dlqTopic,
		Config:   dlqConfig,
	}, nil
}

// SendToDLQ sends a failed message to the DLQ
func (d *DLQProducerImpl) SendToDLQ(ctx context.Context, messageID string, value []byte, headers []Header, failureReason string) error {
	// Add DLQ-specific headers
	dlqHeaders := append(headers, []Header{
		{
			Key:   "x-dlq-source-topic",
			Value: []byte(d.Config.Topic),
		},
		{
			Key:   "x-dlq-failure-reason",
			Value: []byte(failureReason),
		},
		{
			Key:   "x-dlq-timestamp",
			Value: []byte(time.Now().Format(time.RFC3339)),
		},
	}...)

	// Log the DLQ event
	logger.LogEvent("", "kafka_dlq_event", messageID, "info",
		fmt.Sprintf("Sending message to DLQ %s: %s", d.Topic, failureReason))

	// Create Kafka message
	msg := &Message{
		TopicPartition: TopicPartition{Topic: &d.Topic, Partition: PartitionAny},
		Key:            []byte(messageID),
		Value:          value,
		Headers:        dlqHeaders,
		Timestamp:      time.Now(),
	}

	// Produce the message
	err := d.Producer.Produce(msg, nil)
	if err != nil {
		logger.LogErrorEvent("", "kafka_dlq_send_failed", d.Topic, "error",
			fmt.Sprintf("Failed to send message to DLQ: %v", err))
		return fmt.Errorf("failed to send message to DLQ: %w", err)
	}

	// Flush to ensure delivery
	remaining := d.Producer.Flush(int(d.Config.WriteTimeout.Milliseconds()))
	if remaining > 0 {
		logger.LogWarnEvent("", "kafka_dlq_flush_incomplete", d.Topic, "warn",
			fmt.Sprintf("%d messages still in queue after flush timeout", remaining))
	}

	return nil
}

// Close closes the DLQ producer
func (d *DLQProducerImpl) Close() error {
	// Flush any pending messages
	remaining := d.Producer.Flush(int(d.Config.WriteTimeout.Milliseconds()))
	if remaining > 0 {
		logger.LogWarnEvent("", "kafka_dlq_close_incomplete", d.Topic, "warn",
			fmt.Sprintf("%d messages still in queue after close flush timeout", remaining))
	}

	// Close the producer
	d.Producer.Close()
	return nil
}

// GetDLQTopicName returns the DLQ topic name for a given topic
func GetDLQTopicName(topic string) string {
	return topic + "-dlq"
}
