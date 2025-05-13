package confluent

import (
	"context"
	"fmt"
	"time"

	"messaging_service/pkg/logger"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

// KafkaAdmin interface defines admin operations for Kafka
type KafkaAdmin interface {
	// CreateTopic creates a new topic with the given configuration
	CreateTopic(ctx context.Context, topic string, numPartitions, replicationFactor int, configs map[string]string) error

	// DeleteTopic deletes an existing topic
	DeleteTopic(ctx context.Context, topic string) error

	// ListTopics lists all available topics
	ListTopics(ctx context.Context) ([]string, error)

	// HasActiveConsumers checks if a topic has any active consumer groups
	HasActiveConsumers(ctx context.Context, topic string) (bool, error)

	// Close releases any resources held by the admin client
	Close() error
}

// AdminImpl implements the KafkaAdmin interface
type AdminImpl struct {
	adminClient *AdminClient
	config      *KafkaConfig
}

// DefaultTopicConfig returns a default configuration for a new topic
func DefaultTopicConfig() map[string]string {
	return map[string]string{
		"retention.ms":                   fmt.Sprintf("%d", 7*24*time.Hour.Milliseconds()), // 7 days
		"cleanup.policy":                 "delete",
		"compression.type":               "producer",
		"min.insync.replicas":            "1",
		"unclean.leader.election.enable": "false",
		"max.message.bytes":              "1048588", // 1MB
	}
}

// NewAdmin creates a new Kafka admin client
func NewAdmin(cfg KafkaConfig) (KafkaAdmin, error) {
	// Create admin client configuration
	adminConfig := &ConfigMap{
		"bootstrap.servers": buildBrokerString(cfg.Brokers),
	}

	// Create admin client
	adminClient, err := kafka.NewAdminClient(adminConfig)
	if err != nil {
		logger.LogErrorEvent("", "kafka_admin_creation", "", "error",
			fmt.Sprintf("Failed to create Kafka admin client: %v", err))
		return nil, fmt.Errorf("failed to create Kafka admin client: %w", err)
	}

	return &AdminImpl{
		adminClient: adminClient,
		config:      &cfg,
	}, nil
}

// CreateTopic creates a new Kafka topic
func (a *AdminImpl) CreateTopic(ctx context.Context, topic string, numPartitions, replicationFactor int, configs map[string]string) error {
	// Set default values if necessary
	if numPartitions <= 0 {
		numPartitions = 3
	}
	if replicationFactor <= 0 {
		replicationFactor = 3
	}
	if configs == nil {
		configs = DefaultTopicConfig()
	}

	// Create topic configs
	configEntries := make([]ConfigEntry, 0, len(configs))
	for k, v := range configs {
		configEntries = append(configEntries, ConfigEntry{
			Name:  k,
			Value: v,
		})
	}

	// Create topic specification
	topicSpec := TopicSpecification{
		Topic:             topic,
		NumPartitions:     numPartitions,
		ReplicationFactor: replicationFactor,
		Config:            configsToStringMap(configEntries),
	}

	// Create the topic
	result, err := a.adminClient.CreateTopics(ctx, []TopicSpecification{topicSpec})
	if err != nil {
		logger.LogErrorEvent("", "kafka_create_topic_failed", topic, "error",
			fmt.Sprintf("Failed to create Kafka topic: %v", err))
		return fmt.Errorf("failed to create Kafka topic: %w", err)
	}

	// Check for per-topic errors
	if len(result) > 0 && result[0].Error.Code() != ErrNoError {
		logger.LogErrorEvent("", "kafka_create_topic_failed", topic, "error",
			fmt.Sprintf("Failed to create Kafka topic: %v", result[0].Error))
		return fmt.Errorf("failed to create Kafka topic: %v", result[0].Error)
	}

	logger.LogEvent("", "kafka_create_topic_success", topic, "info",
		fmt.Sprintf("Created Kafka topic %s with %d partitions and replication factor %d",
			topic, numPartitions, replicationFactor))
	return nil
}

// DeleteTopic deletes a Kafka topic
func (a *AdminImpl) DeleteTopic(ctx context.Context, topic string) error {
	// Delete the topic
	result, err := a.adminClient.DeleteTopics(ctx, []string{topic})
	if err != nil {
		logger.LogErrorEvent("", "kafka_delete_topic_failed", topic, "error",
			fmt.Sprintf("Failed to delete Kafka topic: %v", err))
		return fmt.Errorf("failed to delete Kafka topic: %w", err)
	}

	// Check for per-topic errors
	if len(result) > 0 && result[0].Error.Code() != ErrNoError {
		logger.LogErrorEvent("", "kafka_delete_topic_failed", topic, "error",
			fmt.Sprintf("Failed to delete Kafka topic: %v", result[0].Error))
		return fmt.Errorf("failed to delete Kafka topic: %v", result[0].Error)
	}

	logger.LogEvent("", "kafka_delete_topic_success", topic, "info",
		fmt.Sprintf("Deleted Kafka topic %s", topic))
	return nil
}

// ListTopics lists all topics in the Kafka cluster
func (a *AdminImpl) ListTopics(ctx context.Context) ([]string, error) {
	// Create metadata object
	metadata, err := a.adminClient.GetMetadata(nil, true, 30000)
	if err != nil {
		logger.LogErrorEvent("", "kafka_list_topics_failed", "", "error",
			fmt.Sprintf("Failed to list Kafka topics: %v", err))
		return nil, fmt.Errorf("failed to list Kafka topics: %w", err)
	}

	// Extract topic names
	topics := make([]string, 0, len(metadata.Topics))
	for topic := range metadata.Topics {
		topics = append(topics, topic)
	}

	return topics, nil
}

// HasActiveConsumers checks if a topic has any active consumer groups
func (a *AdminImpl) HasActiveConsumers(ctx context.Context, topic string) (bool, error) {
	// First check if topic exists
	metadata, err := a.adminClient.GetMetadata(&topic, false, 10000)
	if err != nil {
		logger.LogErrorEvent("", "kafka_check_topic_failed", topic, "error",
			fmt.Sprintf("Failed to check if topic exists: %v", err))
		return false, fmt.Errorf("failed to check if topic exists: %w", err)
	}

	if _, exists := metadata.Topics[topic]; !exists {
		// Topic doesn't exist
		logger.LogEvent("", "kafka_topic_not_found", topic, "info",
			fmt.Sprintf("Topic %s does not exist", topic))
		return false, nil
	}

	// Get the admin client metadata which includes information about the cluster
	clusterMetadata, err := a.adminClient.GetMetadata(nil, true, 30000)
	if err != nil {
		logger.LogErrorEvent("", "kafka_get_metadata_failed", topic, "error",
			fmt.Sprintf("Failed to get cluster metadata: %v", err))
		// Fall back to assume consumers exist (safer)
		return true, nil
	}

	// Check if the broker is healthy
	if len(clusterMetadata.Brokers) == 0 {
		logger.LogErrorEvent("", "kafka_no_brokers", topic, "error",
			"No Kafka brokers available")
		// Assume consumers in case of broker connectivity issues
		return true, nil
	}

	// Make a simple check for any consumer group activity in the cluster
	consumerGroupsList, err := a.adminClient.ListConsumerGroups(ctx)
	if err != nil {
		logger.LogErrorEvent("", "kafka_list_consumer_groups_failed", topic, "error",
			fmt.Sprintf("Failed to list consumer groups: %v", err))
		// Fall back to assuming consumers exist
		return true, nil
	}

	// If there are no consumer groups at all, the topic can't have active consumers
	if len(consumerGroupsList.Valid) == 0 {
		logger.LogEvent("", "kafka_no_consumer_groups", topic, "info",
			fmt.Sprintf("No consumer groups found in the cluster, topic %s has no active consumers", topic))
		return false, nil
	}

	// We know there are consumer groups in the cluster
	// For now, if topic exists and there are consumer groups, we'll assume it has consumers
	// This is a safer approach until we can implement a more detailed check
	// that works reliably with the Confluent Kafka API

	logger.LogEvent("", "kafka_assume_active_consumer", topic, "info",
		fmt.Sprintf("Topic %s exists and consumer groups are active in cluster, assuming topic has consumers", topic))

	// If we want to enforce stricter validation, return false here instead.
	// For now, we're using a safer approach by returning true.
	return true, nil
}

// Close closes the admin client
func (a *AdminImpl) Close() error {
	a.adminClient.Close()
	return nil
}

// configsToStringMap converts a slice of ConfigEntry to a string map
func configsToStringMap(entries []ConfigEntry) map[string]string {
	configs := make(map[string]string)
	for _, entry := range entries {
		configs[entry.Name] = entry.Value
	}
	return configs
}
