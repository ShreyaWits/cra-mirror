package kafka

import (
	"time"

	kafka "github.com/segmentio/kafka-go"
)

type BrokerMessage = kafka.Message
type Header = kafka.Header

func NewKafkaReader(cfg KafkaConfig) *kafka.Reader {
	readerCfg := kafka.ReaderConfig{
		Brokers:         cfg.Brokers,
		Topic:           cfg.Topic,
		GroupID:         cfg.GroupID,
		MinBytes:        cfg.MinBytes,
		MaxBytes:        cfg.MaxBytes,
		StartOffset:     kafka.FirstOffset,
		MaxWait:         500 * time.Millisecond,
		ReadLagInterval: -1,
	}

	// Competing Consumers use GroupID
	if cfg.Mode == CompetingConsumer {
		readerCfg.GroupID = cfg.GroupID
	}

	return kafka.NewReader(readerCfg)
}
