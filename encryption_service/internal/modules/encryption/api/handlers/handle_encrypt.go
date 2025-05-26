package handlers

import (
	"context"
	pb "encryption_microservice/internal/common/proto_gen"
	"encryption_microservice/internal/modules/encryption/api/dtos"
	"encryption_microservice/internal/modules/encryption/api/mapper"
	"fmt"
	"time"

	otelcodes "go.opentelemetry.io/otel/codes"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (h *EncryptionHandlerImpl) Encrypt(ctx context.Context, req *pb.EncryptDataRequest) (resp *pb.EncryptDataResponse, err error) {
	functionName := "Encrypt"
	functionFailed := "Encrypt_Failed"

	// Create context with timeout using caller's context
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	// Start tracing with the caller's context
	ctx, span := h.obs.TracerService.StartTracer(ctx, functionName)
	defer h.obs.TracerService.StopSpan(span)

	// Increment metrics counter
	h.obs.MetricsService.IncrementCounter(ctx, functionName, 1, map[string]string{
		"method": "Encrypt",
	})

	// Generate request ID for correlation
	requestID := getRequestID(ctx)
	h.obs.LoggerService.Info(ctx, fmt.Sprintf("[%s] STEP: Request received", requestID))

	// Handle panics
	defer func() {
		if r := recover(); r != nil {
			// Log only error status without details
			h.obs.LoggerService.Error(ctx, fmt.Sprintf("[%s] ERROR: Internal server error", requestID))
			h.obs.TracerService.SetStatus(span, otelcodes.Error, "Internal server error")
			h.obs.MetricsService.IncrementCounter(ctx, functionFailed, 1, map[string]string{
				"error": "panic",
			})
			resp = nil
			err = status.Error(codes.Internal, "Internal server error")
		}
	}()

	// Extract token - NEVER log the actual token
	token := req.Token
	startTime := time.Now()

	// Get user data - only log status
	userData, customErr := h.userService.GetUserData(ctx, token)
	if customErr != nil {
		// Log only authentication status without details
		h.obs.LoggerService.Error(ctx, fmt.Sprintf("[%s] STATUS: Authentication failed", requestID))
		h.obs.TracerService.SetStatus(span, otelcodes.Error, "Authentication failed")
		h.obs.MetricsService.IncrementCounter(ctx, functionFailed, 1, map[string]string{
			"error": "authentication_failed",
		})
		return nil, status.Error(codes.Unauthenticated, customErr.Error())
	}

	// Determine key ID
	var keyId string
	if req.UserId == nil || *req.UserId == "" {
		keyId = userData.ID
	} else {
		userData, customErr = h.userService.GetUserData(ctx, *req.UserId)
		if customErr != nil {
			// Log only status without details
			h.obs.LoggerService.Error(ctx, fmt.Sprintf("[%s] STATUS: User lookup failed", requestID))
			h.obs.TracerService.SetStatus(span, otelcodes.Error, "User not found")
			h.obs.MetricsService.IncrementCounter(ctx, functionFailed, 1, map[string]string{
				"error": "user_not_found",
			})
			return nil, status.Error(codes.Unauthenticated, customErr.Error())
		}
		keyId = userData.ID
	}

	// Set tracing attributes - only use requestID
	h.obs.TracerService.SetAttributes(span, map[string]string{
		"request_id": requestID,
	})

	// Map request data - never log the content
	mappedRequest := mapper.ConvertFromStructPB(req.Data)
	encryptedReq := &dtos.EncryptRequest{
		Data: mappedRequest,
	}

	// Execute encryption use case - log only step
	h.obs.LoggerService.Info(ctx, fmt.Sprintf("[%s] STEP: Processing", requestID))
	response, customErr := h.encryptionUseCase.Encrypt(ctx, keyId, userData.EDEKPrivate, userData.EDEKPublic, encryptedReq)

	if customErr != nil {
		// Log only error status without details
		h.obs.LoggerService.Error(ctx, fmt.Sprintf("[%s] STATUS: Operation failed", requestID))
		h.obs.TracerService.SetStatus(span, otelcodes.Error, "Operation failed")
		h.obs.MetricsService.IncrementCounter(ctx, functionFailed, 1, map[string]string{
			"error": "operation_failed",
		})
		return nil, status.Error(codes.Internal, customErr.Error())
	}

	// Map response - never log the content
	mappedResponse := mapper.ConvertToStructPB(response.Data)

	// Record metrics for processing time - only basic operational metrics
	processingTime := time.Since(startTime).Milliseconds()
	h.obs.MetricsService.RecordHistogram(ctx, "processing_time_ms", float64(processingTime), nil)

	h.obs.LoggerService.Info(ctx, fmt.Sprintf("[%s] STATUS: Success", requestID))
	h.obs.TracerService.SetStatus(span, otelcodes.Ok, "Success")

	return &pb.EncryptDataResponse{
		Data: mappedResponse,
	}, nil
}

// hashUserID returns a hashed or truncated version of the user ID for logging
// to avoid exposing sensitive information
func hashUserID(userID string) string {
	if len(userID) <= 8 {
		return "id_" + userID[:2] + "***"
	}
	return "id_" + userID[:2] + "***" + userID[len(userID)-2:]
}
