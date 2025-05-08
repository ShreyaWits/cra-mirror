package kafka

import (
	"context"

	"github.com/segmentio/kafka-go"
)

type KafkaReader interface {
	ReadMessage(ctx context.Context) (kafka.Message, error)
	Close() error
}

type KafkaWriter interface {
	WriteMessages(ctx context.Context, msgs ...kafka.Message) error
	Close() error
}

type KafkaConsumerService interface {
	Consume(topic string, handler func(msg kafka.Message))
	Start(ctx context.Context)
	Close()
}

type KafkaProducerService interface {
	Produce(ctx context.Context, value string) error
	Close()
}
