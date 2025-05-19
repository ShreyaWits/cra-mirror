package confluent

import (
	"context"
	"messaging_service/pkg/observability"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

// KafkaMonitorDeliveryReports handles delivery reports from the producer
// This is a common utility used by both Producer and DLQProducer
func KafkaMonitorDeliveryReports(producer KafkaProducerInterface, topic string, obs *observability.ObservabilityStack) {
	functionName := "KafkaMonitorDeliveryReports"
	ctx := context.Background()

	_, span := obs.TracerService.StartTracer(ctx, functionName)
	defer obs.TracerService.StopSpan(span)

	obs.LoggerService.Info(ctx, "Starting delivery report monitoring", map[string]interface{}{
		"topic": topic,
	})

	for e := range producer.Events() {
		switch ev := e.(type) {
		case *kafka.Message:
			if ev.TopicPartition.Error != nil {
				obs.LoggerService.Error(ctx, "Message delivery failed", map[string]interface{}{
					"topic":     topic,
					"error":     ev.TopicPartition.Error.Error(),
					"partition": ev.TopicPartition.Partition,
				})
			} else {
				obs.LoggerService.Info(ctx, "Message delivered successfully", map[string]interface{}{
					"topic":     *ev.TopicPartition.Topic,
					"partition": ev.TopicPartition.Partition,
					"offset":    ev.TopicPartition.Offset,
				})
			}
		}
	}

	obs.LoggerService.Info(ctx, "Delivery report monitoring stopped", map[string]interface{}{
		"topic": topic,
	})
}

// LogKafkaError logs a Kafka error with appropriate metadata
func LogKafkaError(ctx context.Context, eventType, topic, message string, err error, obs *observability.ObservabilityStack) {
	obs.LoggerService.Error(ctx, message, map[string]interface{}{
		"eventType": eventType,
		"topic":     topic,
		"error":     err.Error(),
	})
}

// LogKafkaInfo logs a Kafka informational message with appropriate metadata
func LogKafkaInfo(ctx context.Context, eventType, topic, message string, obs *observability.ObservabilityStack) {
	obs.LoggerService.Info(ctx, message, map[string]interface{}{
		"eventType": eventType,
		"topic":     topic,
	})
}

// LogKafkaWarning logs a Kafka warning with appropriate metadata
func LogKafkaWarning(ctx context.Context, eventType, topic, message string, obs *observability.ObservabilityStack) {
	obs.LoggerService.Warn(ctx, message, map[string]interface{}{
		"eventType": eventType,
		"topic":     topic,
	})
}
