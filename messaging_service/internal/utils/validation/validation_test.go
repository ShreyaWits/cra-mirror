package validation_test

import (
	"fmt"
	"testing"

	pb "cra-protos/messaging_service"
	"messaging_service/internal/validation"
	"messaging_service/pkg/errors"

	"github.com/stretchr/testify/assert"
)

func TestValidatePublishRequest(t *testing.T) {
	tests := []struct {
		name     string
		req      *pb.PublishRequest
		expected *errors.CustomError
	}{
		{
			name:     "Valid request",
			req:      &pb.PublishRequest{Topic: "valid-topic", Value: map[string]string{"key": "message"}},
			expected: nil,
		},
		{
			name:     "Missing topic",
			req:      &pb.PublishRequest{Topic: "", Value: map[string]string{"key": "message"}},
			expected: errors.NewCustomError(errors.MSGErrInvalidTopic, fmt.Errorf("topic is required")),
		},
		{
			name:     "Missing value",
			req:      &pb.PublishRequest{Topic: "valid-topic", Value: nil},
			expected: errors.NewCustomError(errors.PUBErrInvalidMessage, fmt.Errorf("value is required")),
		},
		{
			name:     "Empty value",
			req:      &pb.PublishRequest{Topic: "valid-topic", Value: map[string]string{}},
			expected: errors.NewCustomError(errors.PUBErrInvalidMessage, fmt.Errorf("value is required")),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validation.ValidatePublishRequest(tt.req)
			if tt.expected != nil {
				assert.Error(t, result, tt.expected.Error())
			} else {
				assert.Nil(t, result)
			}
		})
	}
}

func TestValidateSubscribeRequest(t *testing.T) {
	tests := []struct {
		name     string
		req      *pb.SubscribeRequest
		expected *errors.CustomError
	}{
		{
			name:     "Valid request",
			req:      &pb.SubscribeRequest{Topic: "valid-topic", GroupId: "valid-group"},
			expected: nil,
		},
		{
			name:     "Missing topic",
			req:      &pb.SubscribeRequest{Topic: "", GroupId: "valid-group"},
			expected: errors.NewCustomError(errors.MSGErrInvalidTopic, fmt.Errorf("topic is required")),
		},
		{
			name:     "Missing group ID",
			req:      &pb.SubscribeRequest{Topic: "valid-topic", GroupId: ""},
			expected: errors.NewCustomError(errors.MSGErrInvalidGroup, fmt.Errorf("groupId is required")),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validation.ValidateSubscribeRequest(tt.req)
			if tt.expected != nil {
				assert.Error(t, result, tt.expected.Error())
			} else {
				assert.Nil(t, result)
			}
		})
	}
}

func TestValidateCreateTopicRequest(t *testing.T) {
	tests := []struct {
		name     string
		req      *pb.CreateTopicRequest
		expected *errors.CustomError
	}{
		{
			name:     "Valid request",
			req:      &pb.CreateTopicRequest{Topic: "valid-topic"},
			expected: nil,
		},
		{
			name:     "Missing topic",
			req:      &pb.CreateTopicRequest{Topic: ""},
			expected: errors.NewCustomError(errors.MSGErrInvalidTopic, fmt.Errorf("topic is required")),
		},
		{
			name:     "Invalid topic name",
			req:      &pb.CreateTopicRequest{Topic: "invalid_topic!"},
			expected: errors.NewCustomError(errors.MSGErrInvalidTopic, fmt.Errorf("Invalid topic")),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validation.ValidateCreateTopicRequest(tt.req)
			if tt.expected != nil {
				assert.Error(t, result, tt.expected.Error())
			} else {
				assert.Nil(t, result)
			}
		})
	}
}

func TestIsValidTopicName(t *testing.T) {
	tests := []struct {
		name     string
		topic    string
		expected bool
	}{
		{
			name:     "Valid topic name with alphanumeric characters",
			topic:    "valid-topic_123",
			expected: true,
		},
		{
			name:     "Valid topic name with dots",
			topic:    "valid.topic-name",
			expected: true,
		},
		{
			name:     "Valid topic name with underscores",
			topic:    "valid_topic-name",
			expected: true,
		},
		{
			name:     "Valid topic name with hyphens",
			topic:    "valid-topic-name",
			expected: true,
		},
		{
			name:     "Empty topic name",
			topic:    "",
			expected: false,
		},
		{
			name:     "Invalid topic name with spaces",
			topic:    "invalid topic",
			expected: false,
		},
		{
			name:     "Invalid topic name with special characters",
			topic:    "invalid@topic!",
			expected: false,
		},
		{
			name:     "Invalid topic name with only special characters",
			topic:    "!@#$%^&*()",
			expected: false,
		},
		{
			name:     "Invalid topic name with mixed valid and invalid characters",
			topic:    "valid-topic#name",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validation.IsValidTopicName(tt.topic)
			assert.Equal(t, tt.expected, result)
		})
	}
}
