package errors_test

import (
	"messaging_service/pkg/errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
)

func TestGetErrorMessage(t *testing.T) {
	tests := []struct {
		name     string
		code     int
		expected string
	}{
		{
			name:     "400 Bad Request",
			code:     400,
			expected: "Bad Request: Invalid request format.",
		},
		{
			name:     "404 Not Found",
			code:     404,
			expected: "Not Found: Resource does not exist.",
		},
		{
			name:     "500 Internal Server Error",
			code:     500,
			expected: "Internal Server Error: Something went wrong.",
		},
		{
			name:     "Unknown code",
			code:     999,
			expected: "Something went wrong999",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			message := errors.GetErrorMessage(tt.code)
			assert.Equal(t, tt.expected, message)
		})
	}
}

func TestGetAppErrorMessage(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{
			name:     "Invalid request code",
			code:     errors.MSGErrInvalidRequest,
			expected: "Invalid request format",
		},
		{
			name:     "Producer not ready",
			code:     errors.PUBErrProducerNotReady,
			expected: "Kafka producer is not ready",
		},
		{
			name:     "Unknown error code",
			code:     "UNKNOWN",
			expected: "Unknown error occurred",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			message := errors.GetAppErrorMessage(tt.code)
			assert.Equal(t, tt.expected, message)
		})
	}
}

func TestGetGRPCCode(t *testing.T) {
	tests := []struct {
		name         string
		errorCode    string
		expectedCode codes.Code
	}{
		{
			name:         "Invalid argument",
			errorCode:    errors.MSGErrInvalidRequest,
			expectedCode: codes.InvalidArgument,
		},
		{
			name:         "Not found",
			errorCode:    errors.PUBErrTopicNotExists,
			expectedCode: codes.NotFound,
		},
		{
			name:         "Already exists",
			errorCode:    errors.TOPErrTopicExists,
			expectedCode: codes.AlreadyExists,
		},
		{
			name:         "Permission denied",
			errorCode:    errors.KAFErrNotAuthorized,
			expectedCode: codes.PermissionDenied,
		},
		{
			name:         "Unavailable",
			errorCode:    errors.KAFErrBrokerUnavailable,
			expectedCode: codes.Unavailable,
		},
		{
			name:         "Internal error",
			errorCode:    errors.PUBErrPublishFailed,
			expectedCode: codes.Internal,
		},
		{
			name:         "Unknown error code",
			errorCode:    "UNKNOWN_CODE",
			expectedCode: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code := errors.GetGRPCCode(tt.errorCode)
			assert.Equal(t, tt.expectedCode, code)
		})
	}
}
