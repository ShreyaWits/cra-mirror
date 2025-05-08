package kafka

import (
	"context"
	"errors"
	"fmt"
	"notification-service/pkg/logger"
	"time"

	"github.com/segmentio/kafka-go"
)

// KafkaPublisher wraps the Kafka writer for producing messages
type KafkaPublisher struct {
	writer KafkaWriter
}

// InitKafkaPublisher initializes and returns a KafkaPublisher
func InitKafkaPublisher(brokers string) *KafkaPublisher {
	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers:  []string{brokers},
		Balancer: &kafka.LeastBytes{},
		Async:    false,
	})

	logger.Log.Println("Kafka publisher initialized")

	return &KafkaPublisher{
		writer: writer,
	}
}

// Publish sends a message to the specified Kafka topic
func (kp *KafkaPublisher) Publish(ctx context.Context, topic string, key, value []byte) error {
	if kp.writer == nil {
		return errors.New("Kafka writer is not initialized")
	}

	msg := kafka.Message{
		Topic: topic,
		Key:   key,
		Value: value,
		Time:  time.Now(),
	}

	if err := kp.writer.WriteMessages(ctx, msg); err != nil {
		logger.Log.Printf("Failed to publish message to topic %s: %v", topic, err)
		return fmt.Errorf("failed to write Kafka message: %w", err)
	}

	logger.Log.Printf("Message published to topic %s", topic)
	return nil
}

// Close gracefully shuts down the Kafka writer
func (kp *KafkaPublisher) Close() error {
	logger.Log.Println("Closing Kafka publisher...")
	if kp.writer == nil {
		return nil
	}

	if err := kp.writer.Close(); err != nil {
		logger.Log.Printf("Error closing writer: %v", err)
		return fmt.Errorf("failed to close Kafka writer: %w", err)
	}

	logger.Log.Println("Kafka publisher closed.")
	return nil
}
