package kafka

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"time"

	"messaging_service/pkg/logger"

	"github.com/segmentio/kafka-go"
)

// Producer interface defines methods for producing messages to Kafka
type Producer interface {
	// Write sends a message to Kafka without retries
	Write(ctx context.Context, key, value []byte, headers []kafka.Header) error

	// WriteWithRetry sends a message to Kafka with retry logic
	WriteWithRetry(ctx context.Context, key, value []byte, headers []kafka.Header) error

	// Close closes the producer and frees resources
	Close() error

	// Topic returns the topic this producer writes to
	Topic() string
}

// ProducerImpl implements the Producer interface
type ProducerImpl struct {
	writer *kafka.Writer
	config KafkaConfig
}

// NewProducer creates a new Producer instance for a specific topic with the given configuration
func NewProducer(cfg KafkaConfig) Producer {
	if cfg.Topic == "" {
		logger.LogErrorEvent("", "kafka_producer_creation", "", "error", "Kafka topic must be specified")
		panic("Kafka topic must be specified")
	}

	return &ProducerImpl{
		writer: createWriter(cfg),
		config: cfg,
	}
}

// createWriter initializes a Kafka writer with the provided config.
// Should be called only once per producer instance.
func createWriter(cfg KafkaConfig) *kafka.Writer {
	// Set default timeouts if not provided
	readTimeout := cfg.ReadTimeout
	if readTimeout <= 0 {
		readTimeout = 10 * time.Second
	}

	writeTimeout := cfg.WriteTimeout
	if writeTimeout <= 0 {
		writeTimeout = 10 * time.Second
	}

	// Default batch settings
	batchSize := 100
	if cfg.BatchSize > 0 {
		batchSize = cfg.BatchSize
	}

	batchBytes := int64(1024 * 1024) // 1MB default
	if cfg.BatchBytes > 0 {
		batchBytes = int64(cfg.BatchBytes)
	}

	batchTimeout := 1 * time.Second
	if cfg.BatchTimeout > 0 {
		batchTimeout = cfg.BatchTimeout
	}

	// Use configured compression or default to snappy
	compression := kafka.Snappy
	if cfg.CompressionCodec != "" {
		switch cfg.CompressionCodec {
		case "gzip":
			compression = kafka.Gzip
		case "snappy":
			compression = kafka.Snappy
		case "lz4":
			compression = kafka.Lz4
		case "zstd":
			compression = kafka.Zstd
		default:
			// Default to Snappy
			compression = kafka.Snappy
		}
	}

	writer := &kafka.Writer{
		Addr:                   kafka.TCP(cfg.Brokers...),
		Topic:                  cfg.Topic,
		Balancer:               GetBalancer(cfg.BalancerType),
		BatchSize:              batchSize,
		BatchTimeout:           batchTimeout,
		BatchBytes:             batchBytes,
		ReadTimeout:            readTimeout,
		WriteTimeout:           writeTimeout,
		Compression:            compression,
		AllowAutoTopicCreation: true,
	}

	// Configure writer based on delivery semantics
	switch cfg.DeliverySemantics {
	case ExactlyOnce:
		// Exactly-once requires synchronous writes with all brokers acknowledging
		writer.RequiredAcks = kafka.RequireAll
		writer.Async = false

		if cfg.ExactlyOnceConfig.EnableIdempotence {
			writer.RequiredAcks = kafka.RequireAll
			writer.Async = false
		}
	case AtLeastOnce:
		// At-least-once requires at least one broker to acknowledge
		writer.RequiredAcks = kafka.RequireOne
		writer.Async = false // Don't use async for at-least-once guarantees
	default:
		// Default to at-least-once
		writer.RequiredAcks = kafka.RequireOne
		writer.Async = false
	}

	return writer
}

// Write sends a message to Kafka without retries
func (p *ProducerImpl) Write(ctx context.Context, key, value []byte, headers []kafka.Header) error {
	msg := kafka.Message{
		Key:     key,
		Value:   value,
		Headers: headers,
		Time:    time.Now(),
	}

	err := p.writer.WriteMessages(ctx, msg)
	if err != nil {
		logger.LogErrorEvent("", "kafka_write_error", p.config.Topic, "error",
			fmt.Sprintf("Failed to write message to Kafka: %v", err))
		return fmt.Errorf("failed to write message: %w", err)
	}
	return nil
}

// WriteWithRetry sends a message to Kafka with retry logic using exponential backoff with jitter
func (p *ProducerImpl) WriteWithRetry(ctx context.Context, key, value []byte, headers []kafka.Header) error {
	// Add timestamp header if not already present
	hasTimestamp := false
	for _, header := range headers {
		if header.Key == "timestamp" {
			hasTimestamp = true
			break
		}
	}

	if !hasTimestamp {
		headers = append(headers, kafka.Header{
			Key:   "timestamp",
			Value: []byte(time.Now().Format(time.RFC3339)),
		})
	}

	msg := kafka.Message{
		Key:     key,
		Value:   value,
		Headers: headers,
		Time:    time.Now(),
	}

	// Get retry settings from config or use defaults
	maxRetries := p.config.MaxAttempts
	if maxRetries <= 0 {
		maxRetries = 3
	}

	baseBackoffMs := p.config.RetryBackoffMs
	if baseBackoffMs <= 0 {
		baseBackoffMs = 100
	}

	maxBackoffMs := 30000 // 30 seconds max backoff

	var err error
	for i := 0; i < maxRetries; i++ {
		err = p.writer.WriteMessages(ctx, msg)
		if err == nil {
			return nil
		}

		// Log retry attempt
		logger.LogWarnEvent("", "kafka_publish_retry", p.config.Topic, "attempt",
			fmt.Sprintf("%d/%d", i+1, maxRetries))

		// Calculate backoff with exponential increase and jitter
		// Formula: baseBackoff * (2^attempt) + random jitter
		jitter := rand.Intn(100)
		backoffMs := int(math.Min(float64(baseBackoffMs*int(math.Pow(2, float64(i)))+jitter), float64(maxBackoffMs)))
		backoffDuration := time.Duration(backoffMs) * time.Millisecond

		// Wait before retrying
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(backoffDuration):
			continue
		}
	}

	logger.LogErrorEvent("", "kafka_publish_failed", p.config.Topic, "error",
		fmt.Sprintf("Failed to publish message after %d retries: %v", maxRetries, err))
	return fmt.Errorf("failed to publish message after %d retries: %w", maxRetries, err)
}

// Close safely closes the Kafka writer and frees resources
func (p *ProducerImpl) Close() error {
	if err := p.writer.Close(); err != nil {
		logger.LogErrorEvent("", "kafka_close_error", p.config.Topic, "error",
			fmt.Sprintf("Failed to close Kafka writer: %v", err))
		return fmt.Errorf("failed to close writer: %w", err)
	}
	return nil
}

// Topic returns the topic this producer writes to
func (p *ProducerImpl) Topic() string {
	return p.config.Topic
}
