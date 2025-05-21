package errors_test

import (
	stderrors "errors"
	msgerrors "messaging_service/pkg/errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestNewSimpleGRPCError(t *testing.T) {
	// Test creating a simple gRPC error
	err := msgerrors.NewSimpleGRPCError(codes.NotFound, "Resource not found")

	// Verify it's a proper gRPC status error
	st, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.NotFound, st.Code())
	assert.Equal(t, "Resource not found", st.Message())
}

func TestNewGRPCError(t *testing.T) {
	// Test with nil error (edge case)
	err := msgerrors.NewGRPCError(nil)
	st, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.Internal, st.Code())
	assert.Equal(t, "Unknown error", st.Message())

	// Test with custom error with valid error code
	baseErr := stderrors.New("database connection failed")
	customErr := msgerrors.NewCustomError(msgerrors.KAFErrConnectionFailed, baseErr)

	err = msgerrors.NewGRPCError(customErr)
	st, ok = status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.Unavailable, st.Code())
	assert.Contains(t, st.Message(), msgerrors.KAFErrConnectionFailed)
	assert.Contains(t, st.Message(), "Failed to connect to Kafka broker")
	assert.Contains(t, st.Message(), "database connection failed")
}
