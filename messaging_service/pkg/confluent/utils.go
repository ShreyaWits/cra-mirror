package confluent

import (
	"fmt"
	"messaging_service/pkg/logger"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

// KafkaMonitorDeliveryReports handles delivery reports from the producer
// This is a common utility used by both Producer and DLQProducer
func KafkaMonitorDeliveryReports(producer KafkaProducerInterface, topic string) {
	for e := range producer.Events() {
		switch ev := e.(type) {
		case *kafka.Message:
			if ev.TopicPartition.Error != nil {
				logger.LogErrorEvent("", "kafka_delivery_failed", topic, "error",
					fmt.Sprintf("Message delivery failed: %v", ev.TopicPartition.Error))
			} else {
				logger.LogEvent("", "kafka_delivery_success", topic, "info",
					fmt.Sprintf("Message delivered to topic %s [%d] at offset %d",
						*ev.TopicPartition.Topic, ev.TopicPartition.Partition, ev.TopicPartition.Offset))
			}
		}
	}
}

// LogKafkaError logs a Kafka error with appropriate metadata
func LogKafkaError(eventType, topic, message string, err error) {
	logger.LogErrorEvent("", eventType, topic, "error", fmt.Sprintf("%s: %v", message, err))
}

// LogKafkaInfo logs a Kafka informational message with appropriate metadata
func LogKafkaInfo(eventType, topic, message string) {
	logger.LogEvent("", eventType, topic, "info", message)
}

// LogKafkaWarning logs a Kafka warning with appropriate metadata
func LogKafkaWarning(eventType, topic, message string) {
	logger.LogWarnEvent("", eventType, topic, "warn", message)
}
