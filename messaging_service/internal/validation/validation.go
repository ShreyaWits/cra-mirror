package validation

import (
	"fmt"

	pb "cra-protos/messaging_service"
	"messaging_service/pkg/errors"
)

// ValidatePublishRequest validates a PublishRequest
func ValidatePublishRequest(req *pb.PublishRequest) *errors.CustomError {
	if req.Topic == "" {
		return errors.NewCustomError(errors.MSGErrInvalidTopic, fmt.Errorf("topic is required"))
	}
	if len(req.Value) == 0 {
		return errors.NewCustomError(errors.PUBErrInvalidMessage, fmt.Errorf("message value is required"))
	}
	if req.ProducerConfig != nil {
		if err := validateProducerConfig(req.ProducerConfig); err != nil {
			return errors.NewCustomError(errors.PUBErrInvalidConfig, fmt.Errorf("invalid producer config: %w", err))
		}
	}
	return nil
}

// ValidateSubscribeRequest validates a SubscribeRequest
func ValidateSubscribeRequest(req *pb.SubscribeRequest) *errors.CustomError {
	if req.Topic == "" {
		return errors.NewCustomError(errors.MSGErrInvalidTopic, fmt.Errorf("topic is required"))
	}
	if req.GroupId == "" {
		return errors.NewCustomError(errors.MSGErrInvalidGroup, fmt.Errorf("group_id is required"))
	}
	if req.ConsumerConfig != nil {
		if err := validateConsumerConfig(req.ConsumerConfig); err != nil {
			return errors.NewCustomError(errors.SUBErrInvalidConfig, fmt.Errorf("invalid consumer config: %w", err))
		}
	}
	return nil
}

// ValidateCreateTopicRequest validates a CreateTopicRequest
func ValidateCreateTopicRequest(req *pb.CreateTopicRequest) *errors.CustomError {
	if req.Topic == "" {
		return errors.NewCustomError(errors.MSGErrInvalidTopic, fmt.Errorf("topic is required"))
	}
	if !isValidTopicName(req.Topic) {
		return errors.NewCustomError(errors.MSGErrInvalidTopic, fmt.Errorf("invalid topic name: must contain only alphanumeric characters, '.', '_', or '-'"))
	}
	if req.NumPartitions <= 0 {
		return errors.NewCustomError(errors.TOPErrInvalidPartitions, fmt.Errorf("num_partitions must be greater than 0"))
	}
	if req.ReplicationFactor <= 0 {
		return errors.NewCustomError(errors.TOPErrInvalidConfig, fmt.Errorf("replication_factor must be greater than 0"))
	}
	if req.Config != nil {
		if err := validateTopicConfig(req.Config); err != nil {
			return errors.NewCustomError(errors.TOPErrInvalidConfig, fmt.Errorf("invalid topic config: %w", err))
		}
	}
	return nil
}

// validateProducerConfig validates ProducerConfig
func validateProducerConfig(config *pb.ProducerConfig) error {
	if config.Retries < 0 {
		return fmt.Errorf("retries must be non-negative")
	}
	if config.RetryBackoffMs < 0 {
		return fmt.Errorf("retry_backoff_ms must be non-negative")
	}
	return nil
}

// validateConsumerConfig validates ConsumerConfig
func validateConsumerConfig(config *pb.ConsumerConfig) error {
	if config.MaxWaitMs < 0 {
		return fmt.Errorf("max_wait_ms must be non-negative")
	}
	if config.CommitIntervalMs < 0 {
		return fmt.Errorf("commit_interval_ms must be non-negative")
	}
	return nil
}

// validateTopicConfig validates TopicConfig
func validateTopicConfig(config *pb.TopicConfig) error {
	if config.RetentionMs < 0 {
		return fmt.Errorf("retention_ms must be non-negative")
	}
	return nil
}

// isValidTopicName checks if a topic name is valid
func isValidTopicName(name string) bool {
	if name == "" {
		return false
	}
	// Kafka topic names can contain alphanumeric characters, '.', '_', and '-'
	for _, c := range name {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '.' || c == '_' || c == '-') {
			return false
		}
	}
	return true
}
