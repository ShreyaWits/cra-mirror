package handlers

import (
	"context"
	pb "encryption_microservice/internal/common/proto_gen"
	"fmt"
	"time"

	otelcodes "go.opentelemetry.io/otel/codes"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// HandleGenerateEDEK handles the EDEK generation request
func (h *EncryptionHandlerImpl) GenerateEDEK(ctx context.Context, req *pb.GenerateEDEKRequest) (*pb.GenerateEDEKResponse, error) {
	functionName := "GenerateEDEK"
	functionFailed := "GenerateEDEK_Failed"

	// Create context with timeout using caller's context
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	// Start tracing with the caller's context
	ctx, span := h.obs.TracerService.StartTracer(ctx, functionName)
	defer h.obs.TracerService.StopSpan(span)

	// Increment metrics counter
	h.obs.MetricsService.IncrementCounter(ctx, functionName, 1, nil)

	// Generate request ID for correlation
	requestID := getRequestID(ctx)
	h.obs.LoggerService.Info(ctx, fmt.Sprintf("[%s] STEP: Request received", requestID))

	startTime := time.Now()
	token := req.Token // Never log the token

	// Set tracing attributes - only request ID
	h.obs.TracerService.SetAttributes(span, map[string]string{
		"request_id": requestID,
	})

	// Authentication step
	userData, customErr := h.userService.GetUserData(ctx, token)
	if customErr != nil {
		// Log only status without details
		h.obs.LoggerService.Error(ctx, fmt.Sprintf("[%s] STATUS: Authentication failed", requestID))
		h.obs.TracerService.SetStatus(span, otelcodes.Error, "Authentication failed")
		h.obs.MetricsService.IncrementCounter(ctx, functionFailed, 1, map[string]string{
			"error": "authentication_failed",
		})
		return nil, status.Error(codes.Unauthenticated, customErr.Error())
	}

	// Key generation step
	h.obs.LoggerService.Info(ctx, fmt.Sprintf("[%s] STEP: Generating keys", requestID))
	response, customErr := h.encryptionUseCase.GenerateEDEK(ctx, userData.ID)
	if customErr != nil {
		// Log only status without details
		h.obs.LoggerService.Error(ctx, fmt.Sprintf("[%s] STATUS: Key generation failed", requestID))
		h.obs.TracerService.SetStatus(span, otelcodes.Error, "Key generation failed")
		h.obs.MetricsService.IncrementCounter(ctx, functionFailed, 1, map[string]string{
			"error": "operation_failed",
		})
		return nil, status.Error(codes.Internal, customErr.Error())
	}

	// User update step
	h.obs.LoggerService.Info(ctx, fmt.Sprintf("[%s] STEP: Updating user", requestID))
	if updateErr := h.userService.UpdateUser(ctx, userData.ID, response.EDEKPrivate, response.EDEKPublic); updateErr != nil {
		// Log only status without details
		h.obs.LoggerService.Error(ctx, fmt.Sprintf("[%s] STATUS: User update failed", requestID))
		h.obs.TracerService.SetStatus(span, otelcodes.Error, "User update failed")
		h.obs.MetricsService.IncrementCounter(ctx, functionFailed, 1, map[string]string{
			"error": "operation_failed",
		})
		return nil, status.Error(codes.Internal, updateErr.Error())
	}

	// Record metrics for processing time - only basic operational metrics
	processingTime := time.Since(startTime).Milliseconds()
	h.obs.MetricsService.RecordHistogram(ctx, "processing_time_ms", float64(processingTime), nil)

	h.obs.LoggerService.Info(ctx, fmt.Sprintf("[%s] STATUS: Success", requestID))
	h.obs.TracerService.SetStatus(span, otelcodes.Ok, "Success")

	return &pb.GenerateEDEKResponse{
		EDEKPrivate: response.EDEKPrivate,
		EDEKPublic:  response.EDEKPublic,
	}, nil
}
