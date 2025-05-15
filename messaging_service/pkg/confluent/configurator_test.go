package confluent_test

import (
	pb "cra-protos/messaging_service"
	"messaging_service/internal/config"
	"messaging_service/pkg/confluent"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestConfigurator_CreatePublishConfig(t *testing.T) {
	tests := []struct {
		name   string
		conf   *config.Config
		expect func(cfg confluent.KafkaConfig)
	}{
		{
			name: "with full config",
			conf: &config.Config{
				KafkaMaxAttempts:      4,
				KafkaRetryBackoffMs:   150,
				KafkaBatchSize:        300,
				KafkaBatchBytes:       1024 * 1024,
				KafkaBatchTimeoutMs:   500,
				KafkaReadTimeoutMs:    1000,
				KafkaWriteTimeoutMs:   1000,
				KafkaCompressionCodec: "lz4",
			},
			expect: func(cfg confluent.KafkaConfig) {
				assert.Equal(t, 4, cfg.MaxAttempts)
				assert.Equal(t, 150, cfg.RetryBackoffMs)
				assert.Equal(t, 300, cfg.BatchSize)
				assert.Equal(t, 1024*1024, cfg.BatchBytes)
				assert.Equal(t, 500*time.Millisecond, cfg.BatchTimeout)
				assert.Equal(t, time.Second, cfg.ReadTimeout)
				assert.Equal(t, time.Second, cfg.WriteTimeout)
				assert.Equal(t, "lz4", cfg.CompressionType)
				assert.True(t, cfg.ExactlyOnceConfig.EnableIdempotence)
				assert.True(t, cfg.ExactlyOnceConfig.EnableTransactions)
			},
		},
		{
			name: "with nil config (defaults)",
			conf: nil,
			expect: func(cfg confluent.KafkaConfig) {
				assert.Equal(t, 3, cfg.MaxAttempts)
				assert.Equal(t, 100, cfg.RetryBackoffMs)
				assert.Equal(t, 100, cfg.BatchSize)
				assert.Equal(t, 1*1024*1024, cfg.BatchBytes)
				assert.Equal(t, 500*time.Millisecond, cfg.BatchTimeout)
				assert.Equal(t, 5*time.Second, cfg.ReadTimeout)
				assert.Equal(t, 5*time.Second, cfg.WriteTimeout)
				assert.Equal(t, "snappy", cfg.CompressionType)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := confluent.NewConfigurator([]string{"localhost:9092"}, tt.conf)
			cfg := c.CreatePublishConfig("topic-x", &pb.PublishRequest{})
			tt.expect(cfg)
		})
	}
}

func TestConfigurator_CreateSubscribeConfig(t *testing.T) {
	tests := []struct {
		name   string
		conf   *config.Config
		expect func(cfg confluent.KafkaConfig)
	}{
		{
			name: "with full config",
			conf: &config.Config{
				KafkaMaxAttempts:              5,
				KafkaRetryBackoffMs:           200,
				KafkaBatchSize:                500,
				KafkaBatchBytes:               2 * 1024 * 1024,
				KafkaBatchTimeoutMs:           300,
				KafkaReadTimeoutMs:            1500,
				KafkaWriteTimeoutMs:           1500,
				KafkaCompressionCodec:         "gzip",
				KafkaConsumerMaxWaitMs:        250,
				KafkaConsumerCommitIntervalMs: 1000,
				KafkaConsumerSessionTimeoutMs: 10000,
				KafkaConsumerHeartbeatMs:      3000,
				KafkaConsumerMaxPollRecords:   800,
				KafkaConsumerAutoOffsetReset:  "latest",
				KafkaEnableAutoCommit:         true,
				KafkaIsolationLevel:           "read_committed",
			},
			expect: func(cfg confluent.KafkaConfig) {
				assert.Equal(t, 5, cfg.MaxAttempts)
				assert.Equal(t, 200, cfg.RetryBackoffMs)
				assert.Equal(t, 500, cfg.BatchSize)
				assert.Equal(t, 2*1024*1024, cfg.BatchBytes)
				assert.Equal(t, 300*time.Millisecond, cfg.BatchTimeout)
				assert.Equal(t, 1500*time.Millisecond, cfg.ReadTimeout)
				assert.Equal(t, 1500*time.Millisecond, cfg.WriteTimeout)
				assert.Equal(t, "gzip", cfg.CompressionType)

				assert.Equal(t, 250*time.Millisecond, cfg.ConsumerConfig.MaxWait)
				assert.Equal(t, 1000*time.Millisecond, cfg.ConsumerConfig.CommitInterval)
				assert.Equal(t, 10000*time.Millisecond, cfg.ConsumerConfig.SessionTimeout)
				assert.Equal(t, 3000*time.Millisecond, cfg.ConsumerConfig.HeartbeatInterval)
				assert.Equal(t, 800, cfg.ConsumerConfig.MaxPollRecords)
				assert.Equal(t, "latest", cfg.ConsumerConfig.AutoOffsetReset)
				assert.Equal(t, true, cfg.ConsumerConfig.AutoCommit)

				assert.Equal(t, confluent.ExactlyOnce, cfg.DeliverySemantics)
				assert.True(t, cfg.ExactlyOnceConfig.EnableDeduplication)
				assert.Equal(t, "read_committed", cfg.ExactlyOnceConfig.IsolationLevel)
			},
		},
		{
			name: "with full config",
			conf: &config.Config{
				KafkaMaxAttempts:              5,
				KafkaRetryBackoffMs:           200,
				KafkaBatchSize:                500,
				KafkaBatchBytes:               2 * 1024 * 1024,
				KafkaBatchTimeoutMs:           300,
				KafkaReadTimeoutMs:            1500,
				KafkaWriteTimeoutMs:           1500,
				KafkaCompressionCodec:         "gzip",
				KafkaConsumerMaxWaitMs:        250,
				KafkaConsumerCommitIntervalMs: 1000,
				KafkaConsumerSessionTimeoutMs: 10000,
				KafkaConsumerHeartbeatMs:      3000,
				KafkaConsumerMaxPollRecords:   800,
				KafkaConsumerAutoOffsetReset:  "",
				KafkaEnableAutoCommit:         true,
				KafkaIsolationLevel:           "",
			},
			expect: func(cfg confluent.KafkaConfig) {
				assert.Equal(t, 5, cfg.MaxAttempts)
				assert.Equal(t, 200, cfg.RetryBackoffMs)
				assert.Equal(t, 500, cfg.BatchSize)
				assert.Equal(t, 2*1024*1024, cfg.BatchBytes)
				assert.Equal(t, 300*time.Millisecond, cfg.BatchTimeout)
				assert.Equal(t, 1500*time.Millisecond, cfg.ReadTimeout)
				assert.Equal(t, 1500*time.Millisecond, cfg.WriteTimeout)
				assert.Equal(t, "gzip", cfg.CompressionType)

				assert.Equal(t, 250*time.Millisecond, cfg.ConsumerConfig.MaxWait)
				assert.Equal(t, 1000*time.Millisecond, cfg.ConsumerConfig.CommitInterval)
				assert.Equal(t, 10000*time.Millisecond, cfg.ConsumerConfig.SessionTimeout)
				assert.Equal(t, 3000*time.Millisecond, cfg.ConsumerConfig.HeartbeatInterval)
				assert.Equal(t, 800, cfg.ConsumerConfig.MaxPollRecords)
				assert.Equal(t, "latest", cfg.ConsumerConfig.AutoOffsetReset)
				assert.Equal(t, true, cfg.ConsumerConfig.AutoCommit)

				assert.Equal(t, confluent.ExactlyOnce, cfg.DeliverySemantics)
				assert.True(t, cfg.ExactlyOnceConfig.EnableDeduplication)
				assert.Equal(t, "read_committed", cfg.ExactlyOnceConfig.IsolationLevel)
			},
		},
		{
			name: "with nil config (defaults)",
			conf: nil,
			expect: func(cfg confluent.KafkaConfig) {
				assert.Equal(t, 3, cfg.MaxAttempts)
				assert.Equal(t, 100, cfg.RetryBackoffMs)
				assert.Equal(t, 100, cfg.BatchSize)
				assert.Equal(t, 1*1024*1024, cfg.BatchBytes)
				assert.Equal(t, 500*time.Millisecond, cfg.BatchTimeout)
				assert.Equal(t, 5*time.Second, cfg.ReadTimeout)
				assert.Equal(t, 5*time.Second, cfg.WriteTimeout)
				assert.Equal(t, "snappy", cfg.CompressionType)

				assert.Equal(t, 100*time.Millisecond, cfg.ConsumerConfig.MaxWait)
				assert.Equal(t, 1000*time.Millisecond, cfg.ConsumerConfig.CommitInterval)
				assert.Equal(t, 500, cfg.ConsumerConfig.MaxPollRecords)
				assert.Equal(t, false, cfg.ConsumerConfig.AutoCommit)

				assert.Equal(t, "read_committed", cfg.ExactlyOnceConfig.IsolationLevel)
				assert.Equal(t, "latest", cfg.ConsumerConfig.AutoOffsetReset)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := confluent.NewConfigurator([]string{"localhost:9092"}, tt.conf)
			cfg := c.CreateSubscribeConfig("topic-x", "group-xyz", &pb.SubscribeRequest{})
			tt.expect(cfg)
		})
	}
}

func TestConfigurator_CreateAdminConfig(t *testing.T) {
	tests := []struct {
		name   string
		conf   *config.Config
		expect func(cfg confluent.KafkaConfig)
	}{
		{
			name: "with full config",
			conf: &config.Config{
				KafkaReadTimeoutMs:  7000,
				KafkaWriteTimeoutMs: 9000,
			},
			expect: func(cfg confluent.KafkaConfig) {
				assert.Equal(t, 7*time.Second, cfg.ReadTimeout)
				assert.Equal(t, 9*time.Second, cfg.WriteTimeout)
			},
		},
		{
			name: "with nil config (defaults)",
			conf: nil,
			expect: func(cfg confluent.KafkaConfig) {
				assert.Equal(t, 30*time.Second, cfg.ReadTimeout)
				assert.Equal(t, 30*time.Second, cfg.WriteTimeout)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := confluent.NewConfigurator([]string{"localhost:9092"}, tt.conf)
			cfg := c.CreateAdminConfig()
			tt.expect(cfg)
		})
	}
}
func TestConfigurator_CreateTopicConfig(t *testing.T) {
	tests := []struct {
		name   string
		conf   *config.Config
		expect func(cfg confluent.KafkaConfig)
	}{
		{
			name: "with full config",
			conf: &config.Config{
				KafkaNumPartitions:     5,
				KafkaReplicationFactor: 2,
				KafkaBatchSize:         200,
				KafkaBatchBytes:        512 * 1024,
				KafkaBatchTimeoutMs:    300,
				KafkaCompressionCodec:  "zstd",
			},
			expect: func(cfg confluent.KafkaConfig) {
				assert.Equal(t, 5, cfg.NumPartitions)
				assert.Equal(t, 2, cfg.ReplicationFactor)
				assert.Equal(t, 200, cfg.BatchSize)
				assert.Equal(t, 512*1024, cfg.BatchBytes)
				assert.Equal(t, 300*time.Millisecond, cfg.BatchTimeout)
				assert.Equal(t, "zstd", cfg.CompressionType)
			},
		},
		{
			name: "with nil config (defaults)",
			conf: nil,
			expect: func(cfg confluent.KafkaConfig) {
				assert.Equal(t, 3, cfg.NumPartitions)
				assert.Equal(t, 3, cfg.ReplicationFactor)
				assert.Equal(t, 100, cfg.BatchSize)
				assert.Equal(t, 1*1024*1024, cfg.BatchBytes)
				assert.Equal(t, 500*time.Millisecond, cfg.BatchTimeout)
				assert.Equal(t, "snappy", cfg.CompressionType)
			},
		},
		{
			name: "enforces min values for partitions and replication",
			conf: &config.Config{
				KafkaNumPartitions:     0,
				KafkaReplicationFactor: -1,
				KafkaBatchSize:         150,
				KafkaBatchBytes:        256 * 1024,
				KafkaBatchTimeoutMs:    100,
				KafkaCompressionCodec:  "lz4",
			},
			expect: func(cfg confluent.KafkaConfig) {
				assert.Equal(t, 1, cfg.NumPartitions)
				assert.Equal(t, 1, cfg.ReplicationFactor)
				assert.Equal(t, 150, cfg.BatchSize)
				assert.Equal(t, 256*1024, cfg.BatchBytes)
				assert.Equal(t, 100*time.Millisecond, cfg.BatchTimeout)
				assert.Equal(t, "lz4", cfg.CompressionType)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := confluent.NewConfigurator([]string{"localhost:9092"}, tt.conf)
			cfg := c.CreateTopicConfig("example-topic", &pb.CreateTopicRequest{})
			tt.expect(cfg)
		})
	}
}
