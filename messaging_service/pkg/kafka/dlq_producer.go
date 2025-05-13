package kafka

import (
	"context"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
)

// DLQProducer interface defines methods for handling dead-letter queues
type DLQProducer interface {
	SendToDLQ(ctx context.Context, messageID string, value []byte, headers []kafka.Header, failureReason string) error
	Close() error
}

// DLQProducerImpl implements the DLQProducer interface
type DLQProducerImpl struct {
	writer *kafka.Writer
	topic  string
}

// NewDLQProducer creates a new DLQ producer for the specified topic
func NewDLQProducer(cfg KafkaConfig) DLQProducer {
	// Get DLQ topic name from the source topic
	dlqTopic := GetDLQTopicName(cfg.Topic)

	writer := &kafka.Writer{
		Addr:     kafka.TCP(cfg.Brokers...),
		Topic:    dlqTopic,
		Balancer: GetBalancer(cfg.BalancerType),
	}

	return &DLQProducerImpl{
		writer: writer,
		topic:  cfg.Topic,
	}
}

// NewMockDLQProducer creates a new DLQ producer with a mock writer (for testing)
func NewMockDLQProducer(topic string, writer *kafka.Writer) DLQProducer {
	return &DLQProducerImpl{
		writer: writer,
		topic:  topic,
	}
}

// SendToDLQ sends a failed message to the appropriate DLQ
func (d *DLQProducerImpl) SendToDLQ(ctx context.Context, messageID string, value []byte, headers []kafka.Header, failureReason string) error {
	// Generate a message ID if not provided
	if messageID == "" {
		messageID = fmt.Sprintf("%s-%d", d.topic, time.Now().UnixNano())
	}

	// Add DLQ-specific headers
	dlqHeaders := append(headers, []kafka.Header{
		{
			Key:   "x-dlq-source-topic",
			Value: []byte(d.topic),
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
	LogDLQEvent(messageID, d.topic, messageID, failureReason, string(value))

	// Write the message to the DLQ topic
	return d.writer.WriteMessages(ctx, kafka.Message{
		Key:     []byte(messageID),
		Value:   value,
		Headers: dlqHeaders,
	})
}

func (d *DLQProducerImpl) Close() error {
	return d.writer.Close()
}
