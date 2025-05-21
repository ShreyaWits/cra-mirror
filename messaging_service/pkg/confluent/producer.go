

package confluent

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"time"

	"messaging_service/pkg/observability"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

// KafkaProducerInterface defines the interface for the Kafka producer with required methods

// Producer interface defines methods for producing messages to Kafka

// ProducerImpl implements the Producer interface using confluent-kafka-go
type ProducerImpl struct {
	KafkaProducer KafkaProducerInterface
	Config        KafkaConfig
	Topic         string
	obs           *observability.ObservabilityStack
}

// NewProducer creates a new Producer instance for a specific topic
func NewProducer(ctx context.Context, cfg KafkaConfig, obs *observability.ObservabilityStack) (Producer, error) {
	functionName := "NewProducer"

	_, span := obs.TracerService.StartTracer(ctx, functionName)
	defer obs.TracerService.StopSpan(span)

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
		obs.LoggerService.Error(ctx, "Failed to create Kafka producer", map[string]interface{}{
			"topic": cfg.Topic,
			"error": err.Error(),
		})
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
				obs.LoggerService.Info(ctx, "Successfully initialized transactions for producer", map[string]interface{}{
					"topic": cfg.Topic,
				})
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

				obs.LoggerService.Warn(ctx, "Transaction coordinator not ready, retrying", map[string]interface{}{
					"topic":      cfg.Topic,
					"attempt":    i + 1,
					"maxRetries": maxRetries,
					"backoffMs":  backoffMs,
					"error":      initErr.Error(),
				})

				time.Sleep(backoffDuration)
				continue
			}

			// Not a recoverable coordinator error, break the loop
			break
		}

		if initErr != nil {
			// Close producer and return error if initialization fails after retries
			kafkaProducer.Close()
			obs.LoggerService.Error(ctx, "Failed to initialize transactions", map[string]interface{}{
				"topic":    cfg.Topic,
				"attempts": maxRetries,
				"error":    initErr.Error(),
			})
			return nil, fmt.Errorf("failed to initialize transactions after %d attempts: %w", maxRetries, initErr)
		}
	}

	// Start delivery report monitoring goroutine
	go KafkaMonitorDeliveryReports(kafkaProducer, cfg.Topic, obs)

	obs.LoggerService.Info(ctx, "Kafka producer created successfully", map[string]interface{}{
		"topic": cfg.Topic,
	})

	return &ProducerImpl{
		KafkaProducer: kafkaProducer,
		Config:        cfg,
		Topic:         cfg.Topic,
		obs:           obs,
	}, nil
}

// WriteWithRetry sends a message to Kafka with retry logic
func (p *ProducerImpl) WriteWithRetry(ctx context.Context, key, value []byte, headers []Header) error {
	functionName := "WriteWithRetry"

	_, span := p.obs.TracerService.StartTracer(ctx, functionName)
	defer p.obs.TracerService.StopSpan(span)

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
				p.obs.LoggerService.Warn(ctx, "Messages still in queue after flush timeout", map[string]interface{}{
					"topic":     p.Topic,
					"remaining": remaining,
					"timeout":   p.Config.WriteTimeout.String(),
				})
			}
			return nil
		}

		// Log retry attempt
		p.obs.LoggerService.Warn(ctx, "Retrying message publish", map[string]interface{}{
			"topic":      p.Topic,
			"attempt":    i + 1,
			"maxRetries": maxRetries,
		})

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

	p.obs.LoggerService.Error(ctx, "Failed to publish message after retries", map[string]interface{}{
		"topic":    p.Topic,
		"attempts": maxRetries,
		"error":    err.Error(),
	})
	return fmt.Errorf("failed to publish message after %d retries: %w", maxRetries, err)
}

