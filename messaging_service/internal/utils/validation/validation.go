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
	if len(req.Value) == 0 { // Check only the length of the map
		return errors.NewCustomError(errors.PUBErrInvalidMessage, fmt.Errorf("message value is required"))
	}
	return nil
}

// ValidateSubscribeRequest validates a SubscribeRequest
func ValidateSubscribeRequest(req *pb.SubscribeRequest) *errors.CustomError {
	if req.Topic == "" {
		return errors.NewCustomError(errors.MSGErrInvalidTopic, fmt.Errorf("topic is required"))
	}
	if !IsValidTopicName(req.GroupId) {
		return errors.NewCustomError(errors.MSGErrInvalidGroup, fmt.Errorf("invalid group name: must contain only alphanumeric characters, '.', '_', or '-'"))
	}
	return nil
}

// ValidateCreateTopicRequest validates a CreateTopicRequest
func ValidateCreateTopicRequest(req *pb.CreateTopicRequest) *errors.CustomError {
	if req.Topic == "" {
		return errors.NewCustomError(errors.MSGErrInvalidTopic, fmt.Errorf("topic is required"))
	}
	if !IsValidTopicName(req.Topic) {
		return errors.NewCustomError(errors.MSGErrInvalidTopic, fmt.Errorf("invalid topic name: must contain only alphanumeric characters, '.', '_', or '-'"))
	}
	return nil
}

// IsValidTopicName checks if a topic name is valid
func IsValidTopicName(name string) bool {
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
