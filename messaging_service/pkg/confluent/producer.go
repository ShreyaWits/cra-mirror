package confluent

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"time"

	"messaging_service/pkg/logger"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

// KafkaProducerInterface defines the interface for the Kafka producer with required methods

// Producer interface defines methods for producing messages to Kafka

// ProducerImpl implements the Producer interface using confluent-kafka-go
type ProducerImpl struct {
	KafkaProducer KafkaProducerInterface
	Config   KafkaConfig
	Topic    string
}

// NewProducer creates a new Producer instance for a specific topic
func NewProducer(cfg KafkaConfig) (Producer, error) {
	if cfg.Topic == "" {
		return nil, fmt.Errorf("kafka topic must be specified")
	}

	// Create Confluent Kafka producer config
	producerConfig := cfg.ToConfluentProducerConfig()

	// Generate a unique transactional ID for exactly-once semantics
	needsTransactionInit := false
	if cfg.DeliverySemantics == ExactlyOnce && cfg.ExactlyOnceConfig.EnableTransactions {
		transactionID := cfg.ExactlyOnceConfig.TransactionalIDPrefix
		if transactionID == "" {
			transactionID = "txn"
		}
		transactionID = fmt.Sprintf("%s-%s-%d", transactionID, cfg.Topic, time.Now().UnixNano())
		_ = producerConfig.SetKey("transactional.id", transactionID)
		_ = producerConfig.SetKey("enable.idempotence", true)
		needsTransactionInit = true
	}

	// Create producer
	kafkaProducer, err := kafka.NewProducer(producerConfig)
	if err != nil {
		logger.LogErrorEvent("", "kafka_producer_creation", cfg.Topic, "error",
			fmt.Sprintf("Failed to create Kafka producer: %v", err))
		return nil, err
	}

	// Initialize transactions if needed
	if needsTransactionInit && !cfg.SkipTransactionInit {
		maxRetries := 5
		baseBackoffMs := 500
		maxBackoffMs := 10000 // 10 seconds max backoff
		var initErr error

		// The producer must call InitTransactions before any other transaction operations
		for i := 0; i < maxRetries; i++ {
			initCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			initErr = kafkaProducer.InitTransactions(initCtx)
			cancel()

			if initErr == nil {
				logger.LogEvent("", "kafka_transaction_init_success", cfg.Topic, "info",
					"Successfully initialized transactions for producer")
				break
			}

			// Check for coordinator-related errors that might be temporary
			kafkaErr, isKafkaErr := initErr.(kafka.Error)
			if isKafkaErr && (kafkaErr.Code() == kafka.ErrCoordinatorNotAvailable ||
				kafkaErr.Code() == kafka.ErrCoordinatorLoadInProgress) {

				// This is a recoverable error, so we'll retry with backoff
				// Calculate backoff with exponential increase and jitter
				jitter := rand.Intn(100)
				backoffMs := int(math.Min(float64(baseBackoffMs*int(math.Pow(2, float64(i)))+jitter), float64(maxBackoffMs)))
				backoffDuration := time.Duration(backoffMs) * time.Millisecond

				logger.LogWarnEvent("", "kafka_transaction_init_retry", cfg.Topic, "warn",
					fmt.Sprintf("Transaction coordinator not ready, retrying in %d ms (attempt %d/%d): %v",
						backoffMs, i+1, maxRetries, initErr))

				time.Sleep(backoffDuration)
				continue
			}

			// Not a recoverable coordinator error, break the loop
			break
		}

		if initErr != nil {
			// Close producer and return error if initialization fails after retries
			kafkaProducer.Close()
			logger.LogErrorEvent("", "kafka_transaction_init_failed", cfg.Topic, "error",
				fmt.Sprintf("Failed to initialize transactions after %d attempts: %v", maxRetries, initErr))
			return nil, fmt.Errorf("failed to initialize transactions after %d attempts: %w", maxRetries, initErr)
		}
	}

	// Start delivery report monitoring goroutine
	go KafkaMonitorDeliveryReports(kafkaProducer, cfg.Topic)

	return &ProducerImpl{
		KafkaProducer: kafkaProducer,
		Config:   cfg,
		Topic:    cfg.Topic,
	}, nil
}