// BeginTransaction starts a new transaction for exactly-once semantics
func (p *ProducerImpl) BeginTransaction(ctx context.Context) error {
	functionName := "BeginTransaction"

	_, span := p.obs.TracerService.StartTracer(ctx, functionName)
	defer p.obs.TracerService.StopSpan(span)

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
			p.obs.LoggerService.Info(ctx, "Transaction begun successfully", map[string]interface{}{
				"topic": p.Topic,
			})
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

			p.obs.LoggerService.Warn(ctx, "Begin transaction failed, retrying", map[string]interface{}{
				"topic":      p.Topic,
				"attempt":    i + 1,
				"maxRetries": maxRetries,
				"backoffMs":  backoffMs,
				"error":      err.Error(),
			})

			time.Sleep(backoffDuration)
			continue
		}

		// Not a recoverable error, break the loop
		break
	}

	// Report the final error
	p.obs.LoggerService.Error(ctx, "Failed to begin transaction", map[string]interface{}{
		"topic":    p.Topic,
		"attempts": maxRetries,
		"error":    err.Error(),
	})
	return fmt.Errorf("failed to begin transaction after %d attempts: %w", maxRetries, err)
}

// CommitTransaction commits the current transaction
func (p *ProducerImpl) CommitTransaction(ctx context.Context) error {
	functionName := "CommitTransaction"

	_, span := p.obs.TracerService.StartTracer(ctx, functionName)
	defer p.obs.TracerService.StopSpan(span)

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
			p.obs.LoggerService.Info(ctx, "Transaction committed successfully", map[string]interface{}{
				"topic": p.Topic,
			})
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

			p.obs.LoggerService.Warn(ctx, "Commit transaction failed, retrying", map[string]interface{}{
				"topic":      p.Topic,
				"attempt":    i + 1,
				"maxRetries": maxRetries,
				"backoffMs":  backoffMs,
				"error":      err.Error(),
			})

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
	p.obs.LoggerService.Error(ctx, "Failed to commit transaction", map[string]interface{}{
		"topic":    p.Topic,
		"attempts": maxRetries,
		"error":    err.Error(),
	})
	return fmt.Errorf("failed to commit transaction after %d attempts: %w", maxRetries, err)
}

// AbortTransaction aborts the current transaction
func (p *ProducerImpl) AbortTransaction(ctx context.Context) error {
	functionName := "AbortTransaction"

	_, span := p.obs.TracerService.StartTracer(ctx, functionName)
	defer p.obs.TracerService.StopSpan(span)

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
			p.obs.LoggerService.Info(ctx, "Transaction aborted successfully", map[string]interface{}{
				"topic": p.Topic,
			})
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

			p.obs.LoggerService.Warn(ctx, "Abort transaction failed, retrying", map[string]interface{}{
				"topic":      p.Topic,
				"attempt":    i + 1,
				"maxRetries": maxRetries,
				"backoffMs":  backoffMs,
				"error":      err.Error(),
			})

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
	p.obs.LoggerService.Error(ctx, "Failed to abort transaction", map[string]interface{}{
		"topic":    p.Topic,
		"attempts": maxRetries,
		"error":    err.Error(),
	})
	return fmt.Errorf("failed to abort transaction after %d attempts: %w", maxRetries, err)
}

// Close safely closes the Kafka producer and frees resources
func (p *ProducerImpl) Close() error {
	functionName := "Close"

	ctx := context.Background()
	_, span := p.obs.TracerService.StartTracer(ctx, functionName)
	defer p.obs.TracerService.StopSpan(span)

	p.obs.LoggerService.Info(ctx, "Closing Kafka producer", map[string]interface{}{
		"topic": p.Topic,
	})

	// Flush any pending messages
	remaining := p.KafkaProducer.Flush(int(p.Config.WriteTimeout.Milliseconds()))
	if remaining > 0 {
		p.obs.LoggerService.Warn(ctx, "Messages still in queue after close flush timeout", map[string]interface{}{
			"topic":     p.Topic,
			"remaining": remaining,
			"timeout":   p.Config.WriteTimeout.String(),
		})
	}

	// Close the producer
	p.KafkaProducer.Close()

	p.obs.LoggerService.Info(ctx, "Kafka producer closed successfully", map[string]interface{}{
		"topic": p.Topic,
	})

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
