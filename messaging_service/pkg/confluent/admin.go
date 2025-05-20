package confluent

import (
	"context"
	"fmt"
	"time"

	"messaging_service/internal/config"
	"messaging_service/pkg/observability"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

// KafkaAdmin interface defines admin operations for Kafka

// AdminImpl implements the KafkaAdmin interface
type AdminImpl struct {
	adminClient *AdminClient
	config      *config.Config
	obs         *observability.ObservabilityStack
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
func NewAdmin(cfg *config.Config, obs *observability.ObservabilityStack) (KafkaAdmin, error) {
	functionName := "NewAdmin"
	ctx := context.Background()

	_, span := obs.TracerService.StartTracer(ctx, functionName)
	defer obs.TracerService.StopSpan(span)

	// Create admin client configuration
	adminConfig := &ConfigMap{
		"bootstrap.servers": buildBrokerString(cfg.KafkaBrokers),
	}

	// Create admin client
	adminClient, err := kafka.NewAdminClient(adminConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka admin client: %w", err)
	}

	return &AdminImpl{
		adminClient: adminClient,
		config:      cfg,
		obs:         obs,
	}, nil
}

// CreateTopic creates a new Kafka topic
func (a *AdminImpl) CreateTopic(ctx context.Context, topic string, numPartitions, replicationFactor int, configs map[string]string) error {
	functionName := "CreateTopic"

	tCtx, span := a.obs.TracerService.StartTracer(ctx, functionName)
	defer a.obs.TracerService.StopSpan(span)

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
	result, err := a.adminClient.CreateTopics(tCtx, []TopicSpecification{topicSpec})
	if err != nil {
		return fmt.Errorf("failed to create Kafka topic: %w", err)
	}

	// Check for per-topic errors
	if len(result) > 0 && result[0].Error.Code() != ErrNoError {
		return fmt.Errorf("failed to create Kafka topic: %v", result[0].Error)
	}

	return nil
}

// DeleteTopic deletes a Kafka topic
func (a *AdminImpl) DeleteTopic(ctx context.Context, topic string) error {
	functionName := "DeleteTopic"

	tCtx, span := a.obs.TracerService.StartTracer(ctx, functionName)
	defer a.obs.TracerService.StopSpan(span)

	// Delete the topic
	result, err := a.adminClient.DeleteTopics(tCtx, []string{topic})
	if err != nil {
		return fmt.Errorf("failed to delete Kafka topic: %w", err)
	}

	// Check for per-topic errors
	if len(result) > 0 && result[0].Error.Code() != ErrNoError {
		return fmt.Errorf("failed to delete Kafka topic: %v", result[0].Error)
	}

	return nil
}

// ListTopics lists all topics in the Kafka cluster
func (a *AdminImpl) ListTopics(ctx context.Context) ([]string, error) {
	functionName := "ListTopics"

	_, span := a.obs.TracerService.StartTracer(ctx, functionName)
	defer a.obs.TracerService.StopSpan(span)

	// Create metadata object
	metadata, err := a.adminClient.GetMetadata(nil, true, 30000)
	if err != nil {
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
	functionName := "HasActiveConsumers"

	tCtx, span := a.obs.TracerService.StartTracer(ctx, functionName)
	defer a.obs.TracerService.StopSpan(span)

	a.obs.LoggerService.Info(tCtx, fmt.Sprintf("Checking for active consumers on topic: %s", topic))

	// First check if topic exists
	metadata, err := a.adminClient.GetMetadata(&topic, false, 10000)
	if err != nil {
		a.obs.LoggerService.Error(tCtx, fmt.Sprintf("Failed to check if topic exists: %v", err))
		return false, fmt.Errorf("failed to check if topic exists: %w", err)
	}

	if _, exists := metadata.Topics[topic]; !exists {
		a.obs.LoggerService.Info(tCtx, fmt.Sprintf("Topic %s does not exist", topic))
		return false, nil
	}
	a.obs.LoggerService.Info(tCtx, fmt.Sprintf("Topic %s exists", topic))

	// Get the admin client metadata which includes information about the cluster
	clusterMetadata, err := a.adminClient.GetMetadata(nil, true, 30000)
	if err != nil {
		a.obs.LoggerService.Warn(tCtx, fmt.Sprintf("Failed to get cluster metadata: %v. Assuming consumers exist for safety", err))
		// Fall back to assume consumers exist (safer)
		return true, nil
	}

	// Check if the broker is healthy
	if len(clusterMetadata.Brokers) == 0 {
		a.obs.LoggerService.Warn(tCtx, "No brokers found in cluster metadata. Assuming consumers exist for safety")
		// Assume consumers in case of broker connectivity issues
		return true, nil
	}
	a.obs.LoggerService.Info(tCtx, fmt.Sprintf("Found %d brokers in cluster", len(clusterMetadata.Brokers)))

	// Make a simple check for any consumer group activity in the cluster
	consumerGroupsList, err := a.adminClient.ListConsumerGroups(tCtx)
	if err != nil {
		a.obs.LoggerService.Warn(tCtx, fmt.Sprintf("Failed to list consumer groups: %v. Assuming consumers exist for safety", err))
		// Fall back to assuming consumers exist
		return true, nil
	}

	// If there are no consumer groups at all, the topic can't have active consumers
	if len(consumerGroupsList.Valid) == 0 {
		a.obs.LoggerService.Info(tCtx, "No consumer groups found in cluster")
		return false, nil
	}
	a.obs.LoggerService.Info(tCtx, fmt.Sprintf("Found %d consumer groups in cluster", len(consumerGroupsList.Valid)))

	// We know there are consumer groups in the cluster
	// For now, if topic exists and there are consumer groups, we'll assume it has consumers
	// This is a safer approach until we can implement a more detailed check
	// that works reliably with the Confluent Kafka API
	a.obs.LoggerService.Info(tCtx, fmt.Sprintf("Topic %s exists and consumer groups are present. Assuming active consumers", topic))
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