// WriteWithRetry sends a message to Kafka with retry logic
func (p *ProducerImpl) WriteWithRetry(ctx context.Context, key, value []byte, headers []Header) error {
	// Add timestamp header if not already present
	hasTimestamp := false
	for _, header := range headers {
		if string(header.Key) == "timestamp" {
			hasTimestamp = true
			break
		}
	}

	if !hasTimestamp {
		headers = append(headers, Header{
			Key:   "timestamp",
			Value: []byte(time.Now().Format(time.RFC3339)),
		})
	}

	msg := &kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &p.Topic, Partition: kafka.PartitionAny},
		Key:            key,
		Value:          value,
		Headers:        toKafkaHeaders(headers),
		Timestamp:      time.Now(),
	}

	// Get retry settings from config or use defaults
	maxRetries := p.Config.MaxAttempts
	if maxRetries <= 0 {
		maxRetries = 3
	}

	baseBackoffMs := p.Config.RetryBackoffMs
	if baseBackoffMs <= 0 {
		baseBackoffMs = 100
	}

	maxBackoffMs := 30000 // 30 seconds max backoff

	var err error
	for i := 0; i < maxRetries; i++ {
		err = p.KafkaProducer.Produce(msg, nil)
		if err == nil {
			// If the message was sent successfully, flush to ensure it's delivered
			remaining := p.KafkaProducer.Flush(int(p.Config.WriteTimeout.Milliseconds()))
			if remaining > 0 {
				logger.LogWarnEvent("", "kafka_flush_incomplete", p.Topic, "warn",
					fmt.Sprintf("%d messages still in queue after flush timeout", remaining))
			}
			return nil
		}

		// Log retry attempt
		logger.LogWarnEvent("", "kafka_publish_retry", p.Topic, "attempt",
			fmt.Sprintf("%d/%d", i+1, maxRetries))

		// Calculate backoff with exponential increase and jitter
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

	logger.LogErrorEvent("", "kafka_publish_failed", p.Topic, "error",
		fmt.Sprintf("Failed to publish message after %d retries: %v", maxRetries, err))
	return fmt.Errorf("failed to publish message after %d retries: %w", maxRetries, err)
}

// BeginTransaction starts a new transaction for exactly-once semantics
func (p *ProducerImpl) BeginTransaction() error {
	if p.Config.DeliverySemantics != ExactlyOnce || !p.Config.ExactlyOnceConfig.EnableTransactions {
		return fmt.Errorf("transactions only supported with exactly-once semantics")
	}

	// Add retry logic for transaction initiation
	maxRetries := 3
	baseBackoffMs := 200
	maxBackoffMs := 2000 // 2 seconds max backoff
	var err error

	for i := 0; i < maxRetries; i++ {
		// Attempt to begin the transaction
		err = p.KafkaProducer.BeginTransaction()
		if err == nil {
			// Success
			return nil
		}

		// Check if the error is recoverable
		kafkaErr, isKafkaErr := err.(kafka.Error)
		if isKafkaErr && (kafkaErr.Code() == kafka.ErrCoordinatorNotAvailable ||
			kafkaErr.Code() == kafka.ErrCoordinatorLoadInProgress) {

			// This is a recoverable error, so we'll retry with backoff
			// Calculate backoff with exponential increase and jitter
			jitter := rand.Intn(50)
			backoffMs := int(math.Min(float64(baseBackoffMs*int(math.Pow(2, float64(i)))+jitter), float64(maxBackoffMs)))
			backoffDuration := time.Duration(backoffMs) * time.Millisecond

			logger.LogWarnEvent("", "kafka_transaction_begin_retry", p.Topic, "warn",
				fmt.Sprintf("Begin transaction failed, retrying in %d ms (attempt %d/%d): %v",
					backoffMs, i+1, maxRetries, err))

			time.Sleep(backoffDuration)
			continue
		}

		// Not a recoverable error, break the loop
		break
	}

	// Report the final error
	logger.LogErrorEvent("", "kafka_transaction_begin_failed", p.Topic, "error",
		fmt.Sprintf("Failed to begin transaction after %d attempts: %v", maxRetries, err))
	return fmt.Errorf("failed to begin transaction after %d attempts: %w", maxRetries, err)
}

