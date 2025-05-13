package kafka

import (
	"fmt"
	"messaging_service/pkg/logger"

	kafka "github.com/segmentio/kafka-go"
)

type BrokerMessage = kafka.Message
type Header = kafka.Header
type Writer = kafka.Writer

// logKafkaEvents logs Kafka events for debugging
func logKafkaEvents(msg string, args ...interface{}) {
	logger.LogEvent("", "kafka_debug", "", "debug", fmt.Sprintf(msg, args...))
}

// logKafkaErrors logs Kafka errors
func logKafkaErrors(msg string, args ...interface{}) {
	logger.LogEvent("", "kafka_error", "", "error", fmt.Sprintf(msg, args...))
}

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
		Logger:          kafka.LoggerFunc(logKafkaEvents),
		ErrorLogger:     kafka.LoggerFunc(logKafkaErrors),
		MaxAttempts:     cfg.ConsumerConfig.MaxAttempts,
	}

	if cfg.GroupID != "" {
		readerCfg.GroupID = cfg.GroupID
		readerCfg.WatchPartitionChanges = true

		// Get isolation level from config
		if cfg.ConsumerConfig.IsolationLevel == "read_committed" {
			readerCfg.IsolationLevel = kafka.ReadCommitted
		} else {
			readerCfg.IsolationLevel = kafka.ReadUncommitted
		}

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

		// Prioritize explicit config setting over the Mode enum
		if cfg.ConsumerConfig.AutoOffsetReset == "earliest" {
			readerCfg.StartOffset = kafka.FirstOffset // Read all unprocessed messages
		} else if cfg.ConsumerConfig.AutoOffsetReset == "latest" {
			readerCfg.StartOffset = kafka.LastOffset // Only read new messages after joining
		} else {
			// Fall back to Mode enum if auto offset reset isn't explicitly set
			switch cfg.Mode {
			case EarliestOffset:
				readerCfg.StartOffset = kafka.FirstOffset // Read all unprocessed messages
			case LatestOffset:
				readerCfg.StartOffset = kafka.LastOffset // Only read new messages after joining
			default:
				readerCfg.StartOffset = kafka.LastOffset // Default to latest offset
			}
		}

		// Set consumer group specific configurations
		readerCfg.HeartbeatInterval = cfg.ConsumerConfig.HeartbeatInterval
		readerCfg.SessionTimeout = cfg.ConsumerConfig.SessionTimeout
		readerCfg.RebalanceTimeout = cfg.ConsumerConfig.RebalanceTimeout
		readerCfg.RetentionTime = cfg.ConsumerConfig.RetentionTime
		readerCfg.OffsetOutOfRangeError = true
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
			// Log this error
			logKafkaErrors("Failed to set offset to latest: %v", err)
		}
	}

	return reader
}
