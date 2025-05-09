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

// ConsumerError represents an error that can occur during message processing
type ConsumerError struct {
	Err       error
	Permanent bool
	Reason    string
}

func (e ConsumerError) Error() string {
	return fmt.Sprintf("%s: %v", e.Reason, e.Err)
}

// NewConsumerError creates a new consumer error
func NewConsumerError(err error, permanent bool, reason string) ConsumerError {
	return ConsumerError{
		Err:       err,
		Permanent: permanent,
		Reason:    reason,
	}
}

type Consumer struct {
	Reader       *kafka.Reader
	Handler      func([]byte) error
	RetryHandler *RetryHandler
	DLQProducer  *DLQProducer
	Topic        string
	MaxRetries   int
}

func NewConsumer(cfg KafkaConfig, handler func([]byte) error) *Consumer {
	// Check if this topic has a registered DLQ configuration
	dlqConfig, exists := GetDLQConfig(cfg.Topic)
	maxRetries := 5 // default

	if exists {
		maxRetries = dlqConfig.MaxRetries
	}

	return &Consumer{
		Reader:       NewKafkaReader(cfg),
		Handler:      handler,
		RetryHandler: NewRetryHandler(maxRetries, 500*time.Millisecond),
		DLQProducer:  NewDLQProducer(cfg),
		Topic:        cfg.Topic,
		MaxRetries:   maxRetries,
	}
}

func (c *Consumer) Start(ctx context.Context) {
	defer c.Reader.Close()
	defer c.DLQProducer.Close()

	log.Printf("[Kafka][%s] Consumer started", c.Topic)

	for {
		select {
		case <-ctx.Done():
			log.Printf("[Kafka][%s] Consumer stopped due to context cancellation", c.Topic)
			return
		default:
			log.Printf("[Kafka][%s] Waiting for message...", c.Topic)
			m, err := c.Reader.ReadMessage(ctx)
			if err != nil {
				if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
					log.Printf("[Kafka][%s] Context done: %v", c.Topic, err)
					return
				}

				log.Printf("[Kafka][%s] Read error: %v", c.Topic, err)
				time.Sleep(500 * time.Millisecond) // avoid tight loop on error
				continue
			}

			messageID := string(m.Key)
			if messageID == "" {
				messageID = fmt.Sprintf("%s-%d", c.Topic, m.Time.UnixNano())
			}

			c.processMessageWithRetries(ctx, messageID, m)
		}
	}
}

func (c *Consumer) processMessageWithRetries(ctx context.Context, messageID string, m kafka.Message) {
	var attempt int
	var failureReason string

	for {
		err := c.Handler(m.Value)

		if err == nil {
			// Successfully processed
			return
		}

		// Check if this is a Consumer Error with additional metadata
		var consumerErr ConsumerError
		if ce, ok := err.(ConsumerError); ok {
			consumerErr = ce // Use the variable to avoid linter error
			failureReason = consumerErr.Reason

			// If it's a permanent error, don't retry, send directly to DLQ
			if consumerErr.Permanent {
				log.Printf("[Kafka][%s] Permanent error, sending to DLQ: %v", c.Topic, consumerErr.Error())
				c.sendToDLQ(ctx, messageID, m, failureReason)
				return
			}
		} else {
			// Generic error
			failureReason = "Unknown error"
		}

		// Calculate retry delay based on attempt number
		delay := c.RetryHandler.HandleRetry(attempt)
		if delay < 0 {
			// Max retries exceeded, send to DLQ
			log.Printf("[Kafka][%s] Max retries (%d) exceeded, sending to DLQ: %s",
				c.Topic, c.MaxRetries, failureReason)
			c.sendToDLQ(ctx, messageID, m, failureReason)
			return
		}

		log.Printf("[Kafka][%s] Retrying attempt %d in %v: %s",
			c.Topic, attempt+1, delay, failureReason)
		time.Sleep(delay)
		attempt++
	}
}

func (c *Consumer) sendToDLQ(ctx context.Context, messageID string, m kafka.Message, reason string) {
	err := c.DLQProducer.SendToDLQ(ctx, messageID, m.Value, m.Headers, reason)
	if err != nil {
		log.Printf("[Kafka][%s] Failed to send to DLQ: %v", c.Topic, err)
		logger.LogErrorEvent("", "kafka_dlq_send_failed", "", "error",
			fmt.Sprintf("Failed to send message to DLQ: %v", err))
	}
}
