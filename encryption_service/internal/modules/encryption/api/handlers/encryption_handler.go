package handlers

import (
	"context"
	pb "encryption_microservice/internal/common/proto_gen"
	"encryption_microservice/internal/modules/encryption/services/user"
	usecases "encryption_microservice/internal/modules/encryption/usecases/encryption"
	"encryption_microservice/pkg/observability"
	"fmt"
	"time"

	"go.opentelemetry.io/otel/trace"
)

const (
	defaultTimeout = 5 * time.Second
)

// EncryptionHandlerImpl implements the EncryptionHandler interface
type EncryptionHandlerImpl struct {
	pb.UnimplementedEncryptionServiceServer
	encryptionUseCase usecases.EncryptionUseCase
	userService       user.UserService
	obs               *observability.ObservabilityStack
}

// NewEncryptionHandler creates a new instance of EncryptionHandlerImpl
func NewEncryptionHandler(
	encryptionUseCase usecases.EncryptionUseCase,
	userService user.UserService,
	obs *observability.ObservabilityStack,
) *EncryptionHandlerImpl {
	// Validate input arguments
	if encryptionUseCase == nil {
		panic("encryptionUseCase cannot be nil in NewEncryptionHandler")
	}
	if userService == nil {
		panic("userService cannot be nil in NewEncryptionHandler")
	}
	if obs == nil {
		panic("observability stack cannot be nil in NewEncryptionHandler")
	}

	return &EncryptionHandlerImpl{
		encryptionUseCase: encryptionUseCase,
		userService:       userService,
		obs:               obs,
	}
}

// getRequestID generates a unique request ID for tracing and logging
func getRequestID(ctx context.Context) string {
	// Try to get trace ID from OpenTelemetry context
	spanContext := trace.SpanContextFromContext(ctx)
	if spanContext.IsValid() {
		return fmt.Sprintf("req-%s", spanContext.TraceID().String())
	}

	// Try to get span from context (backward compatibility)
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().IsValid() {
		return fmt.Sprintf("req-%s", span.SpanContext().TraceID().String())
	}

	// Legacy approach, try to get span from context value
	if spanVal := ctx.Value("span"); spanVal != nil {
		return fmt.Sprintf("req-%s", spanVal)
	}

	// Generate a simple timestamp-based ID as fallback
	return fmt.Sprintf("req-%d", time.Now().UnixNano())
}
