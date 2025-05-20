package models

// ConfigServiceResponse is the response from the config service
type ConfigServiceResponse struct {
	StatusCode int                      `json:"status_code"`
	Message    string                   `json:"message"`
	Data       MessaggingConfigResponse `json:"data"`
}

// MessaggingConfigResponse holds all configuration for the messaging service
type MessaggingConfigResponse struct {
	// Kafka settings for client connection
	KafkaBrokers          []string `json:"kafkaBrokers" validate:"required,min=1"`
	KafkaAutoCreateTopics string   `json:"kafkaAutoCreateTopics"`

	// Kafka topic configuration
	KafkaNumPartitions     int `json:"kafkaNumPartitions" validate:"required,min=1"`
	KafkaReplicationFactor int `json:"kafkaReplicationFactor" validate:"required,min=1"`

	// Kafka producer configuration
	KafkaBatchSize             int    `json:"kafkaBatchSize" validate:"required,min=1"`
	KafkaBatchBytes            int    `json:"kafkaBatchBytes" validate:"required,min=1"`
	KafkaBatchTimeoutMs        int    `json:"kafkaBatchTimeoutMs" validate:"required,min=1"`
	KafkaCompressionCodec      string `json:"kafkaCompressionCodec" validate:"required"`
	KafkaMaxAttempts           int    `json:"kafkaMaxAttempts" validate:"required,min=1"`
	KafkaRetryBackoffMs        int    `json:"kafkaRetryBackoffMs" validate:"required,min=1"`
	KafkaReadTimeoutMs         int    `json:"kafkaReadTimeoutMs" validate:"required,min=1"`
	KafkaWriteTimeoutMs        int    `json:"kafkaWriteTimeoutMs" validate:"required,min=1"`
	KafkaRequireActiveListener bool   `json:"kafkaRequireActiveListener"`

	// Kafka topic settings
	KafkaRetentionMs int `json:"kafkaRetentionMs" validate:"required,min=1"`

	// Kafka consumer configuration
	KafkaConsumerMaxWaitMs        int    `json:"kafkaConsumerMaxWaitMs" validate:"required,min=1"`
	KafkaConsumerCommitIntervalMs int    `json:"kafkaConsumerCommitIntervalMs" validate:"required,min=1"`
	KafkaConsumerSessionTimeoutMs int    `json:"kafkaConsumerSessionTimeoutMs" validate:"required,min=1"`
	KafkaConsumerHeartbeatMs      int    `json:"kafkaConsumerHeartbeatMs" validate:"required,min=1"`
	KafkaConsumerMaxPollRecords   int    `json:"kafkaConsumerMaxPollRecords" validate:"required,min=1"`
	KafkaConsumerAutoOffsetReset  string `json:"kafkaConsumerAutoOffsetReset" validate:"required,oneof=earliest latest none"`
	KafkaEnableAutoCommit         bool   `json:"kafkaEnableAutoCommit"`
	KafkaIsolationLevel           string `json:"kafkaIsolationLevel" validate:"required,oneof=read_committed read_uncommitted"`
}

// EnvConfig holds all environment configuration for the messaging service
type EnvConfig struct {
	// Server settings
	GrpcPort string `json:"grpcPort" env:"GRPC_PORT" validate:"required,numeric"`
	HttpPort string `json:"httpPort" env:"HTTP_PORT" validate:"required,numeric"`

	// Config service
	ConfigServiceUrl   string `json:"configServiceUrl" env:"CONFIG_SERVICE_URL" validate:"required,url"`
	ConfigServiceToken string `json:"configServiceToken" env:"CONFIG_SERVICE_TOKEN" validate:"required"`

	// Cache settings
	CACHE_TTL int `json:"cacheTTL" env:"CACHE_TTL" validate:"min=1"`

	// Deployment info
	Environment      string  `json:"environment" env:"ENVIRONMENT" validate:"require"`
	ServiceName      string  `json:"serviceName" env:"SERVICE_NAME" validate:"required"`
	ServiceVersion   string  `json:"serviceVersion" env:"SERVICE_VERSION" validate:"required,semver"`
	CacheUrl         string  `json:"cacheUrl" env:"CACHE_URL" validate:"required"`
	ObservabilityUrl string  `json:"ObservabilityUrl" env:"OBSERVABILITY_URL" validate:"required,url"`
	TLSDisabled      bool    `json:"tlsDisabled" env:"TLSDISABLE"`
	SamplingRatio    float64 `json:"samplingRatio" env:"SAMPLINGRATIO" validate:"min=0,max=1"`
}
