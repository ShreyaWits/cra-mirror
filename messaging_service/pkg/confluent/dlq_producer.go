package confluent

import (
	"context"
	"fmt"
	"time"

	"messaging_service/pkg/observability"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

// DLQConfig holds configuration for a dead-letter queue
type DLQConfig struct {
	// MaxRetries before sending to DLQ
	MaxRetries int

	// Suffix to append to topic name to create DLQ topic name
	TopicSuffix string
}

// DefaultDLQConfig returns default configuration for a DLQ
func DefaultDLQConfig() DLQConfig {
	return DLQConfig{
		MaxRetries:  3,
		TopicSuffix: "-dlq",
	}
}

// DLQProducerImpl implements the DLQProducer interface
type DLQProducerImpl struct {
	Producer KafkaProducerInterface
	Topic    string
	Config   KafkaConfig
	obs      *observability.ObservabilityStack
}

// NewDLQProducer creates a new DLQ producer
func NewDLQProducer(cfg KafkaConfig, obs *observability.ObservabilityStack) (DLQProducer, error) {
	functionName := "NewDLQProducer"
	ctx := context.Background()

	_, span := obs.TracerService.StartTracer(ctx, functionName)
	defer obs.TracerService.StopSpan(span)

	// Create DLQ topic name
	dlqTopic := GetDLQTopicName(cfg.Topic)

	// Create a copy of the config with the DLQ topic
	dlqConfig := cfg
	dlqConfig.Topic = dlqTopic

	// Get producer configuration
	producerConfig := dlqConfig.ToConfluentProducerConfig()

	// Create the producer
	producer, err := kafka.NewProducer(producerConfig)
	if err != nil {
		obs.LoggerService.Error(ctx, "Failed to create DLQ producer", map[string]interface{}{
			"topic": dlqTopic,
			"error": err.Error(),
		})
		return nil, fmt.Errorf("failed to create DLQ producer: %w", err)
	}

	// Start monitoring delivery reports
	go KafkaMonitorDeliveryReports(producer, dlqTopic, obs)

	obs.LoggerService.Info(ctx, "DLQ producer created successfully", map[string]interface{}{
		"topic": dlqTopic,
	})

	return &DLQProducerImpl{
		Producer: producer,
		Topic:    dlqTopic,
		Config:   dlqConfig,
		obs:      obs,
	}, nil
}

// SendToDLQ sends a failed message to the DLQ
func (d *DLQProducerImpl) SendToDLQ(ctx context.Context, messageID string, value []byte, headers []Header, failureReason string) error {
	functionName := "SendToDLQ"

	_, span := d.obs.TracerService.StartTracer(ctx, functionName)
	defer d.obs.TracerService.StopSpan(span)

	// Add DLQ-specific headers
	dlqHeaders := append(headers, []Header{
		{
			Key:   "x-dlq-source-topic",
			Value: []byte(d.Config.Topic),
		},
		{
			Key:   "x-dlq-failure-reason",
			Value: []byte(failureReason),
		},
		{
			Key:   "x-dlq-timestamp",
			Value: []byte(time.Now().Format(time.RFC3339)),
		},
	}...)

	// Log the DLQ event
	d.obs.LoggerService.Info(ctx, "Sending message to DLQ", map[string]interface{}{
		"topic":         d.Topic,
		"messageID":     messageID,
		"failureReason": failureReason,
		"sourceTopic":   d.Config.Topic,
		"headerCount":   len(dlqHeaders),
	})

	// Create Kafka message
	msg := &Message{
		TopicPartition: TopicPartition{Topic: &d.Topic, Partition: PartitionAny},
		Key:            []byte(messageID),
		Value:          value,
		Headers:        dlqHeaders,
		Timestamp:      time.Now(),
	}

	// Produce the message
	err := d.Producer.Produce(msg, nil)
	if err != nil {
		d.obs.LoggerService.Error(ctx, "Failed to send message to DLQ", map[string]interface{}{
			"topic":     d.Topic,
			"messageID": messageID,
			"error":     err.Error(),
		})
		return fmt.Errorf("failed to send message to DLQ: %w", err)
	}

	// Flush to ensure delivery
	remaining := d.Producer.Flush(int(d.Config.WriteTimeout.Milliseconds()))
	if remaining > 0 {
		d.obs.LoggerService.Warn(ctx,"Messages still in queue after flush timeout", map[string]interface{}{
			"topic":     d.Topic,
			"remaining": remaining,
			"timeout":   d.Config.WriteTimeout.String(),
		})
	}

	d.obs.LoggerService.Info(ctx, "Message successfully sent to DLQ", map[string]interface{}{
		"topic":     d.Topic,
		"messageID": messageID,
	})

	return nil
}

// Close closes the DLQ producer
func (d *DLQProducerImpl) Close() error {
	functionName := "Close"

	ctx := context.Background()
	_, span := d.obs.TracerService.StartTracer(ctx, functionName)
	defer d.obs.TracerService.StopSpan(span)

	d.obs.LoggerService.Info(ctx, "Closing DLQ producer", map[string]interface{}{
		"topic": d.Topic,
	})

	// Flush any pending messages
	remaining := d.Producer.Flush(int(d.Config.WriteTimeout.Milliseconds()))
	if remaining > 0 {
		d.obs.LoggerService.Warn(ctx,"Messages still in queue after close flush timeout", map[string]interface{}{
			"topic":     d.Topic,
			"remaining": remaining,
			"timeout":   d.Config.WriteTimeout.String(),
		})
	}

	// Close the producer
	d.Producer.Close()

	d.obs.LoggerService.Info(ctx, "DLQ producer closed successfully", map[string]interface{}{
		"topic": d.Topic,
	})

	return nil
}

// GetDLQTopicName returns the DLQ topic name for a given topic
func GetDLQTopicName(topic string) string {
	return topic + "-dlq"
}
