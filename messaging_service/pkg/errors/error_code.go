package errors

import (
	"fmt"

	"google.golang.org/grpc/codes"
)

// GetErrorMessage returns the error message related to a specific HTTP status code.
func GetErrorMessage(code int) string {
	statusMessages := map[int]string{
		// 4XX Client Errors
		400: "Bad Request: Invalid request format.",
		401: "Unauthorized: Authentication required.",
		403: "Forbidden: Access is denied.",
		404: "Not Found: Resource does not exist.",
		405: "Method Not Allowed: Unsupported request method.",
		409: "Conflict: Duplicate or conflicting request.",
		422: "Unprocessable Entity: Validation failed.",

		// 5XX Server Errors
		500: "Internal Server Error: Something went wrong.",
		502: "Bad Gateway: Invalid upstream response.",
		503: "Service Unavailable: Server is overloaded or under maintenance.",
		504: "Gateway Timeout: No response from upstream service.",
	}

	if message, exists := statusMessages[code]; exists {
		return message
	}
	return fmt.Sprintf("Something went wrong%d", code)
}

// GetAppErrorMessage returns the error message related to a specific application error code.
func GetAppErrorMessage(code string) string {
	statusMessages := map[string]string{
		// MSGxxx (Messaging Service)
		MSGErrInvalidRequest: "Invalid request format",
		MSGErrInvalidTopic:   "Invalid topic name or format",
		MSGErrInvalidGroup:   "Invalid consumer group ID",
		MSGErrArgument:       "Invalid argument",
		// PUBxxx (Publish)
		PUBErrInvalidMessage:   "Invalid message format or content",
		PUBErrPublishFailed:    "Failed to publish message to topic",
		PUBErrInvalidConfig:    "Invalid publisher configuration",
		PUBErrTopicNotExists:   "Topic does not exist",
		PUBErrProducerNotReady: "Kafka producer is not ready",

		// SUBxxx (Subscribe)
		SUBErrInvalidConfig:    "Invalid subscriber configuration",
		SUBErrSubscribeFailed:  "Failed to subscribe to topic",
		SUBErrTopicNotExists:   "Topic does not exist",
		SUBErrConsumerNotReady: "Kafka consumer is not ready",
		SUBErrInvalidGroupID:   "Invalid consumer group ID",
		SUBErrStreamError:      "Error in message stream",

		// TOPxxx (Topic Management)
		TOPErrInvalidConfig:     "Invalid topic configuration",
		TOPErrCreateFailed:      "Failed to create topic",
		TOPErrTopicExists:       "Topic already exists",
		TOPErrInvalidPartitions: "Invalid number of partitions",

		// KAFxxx (Kafka Client)
		KAFErrConnectionFailed:  "Failed to connect to Kafka broker",
		KAFErrBrokerUnavailable: "Kafka broker is unavailable",
		KAFErrNotAuthorized:     "Not authorized to perform operation",
		KAFErrNotAuthenticated:  "Authentication failed",

		// CFGxxx (Configuration Management)
		CFGErrFetchFailed:      "Failed to fetch configuration from API",
		CFGErrValidationFailed: "Configuration validation failed",
		CFGErrMarshalFailed:    "Failed to marshal configuration data",
		CFGErrUnmarshalFailed:  "Failed to unmarshal configuration data",

		// CACxxx (Cache Operations)
		CACErrConnectionFailed: "Failed to connect to cache service",
		CACErrSetFailed:        "Failed to set data in cache",
		CACErrGetFailed:        "Failed to get data from cache",
		CACErrNotFound:         "Data not found in cache",
		CACErrInvalidateFailed: "Failed to invalidate cache entry",
	}

	if message, exists := statusMessages[code]; exists {
		return message
	}
	return "Unknown error occurred"
}

// GetGRPCCode maps internal error codes to gRPC status codes
func GetGRPCCode(errorCode string) codes.Code {
	switch errorCode {
	// Invalid argument errors
	case MSGErrInvalidRequest, MSGErrInvalidTopic, MSGErrInvalidGroup,
		PUBErrInvalidMessage, PUBErrInvalidConfig,
		SUBErrInvalidConfig, SUBErrInvalidGroupID,
		TOPErrInvalidConfig, TOPErrInvalidPartitions,
		CFGErrValidationFailed:
		return codes.InvalidArgument

	// Not found errors
	case PUBErrTopicNotExists, SUBErrTopicNotExists, CACErrNotFound:
		return codes.NotFound

	// Already exists errors
	case TOPErrTopicExists:
		return codes.AlreadyExists

	// Permission denied errors
	case KAFErrNotAuthorized, KAFErrNotAuthenticated:
		return codes.PermissionDenied

	// Unavailable errors
	case PUBErrProducerNotReady, SUBErrConsumerNotReady,
		KAFErrConnectionFailed, KAFErrBrokerUnavailable,
		CACErrConnectionFailed:
		return codes.Unavailable

	// Internal errors for operation failures
	case PUBErrPublishFailed, SUBErrSubscribeFailed, SUBErrStreamError,
		TOPErrCreateFailed, CFGErrFetchFailed, CFGErrMarshalFailed,
		CFGErrUnmarshalFailed, CACErrSetFailed, CACErrGetFailed,
		CACErrInvalidateFailed:
		return codes.Internal

	// Default to internal error for unknown error codes
	default:
		return codes.Internal
	}
}