// CommitTransaction commits the current transaction
func (p *ProducerImpl) CommitTransaction(ctx context.Context) error {
	if p.Config.DeliverySemantics != ExactlyOnce || !p.Config.ExactlyOnceConfig.EnableTransactions {
		return fmt.Errorf("transactions only supported with exactly-once semantics")
	}

	timeout := p.Config.WriteTimeout
	if timeout <= 0 {
		timeout = 5 * time.Second
	}

	// Add retry logic for transaction commits
	maxRetries := 3
	baseBackoffMs := 200
	maxBackoffMs := 2000 // 2 seconds max backoff
	var err error

	for i := 0; i < maxRetries; i++ {
		// Set context timeout for transaction commit
		ctxWithTimeout, cancel := context.WithTimeout(ctx, timeout)

		// Commit the transaction
		err = p.KafkaProducer.CommitTransaction(ctxWithTimeout)
		cancel() // Always cancel the context to avoid leaks

		if err == nil {
			// Success
			return nil
		}

		// Check if the error is recoverable
		kafkaErr, isKafkaErr := err.(kafka.Error)
		if isKafkaErr && (kafkaErr.Code() == kafka.ErrCoordinatorNotAvailable ||
			kafkaErr.Code() == kafka.ErrCoordinatorLoadInProgress) {

			// This is a recoverable error, so we'll retry with backoff
			// Calculate backoff with exponential increase and jitter
			jitter := rand.Intn(50)
			backoffMs := int(math.Min(float64(baseBackoffMs*int(math.Pow(2, float64(i)))+jitter), float64(maxBackoffMs)))
			backoffDuration := time.Duration(backoffMs) * time.Millisecond

			logger.LogWarnEvent("", "kafka_transaction_commit_retry", p.Topic, "warn",
				fmt.Sprintf("Commit transaction failed, retrying in %d ms (attempt %d/%d): %v",
					backoffMs, i+1, maxRetries, err))

			// Wait before retrying, but respect the context
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(backoffDuration):
				continue
			}
		}

		// Not a recoverable error, break the loop
		break
	}

	// Report the final error
	logger.LogErrorEvent("", "kafka_transaction_commit_failed", p.Topic, "error",
		fmt.Sprintf("Failed to commit transaction after %d attempts: %v", maxRetries, err))
	return fmt.Errorf("failed to commit transaction after %d attempts: %w", maxRetries, err)
}

// AbortTransaction aborts the current transaction
func (p *ProducerImpl) AbortTransaction(ctx context.Context) error {
	if p.Config.DeliverySemantics != ExactlyOnce || !p.Config.ExactlyOnceConfig.EnableTransactions {
		return fmt.Errorf("transactions only supported with exactly-once semantics")
	}

	timeout := p.Config.WriteTimeout
	if timeout <= 0 {
		timeout = 5 * time.Second
	}

	// Add retry logic for transaction aborts
	maxRetries := 3
	baseBackoffMs := 200
	maxBackoffMs := 2000 // 2 seconds max backoff
	var err error

	for i := 0; i < maxRetries; i++ {
		// Set context timeout for transaction abort
		ctxWithTimeout, cancel := context.WithTimeout(ctx, timeout)

		// Abort the transaction
		err = p.KafkaProducer.AbortTransaction(ctxWithTimeout)
		cancel() // Always cancel the context to avoid leaks

		if err == nil {
			// Success
			return nil
		}

		// Check if the error is recoverable
		kafkaErr, isKafkaErr := err.(kafka.Error)
		if isKafkaErr && (kafkaErr.Code() == kafka.ErrCoordinatorNotAvailable ||
			kafkaErr.Code() == kafka.ErrCoordinatorLoadInProgress) {

			// This is a recoverable error, so we'll retry with backoff
			// Calculate backoff with exponential increase and jitter
			jitter := rand.Intn(50)
			backoffMs := int(math.Min(float64(baseBackoffMs*int(math.Pow(2, float64(i)))+jitter), float64(maxBackoffMs)))
			backoffDuration := time.Duration(backoffMs) * time.Millisecond

			logger.LogWarnEvent("", "kafka_transaction_abort_retry", p.Topic, "warn",
				fmt.Sprintf("Abort transaction failed, retrying in %d ms (attempt %d/%d): %v",
					backoffMs, i+1, maxRetries, err))

			// Wait before retrying, but respect the context
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(backoffDuration):
				continue
			}
		}

		// Not a recoverable error, break the loop
		break
	}

	// Report the final error
	logger.LogErrorEvent("", "kafka_transaction_abort_failed", p.Topic, "error",
		fmt.Sprintf("Failed to abort transaction after %d attempts: %v", maxRetries, err))
	return fmt.Errorf("failed to abort transaction after %d attempts: %w", maxRetries, err)
}

// Close safely closes the Kafka producer and frees resources
func (p *ProducerImpl) Close() error {
	// Flush any pending messages
	remaining := p.KafkaProducer.Flush(int(p.Config.WriteTimeout.Milliseconds()))
	if remaining > 0 {
		logger.LogWarnEvent("", "kafka_close_incomplete", p.Topic, "warn",
			fmt.Sprintf("%d messages still in queue after close flush timeout", remaining))
	}

	// Close the producer
	p.KafkaProducer.Close()
	return nil
}

// Helper function to convert our header type to Kafka headers
func toKafkaHeaders(headers []Header) []kafka.Header {
	kafkaHeaders := make([]kafka.Header, len(headers))
	for i, header := range headers {
		kafkaHeaders[i] = kafka.Header{
			Key:   string(header.Key),
			Value: header.Value,
		}
	}
	return kafkaHeaders
}
