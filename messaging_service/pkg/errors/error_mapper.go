package errors

import (
	pb "cra-protos/messaging_service"
	"fmt"
	"log"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// NewGRPCError converts a CustomError into a gRPC error with structured details.
func NewGRPCError(customErr *CustomError) error {
	if customErr == nil {
		return status.Error(codes.Internal, "Unknown error")
	}

	appMessage := GetAppErrorMessage(customErr.ErrorCode)
	fullMessage := customErr.Error()
	if appMessage != "" {
		fullMessage = fmt.Sprintf("%s: %s: %s", customErr.ErrorCode, appMessage, customErr.Err.Error())
	}

	// Get the appropriate gRPC code for this error
	grpcCode := GetGRPCCode(customErr.ErrorCode)

	log.Printf("[GRPC ERROR] Code: %s | MessageCode: %s | Message: %s\n", grpcCode.String(), customErr.ErrorCode, fullMessage)

	st := status.New(grpcCode, fullMessage)

	detail := &pb.ErrorResponse{
		Success:   false,
		ErrorCode: customErr.ErrorCode,
		Message:   fullMessage,
	}

	stWithDetails, err := st.WithDetails(detail)
	if err != nil {
		return status.Error(codes.Internal, "Failed to attach gRPC error details")
	}

	return stWithDetails.Err()
}

// Helper to wrap only message string
func NewSimpleGRPCError(code codes.Code, msg string) error {
	log.Printf("[GRPC ERROR] Code: %s | Message: %s\n", code.String(), msg)
	return status.Error(code, msg)
}
