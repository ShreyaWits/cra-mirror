package kafka

import (
	"context"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
)

// Admin represents a Kafka administration client
type Admin struct {
	brokers []string
	dialer  *kafka.Dialer
	cfg     *KafkaConfig
}

// NewAdmin creates a new Kafka admin client
func NewAdmin(cfg KafkaConfig) *Admin {
	dialer := &kafka.Dialer{
		Timeout:   10 * time.Second,
		DualStack: true,
	}

	return &Admin{
		brokers: cfg.Brokers,
		dialer:  dialer,
		cfg:     &cfg,
	}
}

// CreateTopic creates a new Kafka topic with the specified configuration
func (a *Admin) CreateTopic(ctx context.Context, topic string, numPartitions, replicationFactor int, configs map[string]string) error {
	// Validate topic name
	if topic == "" {
		return fmt.Errorf("topic name cannot be empty")
	}

	// Set default values if not provided
	if numPartitions <= 0 {
		numPartitions = 1
	}
	if replicationFactor <= 0 {
		replicationFactor = 1
	}

	// Convert configs to kafka.ConfigEntry format
	configEntries := make([]kafka.ConfigEntry, 0, len(configs))
	for k, v := range configs {
		configEntries = append(configEntries, kafka.ConfigEntry{
			ConfigName:  k,
			ConfigValue: v,
		})
	}
	var conn *kafka.Conn
	var err error
	// Create a connection to the kafka controller
	for _, broker := range a.brokers {
		conn, err = a.dialer.DialContext(ctx, "tcp", broker)
		if err == nil {
			break // Successful connection
		}
		// Log the error for the failed connection attempt
		fmt.Printf("Failed to connect to broker %s: %v\n", broker, err)
	}
	if err != nil {
		return fmt.Errorf("failed to connect to any kafka broker: %w", err)
	}
	defer conn.Close()

	// Prepare the topic configuration
	topicConfig := kafka.TopicConfig{
		Topic:             topic,
		NumPartitions:     numPartitions,
		ReplicationFactor: replicationFactor,
		ConfigEntries:     configEntries,
	}

	// Create the topic using controller connection
	err = conn.CreateTopics(topicConfig)
	if err != nil {
		return fmt.Errorf("failed to create topic: %w", err)
	}

	return nil
}

// DeleteTopic deletes a Kafka topic
func (a *Admin) DeleteTopic(ctx context.Context, topic string) error {
	conn, err := a.dialer.DialContext(ctx, "tcp", a.brokers[0])
	if err != nil {
		return fmt.Errorf("failed to connect to kafka: %w", err)
	}
	defer conn.Close()

	err = conn.DeleteTopics(topic)
	if err != nil {
		return fmt.Errorf("failed to delete topic: %w", err)
	}

	return nil
}

// ListTopics lists all available Kafka topics
func (a *Admin) ListTopics(ctx context.Context) ([]string, error) {
	conn, err := a.dialer.DialContext(ctx, "tcp", a.brokers[0])
	if err != nil {
		return nil, fmt.Errorf("failed to connect to kafka: %w", err)
	}
	defer conn.Close()

	partitions, err := conn.ReadPartitions()
	if err != nil {
		return nil, fmt.Errorf("failed to read partitions: %w", err)
	}

	// Deduplicate topics
	topicMap := make(map[string]struct{})
	for _, p := range partitions {
		topicMap[p.Topic] = struct{}{}
	}

	// Convert to slice
	topics := make([]string, 0, len(topicMap))
	for topic := range topicMap {
		topics = append(topics, topic)
	}

	return topics, nil
}

// Close closes the admin client
func (a *Admin) Close() error {
	// No explicit close needed for this implementation
	return nil
}
