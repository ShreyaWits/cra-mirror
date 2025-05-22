package handlers_test

import (
	"context"
	"testing"

	"Document-Processing/internal/handlers"
	"Document-Processing/pkg/observability"
	pb "Document-Processing/proto"

	"github.com/stretchr/testify/assert"
)

// TestHealthCheck_Serving tests the Check method for a valid service
func TestHealthCheck_Serving(t *testing.T) {
	handler := handlers.NewHealthHandler(*observability.NewObservabilityStack())

	req := &pb.HealthCheckRequest{
		Service: "documentprocessing.DocumentProcessingService",
	}

	expectedResponse := &pb.HealthCheckResponse{
		Status: pb.HealthCheckResponse_SERVING,
	}

	res, err := handler.Check(context.Background(), req)

	assert.NoError(t, err)
	assert.Equal(t, expectedResponse.Status, res.Status)
}

// TestHealthCheck_EmptyService tests the Check method for an empty service name
func TestHealthCheck_EmptyService(t *testing.T) {
	handler := handlers.NewHealthHandler(*observability.NewObservabilityStack())

	req := &pb.HealthCheckRequest{
		Service: "",
	}

	expectedResponse := &pb.HealthCheckResponse{
		Status: pb.HealthCheckResponse_UNKNOWN,
	}

	res, err := handler.Check(context.Background(), req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "service name is required")
	assert.Equal(t, expectedResponse.Status, res.Status)
}

// TestHealthCheck_UnsupportedService tests the Check method for an unsupported service name
func TestHealthCheck_UnsupportedService(t *testing.T) {
	handler := handlers.NewHealthHandler(*observability.NewObservabilityStack())

	req := &pb.HealthCheckRequest{
		Service: "some.other.Service",
	}

	expectedResponse := &pb.HealthCheckResponse{
		Status: pb.HealthCheckResponse_NOT_SERVING,
	}

	res, err := handler.Check(context.Background(), req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported service: some.other.Service")
	assert.Equal(t, expectedResponse.Status, res.Status)
}
