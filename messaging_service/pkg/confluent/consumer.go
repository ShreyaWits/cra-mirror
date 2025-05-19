package confluent

import (
	"context"
	"fmt"
	"sync"
	"time"

	"messaging_service/pkg/errors"
	"messaging_service/pkg/observability"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

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
	obs          *observability.ObservabilityStack
}

// DLQProducer interface defines methods for sending messages to a dead-letter queue
type DLQProducer interface {
	SendToDLQ(ctx context.Context, messageID string, value []byte, headers []Header, failureReason string) error
	Close() error
}

// NewConsumer creates a new Consumer with the given configuration
func NewConsumer(cfg KafkaConfig, groupID string, handler func([]byte) error, obs *observability.ObservabilityStack) (Consumer, error) {
	functionName := "NewConsumer"
	ctx := context.Background()

	_, span := obs.TracerService.StartTracer(ctx, functionName)
	defer obs.TracerService.StopSpan(span)

	if cfg.Topic == "" {
		return nil, errors.NewCustomError(
			errors.SUBErrInvalidConfig,
			fmt.Errorf("kafka topic must be specified"),
		)
	}

	if groupID == "" {
		return nil, errors.NewCustomError(
			errors.SUBErrInvalidGroupID,
			fmt.Errorf("consumer group ID must be specified"),
		)
	}

	// Create Confluent Kafka consumer config
	consumerConfig := cfg.ToConfluentConsumerConfig(groupID)

	// Create consumer
	consumer, err := kafka.NewConsumer(consumerConfig)
	if err != nil {
		return nil, errors.NewCustomError(
			errors.SUBErrConsumerNotReady,
			fmt.Errorf("failed to create Kafka consumer: %w", err),
		)
	}

	// Create DLQ producer
	dlqProducer, err := NewDLQProducer(cfg, obs)
	if err != nil {
		// Close consumer since we couldn't create DLQ producer
		consumer.Close()
		return nil, errors.NewCustomError(
			errors.SUBErrConsumerNotReady,
			fmt.Errorf("failed to create DLQ producer: %w", err),
		)
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
		obs:          obs,
	}, nil
}

