package kafka

import (
	"context"
	"errors"
	"fmt"
	"log"
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

type Consumer interface {
	Start(ctx context.Context)
}

type ConsumerImpl struct {
	Reader       *kafka.Reader
	Handler      func([]byte) error
	RetryHandler *RetryHandler
	DLQProducer  *DLQProducer
	Topic        string
	MaxRetries   int
	RetryTimeout time.Duration
}

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

func (c *ConsumerImpl) Start(ctx context.Context) {
	defer c.Reader.Close()
	defer c.DLQProducer.Close()

	log.Printf("[Kafka][%s] ConsumerImpl started", c.Topic)

	// Create a buffered channel for message processing
	msgChan := make(chan kafka.Message, 100)

	// Start message processor goroutine
	go func() {
		for msg := range msgChan {
			messageID := string(msg.Key)
			if messageID == "" {
				messageID = fmt.Sprintf("%s-%d", c.Topic, msg.Time.UnixNano())
			}
			c.processMessageWithRetries(ctx, messageID, msg)
		}
	}()

	for {
		select {
		case <-ctx.Done():
			close(msgChan)
			log.Printf("[Kafka][%s] ConsumerImpl stopped due to context cancellation", c.Topic)
			return
		default:
			m, err := c.Reader.ReadMessage(ctx)
			if err != nil {
				if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
					close(msgChan)
					return
				}
				time.Sleep(100 * time.Millisecond)
				continue
			}

			// Check message timestamp
			if time.Since(m.Time) > 3*time.Second {
				continue
			}

			select {
			case msgChan <- m:
			default:
				// If channel is full, process message in current goroutine
				messageID := string(m.Key)
				if messageID == "" {
					messageID = fmt.Sprintf("%s-%d", c.Topic, m.Time.UnixNano())
				}
				c.processMessageWithRetries(ctx, messageID, m)
			}
		}
	}
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
