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
		MinBytes:        cfg.MinBytes,
		MaxBytes:        cfg.MaxBytes,
		MaxWait:         500 * time.Millisecond,
		ReadLagInterval: -1,
	}

	if cfg.GroupID != "" {
		readerCfg.GroupID = cfg.GroupID
		readerCfg.CommitInterval = time.Second
	} else {
		readerCfg.StartOffset = kafka.LastOffset
	}

	reader := kafka.NewReader(readerCfg)

	// Manually seek to the latest offset if there's no group ID (stateless consumer)
	if cfg.GroupID == "" {
		if err := reader.SetOffset(kafka.LastOffset); err != nil {
			// Optional: handle or log this error if needed
		}
	}

	return kafka.NewReader(readerCfg)
}
