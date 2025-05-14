package confluent_test

import (
	"messaging_service/pkg/confluent"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestKafkaConfig_ToConfluentProducerConfig(t *testing.T) {
	tests := []struct {
		name       string
		input      confluent.KafkaConfig
		wantConfig map[string]interface{}
	}{
		{
			name: "exactly-once producer config",
			input: confluent.KafkaConfig{
				Brokers:           []string{"broker1:9092", "broker2:9092"},
				Topic:             "topic-A",
				BatchSize:         1024,
				BatchBytes:        2048,
				BatchTimeout:      300 * time.Millisecond,
				CompressionType:   "gzip",
				ReadTimeout:       4 * time.Second,
				WriteTimeout:      3 * time.Second,
				DeliverySemantics: confluent.ExactlyOnce,
				ExactlyOnceConfig: confluent.ExactlyOnceConfig{
					EnableIdempotence:     true,
					EnableTransactions:    true,
					TransactionalIDPrefix: "prefix",
					TransactionTimeoutMs:  60000,
				},
			},
			wantConfig: map[string]interface{}{
				"bootstrap.servers":      "broker1:9092,broker2:9092",
				"retry.backoff.ms":       0, // default zero unless explicitly set
				"message.max.bytes":      2048,
				"socket.timeout.ms":      4000,
				"message.timeout.ms":     3000,
				"compression.type":       "gzip",
				"batch.size":             1024,
				"linger.ms":              300,
				"enable.idempotence":     true,
				"acks":                   "all",
				"transactional.id":       "prefix-topic-A",
				"transaction.timeout.ms": 60000,
			},
		},
		{
			name: "exactly-once producer config",
			input: confluent.KafkaConfig{
				Brokers:           []string{},
				Topic:             "topic-A",
				BatchSize:         1024,
				BatchBytes:        2048,
				BatchTimeout:      300 * time.Millisecond,
				CompressionType:   "gzip",
				ReadTimeout:       4 * time.Second,
				WriteTimeout:      3 * time.Second,
				DeliverySemantics: confluent.ExactlyOnce,
				ExactlyOnceConfig: confluent.ExactlyOnceConfig{
					EnableIdempotence:     true,
					EnableTransactions:    true,
					TransactionalIDPrefix: "prefix",
					TransactionTimeoutMs:  60000,
				},
			},
			wantConfig: map[string]interface{}{
				"bootstrap.servers":      "localhost:9092",
				"retry.backoff.ms":       0, // default zero unless explicitly set
				"message.max.bytes":      2048,
				"socket.timeout.ms":      4000,
				"message.timeout.ms":     3000,
				"compression.type":       "gzip",
				"batch.size":             1024,
				"linger.ms":              300,
				"enable.idempotence":     true,
				"acks":                   "all",
				"transactional.id":       "prefix-topic-A",
				"transaction.timeout.ms": 60000,
			},
		},
		{
			name: "at-least-once fallback config",
			input: confluent.KafkaConfig{
				Brokers:    []string{"localhost:9092"},
				Topic:      "default-topic",
				BatchBytes: 4096,
			},
			wantConfig: map[string]interface{}{
				"bootstrap.servers": "localhost:9092",
				"retry.backoff.ms":  0,
				"message.max.bytes": 4096,
				"acks":              "1",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfgMap := tt.input.ToConfluentProducerConfig()
			for k, v := range tt.wantConfig {
				got, ok := (*cfgMap)[k]
				assert.True(t, ok, "expected key %s to be present", k)
				assert.Equal(t, v, got, "mismatch on key %s", k)
			}
		})
	}
}
func TestKafkaConfig_ToConfluentConsumerConfig(t *testing.T) {
	tests := []struct {
		name       string
		input      confluent.KafkaConfig
		groupID    string
		wantConfig map[string]interface{}
	}{
		{
			name:    "basic config with exactly-once",
			groupID: "group-1",
			input: confluent.KafkaConfig{
				Brokers: []string{"broker1:9092", "broker2:9092"},
				ConsumerConfig: confluent.ConsumerConfig{
					AutoOffsetReset:   "earliest",
					AutoCommit:        true,
					SessionTimeout:    12 * time.Second,
					HeartbeatInterval: 3 * time.Second,
					CommitInterval:    2 * time.Second,
				},
				ReadTimeout:       4 * time.Second,
				DeliverySemantics: confluent.ExactlyOnce,
				ExactlyOnceConfig: confluent.ExactlyOnceConfig{
					IsolationLevel: "read_committed",
				},
			},
			wantConfig: map[string]interface{}{
				"bootstrap.servers":       "broker1:9092,broker2:9092",
				"group.id":                "group-1",
				"auto.offset.reset":       "earliest",
				"enable.auto.commit":      true,
				"socket.timeout.ms":       4000,
				"session.timeout.ms":      12000,
				"heartbeat.interval.ms":   3000,
				"auto.commit.interval.ms": 2000,
				"isolation.level":         "read_committed",
			},
		},
		{
			name:    "minimal config, no exactly-once",
			groupID: "group-x",
			input: confluent.KafkaConfig{
				Brokers: []string{"localhost:9092"},
				ConsumerConfig: confluent.ConsumerConfig{
					AutoOffsetReset: "latest",
					AutoCommit:      false,
				},
			},
			wantConfig: map[string]interface{}{
				"bootstrap.servers":  "localhost:9092",
				"group.id":           "group-x",
				"auto.offset.reset":  "latest",
				"enable.auto.commit": false,
			},
		},
		{
			name:    "custom session and heartbeat intervals",
			groupID: "group-custom",
			input: confluent.KafkaConfig{
				Brokers: []string{"broker1:9092", "broker2:9092"},
				ConsumerConfig: confluent.ConsumerConfig{
					AutoOffsetReset:   "latest",
					AutoCommit:        true,
					SessionTimeout:    15 * time.Second,
					HeartbeatInterval: 5 * time.Second,
					CommitInterval:    3 * time.Second,
				},
				ReadTimeout:       6 * time.Second,
				DeliverySemantics: confluent.ExactlyOnce,
				ExactlyOnceConfig: confluent.ExactlyOnceConfig{
					IsolationLevel: "read_committed",
				},
			},
			wantConfig: map[string]interface{}{
				"bootstrap.servers":       "broker1:9092,broker2:9092",
				"group.id":                "group-custom",
				"auto.offset.reset":       "latest",
				"enable.auto.commit":      true,
				"socket.timeout.ms":       6000,
				"session.timeout.ms":      15000,
				"heartbeat.interval.ms":   5000,
				"auto.commit.interval.ms": 3000,
				"isolation.level":         "read_committed",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cm := tt.input.ToConfluentConsumerConfig(tt.groupID)
			for key, want := range tt.wantConfig {
				got, exists := (*cm)[key]
				assert.True(t, exists, "expected key %s to exist", key)
				assert.Equal(t, want, got, "expected key %s to be %v, got %v", key, want, got)
			}
		})
	}
}
