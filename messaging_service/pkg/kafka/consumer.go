package kafka

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"messaging_service/pkg/logger"

	"github.com/segmentio/kafka-go"
)

// ConsumerImplError represents an error that can occur during message processing
type ConsumerImplError struct {
	Err       error
	Permanent bool
	Reason    string
}

func (e *ConsumerImplError) Error() string {
	return fmt.Sprintf("%s: %v", e.Reason, e.Err)
}

// NewConsumerImplError creates a new ConsumerImpl error
func NewConsumerImplError(err error, permanent bool, reason string) *ConsumerImplError {
	return &ConsumerImplError{
		Err:       err,
		Permanent: permanent,
		Reason:    reason,
	}
}

// Consumer interface defines methods for consuming messages from Kafka
type Consumer interface {
	Start(ctx context.Context)
	processMessageWithRetries(ctx context.Context, messageID string, m kafka.Message)
	invokeHandlerWithTimeout(ctx context.Context, payload []byte) error
	sendToDLQ(ctx context.Context, messageID string, m kafka.Message, reason string)
}

// ConsumerImpl implements the Consumer interface
type ConsumerImpl struct {
	Reader       *kafka.Reader
	Handler      func([]byte) error
	RetryHandler *RetryHandler
	DLQProducer  DLQProducer
	Topic        string
	MaxRetries   int
	RetryTimeout time.Duration
}

// NewKafkaConsumer creates a new Consumer with default implementation
func NewKafkaConsumer(cfg KafkaConfig, handler func([]byte) error) Consumer {
	return NewConsumer(cfg, handler)
}

// NewConsumer creates a new Consumer with customizable components (useful for testing)
func NewConsumer(cfg KafkaConfig, handler func([]byte) error) Consumer {
	dlqConfig, exists := GetDLQConfig(cfg.Topic)
	maxRetries := 5
	if exists {
		maxRetries = dlqConfig.MaxRetries
	}

	return &ConsumerImpl{
		Reader:       NewKafkaReader(cfg),
		Handler:      handler,
		RetryHandler: NewRetryHandler(maxRetries, 500*time.Millisecond),
		DLQProducer:  NewDLQProducer(cfg),
		Topic:        cfg.Topic,
		MaxRetries:   maxRetries,
		RetryTimeout: 5 * time.Second,
	}
}

// NewConsumerWithCustomDLQ creates a consumer with a custom DLQ producer (useful for testing)
func NewConsumerWithCustomDLQ(reader *kafka.Reader, handler func([]byte) error,
	retryHandler *RetryHandler, dlqProducer DLQProducer, topic string, maxRetries int) Consumer {
	return &ConsumerImpl{
		Reader:       reader,
		Handler:      handler,
		RetryHandler: retryHandler,
		DLQProducer:  dlqProducer,
		Topic:        topic,
		MaxRetries:   maxRetries,
		RetryTimeout: 5 * time.Second,
	}
}

func (c *ConsumerImpl) Start(ctx context.Context) {
	defer func() {
		_ = c.Reader.Close()
		_ = c.DLQProducer.Close()
		log.Printf("[Kafka][%s] ConsumerImpl stopped", c.Topic)
	}()

	const (
		bufferSize  = 1000
		workerCount = 10
	)
	msgChan := make(chan kafka.Message, bufferSize)

	// Cancelable context
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Start worker pool
	var wg sync.WaitGroup
	wg.Add(workerCount)
	for i := 0; i < workerCount; i++ {
		go func(workerID int) {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case msg, ok := <-msgChan:
					if !ok {
						return
					}
					messageID := string(msg.Key)
					if messageID == "" {
						messageID = fmt.Sprintf("%s-%d", c.Topic, msg.Time.UnixNano())
					}
					c.processMessageWithRetries(ctx, messageID, msg)
				}
			}
		}(i)
	}

	// Main message reading loop
readLoop:
	for {
		select {
		case <-ctx.Done():
			break readLoop
		default:
			msg, err := c.Reader.ReadMessage(ctx)
			if err != nil {
				if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
					break readLoop
				}
				time.Sleep(50 * time.Millisecond)
				continue
			}

			// Optionally skip stale messages
			if time.Since(msg.Time) > 5*time.Second {
				log.Printf("[Kafka][%s] Skipping stale message at offset %d", c.Topic, msg.Offset)
				continue
			}

			select {
			case msgChan <- msg:
			default:
				// Drop to a separate goroutine if channel is full to avoid blocking
				go func(m kafka.Message) {
					messageID := string(m.Key)
					if messageID == "" {
						messageID = fmt.Sprintf("%s-%d", c.Topic, m.Time.UnixNano())
					}
					c.processMessageWithRetries(ctx, messageID, m)
				}(msg)
			}
		}
	}

	// Shutdown
	close(msgChan)
	wg.Wait()
}

func (c *ConsumerImpl) processMessageWithRetries(ctx context.Context, messageID string, m kafka.Message) {
	var attempt int
	var failureReason string

	for {
		handlerErr := c.invokeHandlerWithTimeout(ctx, m.Value)
		if handlerErr == nil {
			return
		}

		var consumerErr *ConsumerImplError
		if errors.As(handlerErr, &consumerErr) {
			failureReason = consumerErr.Reason
			if consumerErr.Permanent {
				c.sendToDLQ(ctx, messageID, m, failureReason)
				return
			}
		} else {
			failureReason = "unknown error"
		}

		delay := c.RetryHandler.HandleRetry(attempt)
		if delay < 0 {
			c.sendToDLQ(ctx, messageID, m, failureReason)
			return
		}

		time.Sleep(delay)
		attempt++
	}
}

func (c *ConsumerImpl) invokeHandlerWithTimeout(ctx context.Context, payload []byte) error {
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	resultChan := make(chan error, 1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				resultChan <- NewConsumerImplError(fmt.Errorf("panic: %v", r), true, "panic in handler")
			}
		}()
		resultChan <- c.Handler(payload)
	}()

	select {
	case <-ctxWithTimeout.Done():
		return NewConsumerImplError(ctxWithTimeout.Err(), false, "handler timeout")
	case err := <-resultChan:
		return err
	}
}

func (c *ConsumerImpl) sendToDLQ(ctx context.Context, messageID string, m kafka.Message, reason string) {
	err := c.DLQProducer.SendToDLQ(ctx, messageID, m.Value, m.Headers, reason)
	if err != nil {
		log.Printf("[Kafka][%s] Failed to send messageID=%s to DLQ at offset=%d: %v",
			c.Topic, messageID, m.Offset, err)
		logger.LogErrorEvent("", "kafka_dlq_send_failed", "", "error",
			fmt.Sprintf("Failed to send message to DLQ: %v", err))
	}
}
