package handlers

import (
	"context"
	"fmt"

	"Document-Processing/pkg/observability"
	pb "Document-Processing/proto"
	"go.opentelemetry.io/otel/codes"
)

type HealthHandler struct {
	pb.UnimplementedHealthServiceServer
	observability observability.ObservabilityStack
}

func NewHealthHandler(obs observability.ObservabilityStack) *HealthHandler {
	return &HealthHandler{
		observability: obs,
	}
}

func (h *HealthHandler) Check(ctx context.Context, req *pb.HealthCheckRequest) (*pb.HealthCheckResponse, error) {
	traceCtx, span := h.observability.TracerService.StartTracer(ctx, "HealthHandler.Check")
	defer span.End()

	h.observability.TracerService.SetAttributes(span, map[string]string{
		"service.requested": req.Service,
		"handler":           "HealthHandler",
	})

	h.observability.LoggerService.Info(traceCtx, "Received health check request", req.Service)

	// Validation: service name must be provided
	if req.Service == "" {
		err := fmt.Errorf("service name is required")

		h.observability.LoggerService.Error(traceCtx, err.Error(), req.Service)
		h.observability.MetricsService.IncrementCounter(traceCtx, "health_check_failure", 1, map[string]string{
			"error":   "missing_service_name",
			"handler": "HealthHandler",
		})
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())

		return &pb.HealthCheckResponse{
			Status: pb.HealthCheckResponse_UNKNOWN,
		}, err
	}

	// Only documentprocessing is supported
	if req.Service != "documentprocessing.DocumentProcessingService" {
		err := fmt.Errorf("unsupported service: %s", req.Service)

		h.observability.LoggerService.Error(traceCtx, err.Error(), req.Service)
		h.observability.MetricsService.IncrementCounter(traceCtx, "health_check_failure", 1, map[string]string{
			"error":   "unsupported_service",
			"service": req.Service,
		})
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())

		return &pb.HealthCheckResponse{
			Status: pb.HealthCheckResponse_NOT_SERVING,
		}, err
	}

	// Successful check
	h.observability.LoggerService.Info(traceCtx, "Health check successful", req.Service)
	h.observability.MetricsService.IncrementCounter(traceCtx, "health_check_success", 1, map[string]string{
		"service": req.Service,
	})

	span.SetStatus(codes.Ok, "Service is healthy")
	return &pb.HealthCheckResponse{
		Status: pb.HealthCheckResponse_SERVING,
	}, nil
}
