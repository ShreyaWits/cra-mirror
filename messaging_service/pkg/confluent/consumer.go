package confluent

import (
	"context"
	"fmt"
	"sync"
	"time"

	"messaging_service/pkg/logger"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

// ConsumerError represents an error that can occur during message processing
type ConsumerError struct {
	Err       error
	Permanent bool
	Reason    string
}

func (e *ConsumerError) Error() string {
	return fmt.Sprintf("%s: %v", e.Reason, e.Err)
}

// NewConsumerError creates a new ConsumerError
func NewConsumerError(err error, permanent bool, reason string) *ConsumerError {
	return &ConsumerError{
		Err:       err,
		Permanent: permanent,
		Reason:    reason,
	}
}

// Consumer interface defines methods for consuming messages from Kafka
type Consumer interface {
	// Start starts consuming messages from Kafka
	Start(ctx context.Context)

	// Close closes the consumer and releases resources
	Close() error
}

// ConfluentConsumer implements the Consumer interface using confluent-kafka-go
type ConfluentConsumer struct {
	consumer     *KafkaConsumer
	config       KafkaConfig
	topic        string
	groupID      string
	handler      func([]byte) error
	maxRetries   int
	retryTimeout time.Duration
	dlqProducer  DLQProducer
}

// DLQProducer interface defines methods for sending messages to a dead-letter queue
type DLQProducer interface {
	SendToDLQ(ctx context.Context, messageID string, value []byte, headers []Header, failureReason string) error
	Close() error
}

// NewConsumer creates a new Consumer with the given configuration
func NewConsumer(cfg KafkaConfig, groupID string, handler func([]byte) error) (Consumer, error) {
	if cfg.Topic == "" {
		return nil, fmt.Errorf("kafka topic must be specified")
	}

	if groupID == "" {
		return nil, fmt.Errorf("consumer group ID must be specified")
	}

	// Create Confluent Kafka consumer config
	consumerConfig := cfg.ToConfluentConsumerConfig(groupID)

	// Create consumer
	consumer, err := kafka.NewConsumer(consumerConfig)
	if err != nil {
		logger.LogErrorEvent("", "kafka_consumer_creation", cfg.Topic, "error",
			fmt.Sprintf("Failed to create Kafka consumer: %v", err))
		return nil, err
	}

	// Create DLQ producer
	dlqProducer, err := NewDLQProducer(cfg)
	if err != nil {
		// Close consumer since we couldn't create DLQ producer
		consumer.Close()
		logger.LogErrorEvent("", "kafka_dlq_producer_creation", cfg.Topic, "error",
			fmt.Sprintf("Failed to create DLQ producer: %v", err))
		return nil, err
	}

	// Set default retry settings
	maxRetries := 3
	if cfg.MaxAttempts > 0 {
		maxRetries = cfg.MaxAttempts
	}

	retryTimeout := 5 * time.Second
	if cfg.WriteTimeout > 0 {
		retryTimeout = cfg.WriteTimeout
	}

	return &ConfluentConsumer{
		consumer:     consumer,
		config:       cfg,
		topic:        cfg.Topic,
		groupID:      groupID,
		handler:      handler,
		maxRetries:   maxRetries,
		retryTimeout: retryTimeout,
		dlqProducer:  dlqProducer,
	}, nil
}

// Start begins consuming messages from Kafka
func (c *ConfluentConsumer) Start(ctx context.Context) {
	// Subscribe to the topic
	err := c.consumer.Subscribe(c.topic, nil)
	if err != nil {
		logger.LogErrorEvent("", "kafka_subscribe_error", c.topic, "error",
			fmt.Sprintf("Failed to subscribe to topic: %v", err))
		return
	}

	// Set up worker pool
	const workerCount = 10
	msgChan := make(chan *Message, 100)

	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Start worker goroutines
	wg.Add(workerCount)
	for i := 0; i < workerCount; i++ {
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case msg, ok := <-msgChan:
					if !ok {
						return
					}
					c.processMessage(ctx, msg)
				}
			}
		}()
	}

	// Main polling loop
	logger.LogEvent("", "kafka_consumer_start", c.topic, "info",
		fmt.Sprintf("Started consuming from topic %s with group %s", c.topic, c.groupID))

	defer func() {
		close(msgChan)
		wg.Wait()
		logger.LogEvent("", "kafka_consumer_stop", c.topic, "info",
			fmt.Sprintf("Stopped consuming from topic %s with group %s", c.topic, c.groupID))
	}()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			msg, err := c.consumer.ReadMessage(100 * time.Millisecond)
			if err != nil {
				// Ignore timeout errors, which happen when no messages are available
				if err.(Error).Code() != ErrTimedOut {
					logger.LogErrorEvent("", "kafka_read_error", c.topic, "error",
						fmt.Sprintf("Error reading message: %v", err))
				}
				continue
			}

			// For exactly-once semantics with transactions, we only want to see committed messages
			if c.config.DeliverySemantics == ExactlyOnce &&
				c.config.ExactlyOnceConfig.IsolationLevel == "read_committed" {
				// Message is already filtered at broker level with isolation.level=read_committed
			}

			// Send message to worker pool
			select {
			case msgChan <- msg:
				// Message sent to worker
			default:
				// Channel full, process in this goroutine
				c.processMessage(ctx, msg)
			}
		}
	}
}

