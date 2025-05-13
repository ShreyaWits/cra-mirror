package kafka

import (
	"fmt"
	"messaging_service/pkg/logger"
)

// DLQConfigProvider interface defines methods for getting DLQ configuration
type DLQConfigProvider interface {
	GetDLQTopicName(sourceTopic string) string
	GetDLQConfig(topic string) (DLQTopicConfig, bool)
	LogDLQEvent(requestID, topic, messageID, reason, message string)
}

// DefaultDLQConfigProvider implements the DLQConfigProvider interface
type DefaultDLQConfigProvider struct{}

// NewDLQConfigProvider creates a new DLQ config provider
func NewDLQConfigProvider() DLQConfigProvider {
	return &DefaultDLQConfigProvider{}
}

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

// GetDLQTopicName returns the DLQ topic name for a given source topic (interface method)
func (d *DefaultDLQConfigProvider) GetDLQTopicName(sourceTopic string) string {
	return GetDLQTopicName(sourceTopic)
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

// GetDLQConfig returns DLQ configuration for a given topic (interface method)
func (d *DefaultDLQConfigProvider) GetDLQConfig(topic string) (DLQTopicConfig, bool) {
	return GetDLQConfig(topic)
}

// LogDLQEvent logs a message sent to DLQ
func LogDLQEvent(requestID, topic, messageID, reason, message string) {
	logger.LogEvent(requestID, "dlq_event", messageID, reason, message+" [DLQ topic: "+topic+"]")
}

// LogDLQEvent logs a message sent to DLQ (interface method)
func (d *DefaultDLQConfigProvider) LogDLQEvent(requestID, topic, messageID, reason, message string) {
	LogDLQEvent(requestID, topic, messageID, reason, message)
}
