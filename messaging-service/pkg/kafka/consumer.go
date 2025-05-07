package kafka

import (
	"context"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
)

type Consumer struct {
	Reader       *kafka.Reader
	Handler      func([]byte) error
	RetryHandler *RetryHandler
	DLQProducer  *DLQProducer
	Topic        string
}

func NewConsumer(cfg KafkaConfig, handler func([]byte) error) *Consumer {
	return &Consumer{
		Reader:       NewKafkaReader(cfg),
		Handler:      handler,
		RetryHandler: NewRetryHandler(5, 500*time.Millisecond), // 5 retries, 500ms base delay
		DLQProducer:  NewDLQProducer(cfg),
		Topic:        cfg.Topic,
	}
}

func (c *Consumer) Start(ctx context.Context) {
	defer c.Reader.Close()
	defer c.DLQProducer.Close()

	for {
		m, err := c.Reader.ReadMessage(ctx)
		if err != nil {
			log.Printf("[Kafka][%s] Read error: %v", c.Topic, err)
			continue
		}

		var attempt int
		for {
			err = c.Handler(m.Value)
			if err == nil {
				break
			}

			delay := c.RetryHandler.HandleRetry(attempt)
			if delay < 0 {
				log.Printf("[Kafka][%s] Sending to DLQ after %d attempts: %s", c.Topic, attempt, string(m.Value))
				dlqErr := c.DLQProducer.SendToDLQ(ctx, string(m.Key), m.Value, m.Headers)
				if dlqErr != nil {
					log.Printf("[Kafka][%s] DLQ send failed: %v", c.Topic, dlqErr)
				}
				break
			}

			log.Printf("[Kafka][%s] Retrying attempt %d in %v", c.Topic, attempt+1, delay)
			time.Sleep(delay)
			attempt++
		}
	}
}
