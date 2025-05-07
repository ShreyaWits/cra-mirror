package kafka

import (
	"context"

	"github.com/segmentio/kafka-go"
)

type DLQProducer struct {
	writer *kafka.Writer
}

func NewDLQProducer(cfg KafkaConfig) *DLQProducer {
	writer := &kafka.Writer{
		Addr:     kafka.TCP(cfg.Brokers...),
		Topic:    cfg.Topic,
		Balancer: GetBalancer(cfg.BalancerType),
	}
	return &DLQProducer{writer: writer}
}

func (d *DLQProducer) SendToDLQ(ctx context.Context, key string, value []byte, headers []kafka.Header) error {
	return d.writer.WriteMessages(ctx, kafka.Message{
		Key:     []byte(key),
		Value:   value,
		Headers: headers,
	})
}

func (d *DLQProducer) Close() error {
	return d.writer.Close()
}
