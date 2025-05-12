package kafka

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
)

type KafkaAdmin interface {
	CreateTopic(ctx context.Context, topic string, numPartitions, replicationFactor int, configs map[string]string) error
	DeleteTopic(ctx context.Context, topic string) error
	ListTopics(ctx context.Context) ([]string, error)
	Close() error
}

type Admin struct {
	brokers []string
	dialer  *kafka.Dialer
	cfg     *KafkaConfig
}

// DefaultTopicConfig returns default configuration for a new topic
func DefaultTopicConfig() map[string]string {
	return map[string]string{
		"retention.ms":         fmt.Sprintf("%d", 7*24*time.Hour.Milliseconds()), // 7 days retention
		"cleanup.policy":       "delete",
		"compression.type":     "producer",
		"delete.retention.ms":  "86400000", // 1 day
		"file.delete.delay.ms": "60000",    // 1 minute
		"flush.messages":       "9223372036854775807",
		"flush.ms":             "9223372036854775807",
		"follower.replication.throttled.replicas": "",
		"leader.replication.throttled.replicas":   "",
		"message.downconversion.enable":           "true",
		"min.compaction.lag.ms":                   "0",
		"min.insync.replicas":                     "1",
		"segment.jitter.ms":                       "0",
		"segment.ms":                              "604800000",  // 7 days
		"segment.bytes":                           "1073741824", // 1GB
		"unclean.leader.election.enable":          "false",
		"message.timestamp.type":                  "CreateTime",
		"message.timestamp.difference.max.ms":     "9223372036854775807",
		"message.format.version":                  "2.8-IV1",
		"max.message.bytes":                       "1048588", // 1MB
	}
}

// NewAdmin creates a new Kafka admin client
func NewAdmin(cfg KafkaConfig) KafkaAdmin {
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

// CreateTopic creates a new Kafka topic with configuration
func (a *Admin) CreateTopic(ctx context.Context, topic string, numPartitions, replicationFactor int, configs map[string]string) error {
	if topic == "" {
		return errors.New("topic name cannot be empty")
	}

	// Use default values if not provided
	if numPartitions <= 0 {
		numPartitions = DefaultPartitions
	}
	if replicationFactor <= 0 {
		replicationFactor = DefaultReplicationFactor
	}

	// Merge default configs with provided configs
	defaultConfigs := DefaultTopicConfig()
	if configs == nil {
		configs = make(map[string]string)
	}
	for k, v := range defaultConfigs {
		if _, exists := configs[k]; !exists {
			configs[k] = v
		}
	}

	// Add delivery semantics specific configs
	if a.cfg != nil {
		switch a.cfg.DeliverySemantics {
		case ExactlyOnce:
			configs["transaction.max.timeout.ms"] = "900000" // 15 minutes
			configs["transaction.state.log.replication.factor"] = fmt.Sprintf("%d", replicationFactor)
			configs["transaction.state.log.min.isr"] = "2"
		}
	}

	configEntries := make([]kafka.ConfigEntry, 0, len(configs))
	for k, v := range configs {
		configEntries = append(configEntries, kafka.ConfigEntry{
			ConfigName:  k,
			ConfigValue: v,
		})
	}

	var lastErr error
	for _, broker := range a.brokers {
		conn, err := a.dialer.DialContext(ctx, "tcp", broker)
		if err != nil {
			fmt.Printf("Warning: failed to connect to broker %s: %v\n", broker, err)
			lastErr = err
			continue
		}
		defer conn.Close()

		err = conn.CreateTopics(kafka.TopicConfig{
			Topic:             topic,
			NumPartitions:     numPartitions,
			ReplicationFactor: replicationFactor,
			ConfigEntries:     configEntries,
		})
		if err != nil {
			fmt.Printf("Error: failed to create topic %s on broker %s: %v\n", topic, broker, err)
			lastErr = err
			continue
		}

		fmt.Printf("Success: created topic %s on broker %s with %d partitions and replication factor %d\n",
			topic, broker, numPartitions, replicationFactor)
		return nil
	}

	return fmt.Errorf("failed to create topic %q on all brokers: %w", topic, lastErr)
}

// DeleteTopic deletes a Kafka topic
func (a *Admin) DeleteTopic(ctx context.Context, topic string) error {
	if topic == "" {
		return errors.New("topic name cannot be empty")
	}

	var lastErr error
	for _, broker := range a.brokers {
		conn, err := a.dialer.DialContext(ctx, "tcp", broker)
		if err != nil {
			fmt.Printf("Warning: failed to connect to broker %s: %v\n", broker, err)
			lastErr = err
			continue
		}
		defer conn.Close()

		if err := conn.DeleteTopics(topic); err != nil {
			fmt.Printf("Error: failed to delete topic %s on broker %s: %v\n", topic, broker, err)
			lastErr = err
			continue
		}

		fmt.Printf("Success: deleted topic %s on broker %s\n", topic, broker)
		return nil
	}

	return fmt.Errorf("failed to delete topic %q on all brokers: %w", topic, lastErr)
}

// ListTopics returns a list of all Kafka topics
func (a *Admin) ListTopics(ctx context.Context) ([]string, error) {
	var lastErr error
	for _, broker := range a.brokers {
		conn, err := a.dialer.DialContext(ctx, "tcp", broker)
		if err != nil {
			fmt.Printf("Warning: failed to connect to broker %s: %v\n", broker, err)
			lastErr = err
			continue
		}
		defer conn.Close()

		partitions, err := conn.ReadPartitions()
		if err != nil {
			fmt.Printf("Error: failed to read partitions from broker %s: %v\n", broker, err)
			lastErr = err
			continue
		}

		topicMap := make(map[string]struct{})
		for _, p := range partitions {
			topicMap[p.Topic] = struct{}{}
		}

		topics := make([]string, 0, len(topicMap))
		for topic := range topicMap {
			topics = append(topics, topic)
		}

		return topics, nil
	}

	return nil, fmt.Errorf("failed to list topics: %w", lastErr)
}

// Close is a placeholder for future cleanup, currently does nothing
func (a *Admin) Close() error {
	return nil
}