// Start begins consuming messages from Kafka
func (c *ConfluentConsumer) Start(ctx context.Context) {
	functionName := "Start"

	tCtx, span := c.obs.TracerService.StartTracer(ctx, functionName)
	defer c.obs.TracerService.StopSpan(span)

	// Subscribe to the topic
	err := c.consumer.Subscribe(c.topic, nil)
	if err != nil {
		c.obs.LoggerService.Error(tCtx, "Failed to subscribe to topic", map[string]interface{}{
			"topic":   c.topic,
			"groupID": c.groupID,
			"error":   err.Error(),
		})
		return
	}

	c.obs.LoggerService.Info(tCtx, tCtx, "Successfully subscribed to topic", map[string]interface{}{
		"topic":   c.topic,
		"groupID": c.groupID,
	})

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

	c.obs.LoggerService.Info(tCtx, "Started consumer workers", map[string]interface{}{
		"workerCount": workerCount,
		"topic":       c.topic,
	})

	defer func() {
		close(msgChan)
		wg.Wait()
		c.obs.LoggerService.Info(tCtx, "Consumer workers stopped", map[string]interface{}{
			"topic": c.topic,
		})
	}()

	for {
		select {
		case <-ctx.Done():
			c.obs.LoggerService.Info(tCtx, "Consumer context cancelled, stopping consumption", map[string]interface{}{
				"topic": c.topic,
			})
			return
		default:
			msg, err := c.consumer.ReadMessage(100 * time.Millisecond)
			if err != nil {
				// Ignore timeout errors, which happen when no messages are available
				if err.(Error).Code() != ErrTimedOut {
					c.obs.LoggerService.Warn(tCtx, "Error reading message", map[string]interface{}{
						"topic": c.topic,
						"error": err.Error(),
					})
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
				c.obs.LoggerService.Warn(tCtx, "Worker channel full, processing message in main goroutine", map[string]interface{}{
					"topic": c.topic,
				})
				c.processMessage(ctx, msg)
			}
		}
	}
}

// processMessage processes a single Kafka message
func (c *ConfluentConsumer) processMessage(ctx context.Context, msg *Message) {
	functionName := "ProcessMessage"

	_, span := c.obs.TracerService.StartTracer(ctx, functionName)
	defer c.obs.TracerService.StopSpan(span)

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
		if attempt > 0 {
			c.obs.LoggerService.Info(ctx, "Retrying message processing", map[string]interface{}{
				"topic":      c.topic,
				"messageID":  messageID,
				"attempt":    attempt,
				"maxRetries": c.maxRetries,
			})
		}

		err = c.invokeHandler(ctx, msg.Value)
		if err == nil {
			// Successfully processed, commit offset
			if !c.config.ConsumerConfig.AutoCommit {
				_, err = c.consumer.CommitMessage(msg)
				if err != nil {
					c.obs.LoggerService.Warn(ctx, "Failed to commit message offset", map[string]interface{}{
						"topic":     c.topic,
						"messageID": messageID,
						"error":     err.Error(),
					})
				}
			}
			return
		}

		// Check if this is a permanent error
		if e, ok := err.(*errors.CustomError); ok {
			permanentFailure = true // CustomError indicates permanent failure
			failureReason = e.Error()
			if permanentFailure {
				c.obs.LoggerService.Warn(ctx, "Permanent error encountered, skipping retries", map[string]interface{}{
					"topic":     c.topic,
					"messageID": messageID,
					"error":     failureReason,
				})
				break // Don't retry permanent failures
			}
		} else {
			// Default error handling if not a custom error
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
		c.obs.LoggerService.Warn(ctx, "All retries failed, sending message to DLQ", map[string]interface{}{
			"topic":     c.topic,
			"messageID": messageID,
			"attempts":  attempt,
			"reason":    failureReason,
		})

		dlqErr := c.dlqProducer.SendToDLQ(ctx, messageID, msg.Value, msg.Headers, failureReason)
		if dlqErr != nil {
			c.obs.LoggerService.Error(ctx, "Failed to send message to DLQ", map[string]interface{}{
				"topic":     c.topic,
				"messageID": messageID,
				"error":     dlqErr.Error(),
			})
		}
	}

	// Commit the message after sending to DLQ if not auto-committing
	if !c.config.ConsumerConfig.AutoCommit {
		_, err = c.consumer.CommitMessage(msg)
		if err != nil {
			c.obs.LoggerService.Warn(ctx, "Failed to commit message offset after DLQ", map[string]interface{}{
				"topic":     c.topic,
				"messageID": messageID,
				"error":     err.Error(),
			})
		}
	}
}

// invokeHandler invokes the message handler with timeout
func (c *ConfluentConsumer) invokeHandler(ctx context.Context, payload []byte) error {
	functionName := "InvokeHandler"

	_, span := c.obs.TracerService.StartTracer(ctx, functionName)
	defer c.obs.TracerService.StopSpan(span)

	ctxWithTimeout, cancel := context.WithTimeout(ctx, c.retryTimeout)
	defer cancel()

	resultChan := make(chan error, 1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				c.obs.LoggerService.Error(ctx, "Handler panic recovered", map[string]interface{}{
					"topic": c.topic,
					"panic": fmt.Sprintf("%v", r),
				})
				resultChan <- errors.NewCustomError(
					errors.SUBErrSubscribeFailed,
					fmt.Errorf("panic: %v", r),
				)
			}
		}()
		resultChan <- c.handler(payload)
	}()

	select {
	case <-ctxWithTimeout.Done():
		c.obs.LoggerService.Warn(ctx, "Handler timeout", map[string]interface{}{
			"topic":   c.topic,
			"timeout": c.retryTimeout.String(),
		})
		return errors.NewCustomError(
			errors.SUBErrSubscribeFailed,
			fmt.Errorf("handler timeout: %w", ctxWithTimeout.Err()),
		)
	case err := <-resultChan:
		return err
	}
}

// Close closes the consumer and releases resources
func (c *ConfluentConsumer) Close() error {
	functionName := "Close"

	ctx := context.Background()
	_, span := c.obs.TracerService.StartTracer(ctx, functionName)
	defer c.obs.TracerService.StopSpan(span)

	c.obs.LoggerService.Info(ctx, "Closing consumer", map[string]interface{}{
		"topic":   c.topic,
		"groupID": c.groupID,
	})

	var err error
	if c.consumer != nil {
		err = c.consumer.Close()
		if err != nil {
			c.obs.LoggerService.Error(ctx, "Error closing consumer", map[string]interface{}{
				"topic":   c.topic,
				"groupID": c.groupID,
				"error":   err.Error(),
			})
		}
	}

	if c.dlqProducer != nil {
		dlqErr := c.dlqProducer.Close()
		if dlqErr != nil {
			c.obs.LoggerService.Error(ctx, "Error closing DLQ producer", map[string]interface{}{
				"topic": c.topic,
				"error": dlqErr.Error(),
			})
			if err == nil {
				err = dlqErr
			}
		}
	}

	c.obs.LoggerService.Info(ctx, "Consumer closed successfully", map[string]interface{}{
		"topic":   c.topic,
		"groupID": c.groupID,
	})

	return err
}
