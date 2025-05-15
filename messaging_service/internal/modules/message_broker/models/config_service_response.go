package models

type ConfigServiceResponse struct {
	StatusCode int                      `json:"status_code"`
	Message    string                   `json:"message"`
	Data       MessaggingConfigResponse `json:"data"`
}

// Config holds all configuration for the application
type MessaggingConfigResponse struct {
	// Server settings
	GrpcPort string `json:"grpcPort"`
	HttpPort string `json:"httpPort"`

	// Kafka settings for client connection
	KafkaBrokers          []string `json:"kafkaBrokers"`
	KafkaAutoCreateTopics string   `json:"kafkaAutoCreateTopics"`

	// Kafka topic configuration
	KafkaNumPartitions     int `json:"kafkaNumPartitions"`
	KafkaReplicationFactor int `json:"kafkaReplicationFactor"`

	// Kafka producer configuration
	KafkaBatchSize             int    `json:"kafkaBatchSize"`
	KafkaBatchBytes            int    `json:"kafkaBatchBytes"`
	KafkaBatchTimeoutMs        int    `json:"kafkaBatchTimeoutMs"`
	KafkaCompressionCodec      string `json:"kafkaCompressionCodec"`
	KafkaMaxAttempts           int    `json:"kafkaMaxAttempts"`
	KafkaRetryBackoffMs        int    `json:"kafkaRetryBackoffMs"`
	KafkaReadTimeoutMs         int    `json:"kafkaReadTimeoutMs"`
	KafkaWriteTimeoutMs        int    `json:"kafkaWriteTimeoutMs"`
	KafkaRequireActiveListener bool   `json:"kafkaRequireActiveListener"`

	// Kafka topic settings
	KafkaRetentionMs int `json:"kafkaRetentionMs"`

	// Kafka consumer configuration
	KafkaConsumerMaxWaitMs        int    `json:"kafkaConsumerMaxWaitMs"`
	KafkaConsumerCommitIntervalMs int    `json:"kafkaConsumerCommitIntervalMs"`
	KafkaConsumerSessionTimeoutMs int    `json:"kafkaConsumerSessionTimeoutMs"`
	KafkaConsumerHeartbeatMs      int    `json:"kafkaConsumerHeartbeatMs"`
	KafkaConsumerMaxPollRecords   int    `json:"kafkaConsumerMaxPollRecords"`
	KafkaConsumerAutoOffsetReset  string `json:"kafkaConsumerAutoOffsetReset"`
	KafkaEnableAutoCommit         bool   `json:"kafkaEnableAutoCommit"`
	KafkaIsolationLevel           string `json:"kafkaIsolationLevel"`

	// Config service
	ConfigServiceUrl   string `json:"configServiceUrl"`
	ConfigServiceToken string `json:"configServiceToken"`

	// Deployment info
	Environment string `json:"environment"`
	ServiceName string `json:"serviceName"`
}
