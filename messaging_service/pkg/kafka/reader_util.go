package kafka

import (
	kafka "github.com/segmentio/kafka-go"
)

type BrokerMessage = kafka.Message
type Header = kafka.Header
type Writer = kafka.Writer

// NewKafkaReader creates a kafka reader with appropriate configuration based on delivery semantics
func NewKafkaReader(cfg KafkaConfig) *kafka.Reader {
	readerCfg := kafka.ReaderConfig{
		Brokers:         cfg.Brokers,
		Topic:           cfg.Topic,
		MinBytes:        cfg.MinBytes,
		MaxBytes:        cfg.MaxBytes,
		MaxWait:         cfg.ConsumerConfig.MaxWait,
		ReadBackoffMin:  cfg.ConsumerConfig.ReadBackoffMin,
		ReadBackoffMax:  cfg.ConsumerConfig.ReadBackoffMax,
		ReadLagInterval: -1,
	}

	if cfg.GroupID != "" {
		readerCfg.GroupID = cfg.GroupID

		// Configure delivery semantics
		switch cfg.DeliverySemantics {
		case ExactlyOnce:
			// Exactly-once requires additional tracking
			readerCfg.CommitInterval = cfg.ConsumerConfig.CommitInterval
			readerCfg.IsolationLevel = kafka.ReadCommitted
		default:
			// For at-least-once, commit after processing
			readerCfg.CommitInterval = cfg.ConsumerConfig.CommitInterval
			readerCfg.IsolationLevel = kafka.ReadUncommitted
		}

		// Set the starting offset based on the consumer mode
		switch cfg.Mode {
		case EarliestOffset:
			readerCfg.StartOffset = kafka.FirstOffset // Read all unprocessed messages
		case LatestOffset:
			readerCfg.StartOffset = kafka.LastOffset // Only read new messages after joining
		default:
			readerCfg.StartOffset = kafka.LastOffset // Default to latest offset
		}

		// Set consumer group specific configurations
		readerCfg.HeartbeatInterval = cfg.ConsumerConfig.HeartbeatInterval
		readerCfg.SessionTimeout = cfg.ConsumerConfig.SessionTimeout
		readerCfg.RebalanceTimeout = cfg.ConsumerConfig.RebalanceTimeout
		readerCfg.MaxAttempts = cfg.ConsumerConfig.MaxAttempts

		// Set isolation level based on configuration
		if cfg.ConsumerConfig.IsolationLevel == "read_committed" {
			readerCfg.IsolationLevel = kafka.ReadCommitted
		} else {
			readerCfg.IsolationLevel = kafka.ReadUncommitted
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
