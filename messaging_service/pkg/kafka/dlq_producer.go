package kafka

import (
	"context"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
)

type DLQProducer struct {
	writer *kafka.Writer
	topic  string
}

// NewDLQProducer creates a new DLQ producer for the specified topic
func NewDLQProducer(cfg KafkaConfig) *DLQProducer {
	// Get DLQ topic name from the source topic
	dlqTopic := GetDLQTopicName(cfg.Topic)

	writer := &kafka.Writer{
		Addr:     kafka.TCP(cfg.Brokers...),
		Topic:    dlqTopic,
		Balancer: GetBalancer(cfg.BalancerType),
	}

	return &DLQProducer{
		writer: writer,
		topic:  cfg.Topic,
	}
}

// SendToDLQ sends a failed message to the appropriate DLQ
func (d *DLQProducer) SendToDLQ(ctx context.Context, messageID string, value []byte, headers []kafka.Header, failureReason string) error {
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

func (d *DLQProducer) Close() error {
	return d.writer.Close()
}
