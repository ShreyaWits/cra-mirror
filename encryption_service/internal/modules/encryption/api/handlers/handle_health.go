package handlers

import (
	"context"
	pb "encryption_microservice/internal/common/proto_gen"
	"fmt"
	"time"

	otelcodes "go.opentelemetry.io/otel/codes"
)

// HandleHealth handles the health check request
func (h *EncryptionHandlerImpl) HealthCheck(ctx context.Context, req *pb.HealthCheckRequest) (*pb.HealthCheckResponse, error) {
	functionName := "HealthCheck"

	// Start tracing with the caller's context
	ctx, span := h.obs.TracerService.StartTracer(ctx, functionName)
	defer h.obs.TracerService.StopSpan(span)

	// Increment metrics counter
	h.obs.MetricsService.IncrementCounter(ctx, functionName, 1, nil)

	// Generate request ID for correlation
	requestID := getRequestID(ctx)

	startTime := time.Now()

	// Set tracing attributes - only include request ID
	h.obs.TracerService.SetAttributes(span, map[string]string{
		"request_id": requestID,
	})

	// Log health check step
	h.obs.LoggerService.Info(ctx, fmt.Sprintf("[%s] STEP: Health check", requestID))

	// Record metrics for processing time - only basic operational metrics
	processingTime := time.Since(startTime).Milliseconds()
	h.obs.MetricsService.RecordHistogram(ctx, "processing_time_ms", float64(processingTime), nil)

	// Increment success counter
	h.obs.MetricsService.IncrementCounter(ctx, "operation_success", 1, nil)

	h.obs.LoggerService.Info(ctx, fmt.Sprintf("[%s] STATUS: Healthy", requestID))
	h.obs.TracerService.SetStatus(span, otelcodes.Ok, "Success")

	return &pb.HealthCheckResponse{
		HealthStatus: "OK",
	}, nil
}
