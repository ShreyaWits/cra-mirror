package kafka

import (
	"context"
	"log"

	"github.com/segmentio/kafka-go"
)

type Producer struct {
	Writer *kafka.Writer
}

func NewProducer(cfg KafkaConfig) *Producer {
	return &Producer{
		Writer: &kafka.Writer{
			Addr:         kafka.TCP(cfg.Brokers...),
			Topic:        cfg.Topic,
			Balancer:     GetBalancer(cfg.BalancerType),
			RequiredAcks: kafka.RequireAll,
		},
	}
}

func (p *Producer) Write(ctx context.Context, key, value []byte) error {
	err := p.Writer.WriteMessages(ctx, kafka.Message{
		Key:   key,
		Value: value,
	})
	if err != nil {
		log.Printf("Kafka write error: %v", err)
	}
	return err
}

func (p *Producer) Close() error {
	return p.Writer.Close()
}
