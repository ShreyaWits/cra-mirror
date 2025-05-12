package kafka

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"messaging_service/pkg/logger"

	"github.com/segmentio/kafka-go"
)

type BaseProducer interface {
	Write(ctx context.Context, writer *kafka.Writer, key, value []byte, headers []kafka.Header) error
	WriteWithRetry(ctx context.Context, writer *kafka.Writer, key, value []byte, headers []kafka.Header, maxRetries int) error
	Close(writer *kafka.Writer) error
	GetWriter(cfg KafkaConfig) *kafka.Writer
}

type Producer struct {
	mu sync.RWMutex
	// Removed writers map
}

// createWriter initializes the Kafka writer with the provided config.
// It should be done once during the application startup.
func CreateWriter(cfg KafkaConfig) *kafka.Writer {
	writer := &kafka.Writer{
		Addr:     kafka.TCP(cfg.Brokers...),
		Topic:    cfg.Topic,
		Balancer: GetBalancer(cfg.BalancerType),
		// Performance optimizations
		BatchSize:              1000,                  // Increased batch size for better throughput
		BatchTimeout:           10 * time.Millisecond, // Slightly increased to allow more batching
		BatchBytes:             2 * 1024 * 1024,       // 2MB batch size for better compression
		RequiredAcks:           kafka.RequireOne,      // Only wait for leader acknowledgment
		MaxAttempts:            3,                     // Reduced retry attempts
		ReadTimeout:            50 * time.Millisecond, // Reduced read timeout
		WriteTimeout:           50 * time.Millisecond, // Reduced write timeout
		Async:                  true,                  // Enable async publishing
		Compression:            kafka.Snappy,          // Enable compression
		AllowAutoTopicCreation: true,                  // Allow auto topic creation
	}

	// Set delivery semantics based on configuration
	switch cfg.DeliverySemantics {
	case ExactlyOnce:
		writer.RequiredAcks = kafka.RequireAll
		writer.MaxAttempts = 10
		writer.Async = false // Exactly-once requires synchronous publishing

		// Configure exactly-once settings
		if cfg.ExactlyOnceConfig.EnableIdempotence {
			// Enable idempotence by requiring all acks and disabling async
			writer.RequiredAcks = kafka.RequireAll
			writer.Async = false
		}
		if cfg.ExactlyOnceConfig.EnableTransactions {
			// For transactions, we need to use a transactional producer
			// This is handled at the application level by using BeginTransaction/CommitTransaction
			writer.RequiredAcks = kafka.RequireAll
			writer.Async = false
		}
	default:
		// At-least-once with performance optimizations
		writer.RequiredAcks = kafka.RequireOne
		writer.MaxAttempts = 3
		writer.Async = true
	}

	return writer
}

// NewProducer initializes a new producer
func NewProducer() *Producer {
	return &Producer{}
}

// GetWriter creates a new writer for each call
func (p *Producer) GetWriter(cfg KafkaConfig) *kafka.Writer {
	return CreateWriter(cfg)
}

// Write sends a message to Kafka with the provided writer.
func (p *Producer) Write(ctx context.Context, writer *kafka.Writer, key, value []byte, headers []kafka.Header) error {
	err := writer.WriteMessages(ctx, kafka.Message{
		Key:     key,
		Value:   value,
		Headers: headers,
		Time:    time.Now(), // Add timestamp for better monitoring
	})
	if err != nil {
		log.Printf("Kafka write error: %v", err)
		return fmt.Errorf("failed to write message: %w", err)
	}
	return nil
}

// WriteWithRetry tries to send a message to Kafka with optimized retry logic.
func (p *Producer) WriteWithRetry(ctx context.Context, writer *kafka.Writer, key, value []byte, headers []kafka.Header, maxRetries int) error {
	// Add timestamp header
	headers = append(headers, kafka.Header{
		Key:   "timestamp",
		Value: []byte(time.Now().Format(time.RFC3339)),
	})

	msg := kafka.Message{
		Key:     key,
		Value:   value,
		Headers: headers,
		Time:    time.Now(), // Set message timestamp
	}

	var err error
	for i := 0; i < maxRetries; i++ {
		err = writer.WriteMessages(ctx, msg)
		if err == nil {
			return nil
		}

		// Log retry attempt
		logger.LogWarnEvent("", "publish_retry", writer.Topic, "attempt", fmt.Sprintf("%d/%d", i+1, maxRetries))

		// Wait before retrying
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Duration(i+1) * 100 * time.Millisecond):
			continue
		}
	}

	return fmt.Errorf("failed to publish message after %d retries: %w", maxRetries, err)
}

// Close safely closes the Kafka writer.
func (p *Producer) Close(writer *kafka.Writer) error {
	if err := writer.Close(); err != nil {
		log.Printf("Error closing writer: %v", err)
		return fmt.Errorf("failed to close writer: %w", err)
	}
	return nil
}