// processMessage processes a single Kafka message
func (c *ConfluentConsumer) processMessage(ctx context.Context, msg *Message) {
	// Extract message ID from key
	messageID := string(msg.Key)
	if messageID == "" {
		messageID = fmt.Sprintf("%s-%d", c.topic, msg.Timestamp.UnixNano())
	}

	// Process with retries
	var attempt int
	var err error
	var permanentFailure bool
	var failureReason string

	for attempt = 0; attempt < c.maxRetries; attempt++ {
		err = c.invokeHandler(ctx, msg.Value)
		if err == nil {
			// Successfully processed, commit offset
			if !c.config.ConsumerConfig.AutoCommit {
				_, err = c.consumer.CommitMessage(msg)
				if err != nil {
					logger.LogErrorEvent("", "kafka_commit_failed", c.topic, "error",
						fmt.Sprintf("Failed to commit offset for message %s: %v", messageID, err))
				}
			}
			return
		}

		// Check if this is a permanent error
		var consumerErr *ConsumerError
		if e, ok := err.(*ConsumerError); ok {
			consumerErr = e
			permanentFailure = consumerErr.Permanent
			failureReason = consumerErr.Reason
			if permanentFailure {
				break // Don't retry permanent failures
			}
		} else {
			// Default error handling if not a consumer error
			failureReason = "unknown error"
		}

		// Calculate backoff duration
		backoff := time.Duration(100*(2^attempt)) * time.Millisecond
		if backoff > time.Second {
			backoff = time.Second
		}

		// Wait before retrying
		select {
		case <-ctx.Done():
			return
		case <-time.After(backoff):
			continue
		}
	}

	// If we got here, all retries failed - send to DLQ
	if c.dlqProducer != nil {
		dlqErr := c.dlqProducer.SendToDLQ(ctx, messageID, msg.Value, msg.Headers, failureReason)
		if dlqErr != nil {
			logger.LogErrorEvent("", "kafka_dlq_send_failed", c.topic, "error",
				fmt.Sprintf("Failed to send message to DLQ: %v", dlqErr))
		} else {
			logger.LogEvent("", "kafka_dlq_sent", c.topic, "info",
				fmt.Sprintf("Message %s sent to DLQ after %d retries: %s", messageID, attempt, failureReason))
		}
	}

	// Commit the message after sending to DLQ if not auto-committing
	if !c.config.ConsumerConfig.AutoCommit {
		_, err = c.consumer.CommitMessage(msg)
		if err != nil {
			logger.LogErrorEvent("", "kafka_commit_failed", c.topic, "error",
				fmt.Sprintf("Failed to commit offset after DLQ for message %s: %v", messageID, err))
		}
	}
}

// invokeHandler invokes the message handler with timeout
func (c *ConfluentConsumer) invokeHandler(ctx context.Context, payload []byte) error {
	ctxWithTimeout, cancel := context.WithTimeout(ctx, c.retryTimeout)
	defer cancel()

	resultChan := make(chan error, 1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				resultChan <- NewConsumerError(fmt.Errorf("panic: %v", r), true, "handler_panic")
			}
		}()
		resultChan <- c.handler(payload)
	}()

	select {
	case <-ctxWithTimeout.Done():
		return NewConsumerError(ctxWithTimeout.Err(), false, "handler_timeout")
	case err := <-resultChan:
		return err
	}
}

// Close closes the consumer and releases resources
func (c *ConfluentConsumer) Close() error {
	var err error
	if c.consumer != nil {
		err = c.consumer.Close()
		if err != nil {
			logger.LogErrorEvent("", "kafka_consumer_close_error", c.topic, "error",
				fmt.Sprintf("Error closing consumer: %v", err))
		}
	}

	if c.dlqProducer != nil {
		dlqErr := c.dlqProducer.Close()
		if dlqErr != nil && err == nil {
			err = dlqErr
		}
	}

	return err
}
