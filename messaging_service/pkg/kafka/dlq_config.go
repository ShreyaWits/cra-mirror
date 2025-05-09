package kafka

import (
	"fmt"
	"messaging_service/pkg/logger"
)

// DLQTopicConfig holds configuration for a DLQ topic
type DLQTopicConfig struct {
	SourceTopic string
	DLQTopic    string
	MaxRetries  int
}

// GetDLQTopicName returns the DLQ topic name for a given source topic
func GetDLQTopicName(sourceTopic string) string {
	return fmt.Sprintf("%s.dlq", sourceTopic)
}

// RegisteredDLQTopics holds all registered DLQ topics
var RegisteredDLQTopics = map[string]DLQTopicConfig{
	"gpc": {
		SourceTopic: "gpc",
		DLQTopic:    "gpc.dlq",
		MaxRetries:  5,
	},
	// Add more topics as needed
}

// GetDLQConfig returns DLQ configuration for a given topic
func GetDLQConfig(topic string) (DLQTopicConfig, bool) {
	config, exists := RegisteredDLQTopics[topic]
	return config, exists
}

// LogDLQEvent logs a message sent to DLQ
func LogDLQEvent(requestID, topic, messageID, reason, message string) {
	logger.LogDLQEvent(requestID, topic, messageID, reason, message)
}
