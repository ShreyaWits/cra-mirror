package kafka

import (
	"time"

	kafka "github.com/segmentio/kafka-go"
)

type BrokerMessage = kafka.Message
type Header = kafka.Header

// NewKafkaReader creates a kafka reader with appropriate configuration based on delivery semantics
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

		// Configure delivery semantics
		switch cfg.DeliverySemantics {
		case ExactlyOnce:
			// Exactly-once requires additional tracking; set a reasonable commit interval
			readerCfg.CommitInterval = time.Second
			// Application needs to implement idempotent processing for exactly-once
		default:
			// For at-least-once, commit after processing to ensure we don't miss messages
			readerCfg.CommitInterval = time.Second * 5
			readerCfg.StartOffset = kafka.FirstOffset // Start from earliest unprocessed message
		}
	} else {
		// Stateless consumer
		if cfg.DeliverySemantics == AtLeastOnce {
			readerCfg.StartOffset = kafka.FirstOffset // Ensure we read all messages
		} else {
			readerCfg.StartOffset = kafka.LastOffset // Only read newest
		}
	}

	reader := kafka.NewReader(readerCfg)

	// Manually seek to the latest offset if there's no group ID (stateless consumer)
	if cfg.GroupID == "" && cfg.DeliverySemantics != AtLeastOnce {
		if err := reader.SetOffset(kafka.LastOffset); err != nil {
			// Optional: handle or log this error if needed
		}
	}

	return reader
}
