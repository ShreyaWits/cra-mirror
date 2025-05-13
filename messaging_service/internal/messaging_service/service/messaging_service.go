package service

import (
	"context"
	pb "cra-protos/messaging_service"
	errors "messaging_service/pkg/errors"
)

// MessagingService defines the interface for interacting with the messaging system
type MessagingService interface {
	// PublishMessage publishes a message to a topic
	PublishMessage(ctx context.Context, cfg interface{}, req *pb.PublishRequest) *errors.CustomError

	// ConsumeMessage consumes messages from a topic and sends them to the client stream
	ConsumeMessage(stream pb.MessagingService_SubscribeV1Server, cfg interface{}) *errors.CustomError

	// CreateTopic creates a new topic with the specified configuration
	CreateTopic(ctx context.Context, req *pb.CreateTopicRequest, cfg interface{}) *errors.CustomError
}

// KafkaAdmin defines the interface for Kafka admin operations
// This interface is used by the ConfluentMessagingService implementation
type KafkaAdmin interface {
	CreateTopic(ctx context.Context, topic string, numPartitions, replicationFactor int, configs map[string]string) error
	DeleteTopic(ctx context.Context, topic string) error
	ListTopics(ctx context.Context) ([]string, error)
	Close() error
}

// Helper function to convert CleanupPolicy enum to string
// Kept for utility purposes
func getCleanupPolicyString(policy pb.CleanupPolicy) string {
	switch policy {
	case pb.CleanupPolicy_CLEANUP_POLICY_DELETE:
		return "delete"
	case pb.CleanupPolicy_CLEANUP_POLICY_COMPACT:
		return "compact"
	default:
		return "delete"
	}
}
