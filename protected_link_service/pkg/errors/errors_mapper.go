package errors

import (
	"fmt"
	"log"
	"protected_link/pkg/grpc/proto"

	// pb "protected_link/pkg/grpc/proto/auth.proto"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/anypb"
)

// // NewResponse creates a success response with data.
// func NewResponse[T any](c *fiber.Ctx, message string, code int, messageCode string, data T) error {

// 	// Log errors if applicable
// 	if code >= 400 {
// 		log.Printf("[ERROR] %s | Code: %d\n", message, code)
// 		return NewErrorResponse(c, message, code, messageCode)
// 	}

// 	var requestData any
// 	// Handle data being nil or empty
// 	if isNil(data) {
// 		requestData = nil
// 	} else {
// 		requestData = data
// 	}
// 	// Create success response
// 	baseResponse := dtos.BaseResponse{
// 		Message: message,
// 		Success: true,
// 		Code:    code,
// 		Data:    requestData,
// 	}

// 	// Log response
// 	log.Printf("[DEBUG] Response Body: %+v\n", baseResponse)

// 	// Send JSON response
// 	return c.Status(code).JSON(baseResponse)
// }

// // NewErrorResponse creates a structured error response.
// func NewErrorResponse(c *fiber.Ctx, message string, code int, messageCode string) error {

// 	if messageCode == "" {
// 		fmt.Printf("No Error Code Define")
// 	}
// 	codeMesssage := GetErrorMessage(code)

// 	if codeMesssage != "" {
// 		message = codeMesssage + ": " + message
// 	}

// 	if codeMesssage == "" {
// 		message = GetErrorMessage(code)
// 	}
// 	// Log error
// 	log.Printf("[ERROR] %s | Code: %d | Details: %s\n", message, code, message)

// 	// Create error response

// 	baseResponse := dtos.BaseResponse{
// 		Message:   message,
// 		Success:   false,
// 		Code:      code,
// 		ErrorCode: messageCode,
// 	}

// 	return c.Status(code).JSON(baseResponse)
// }

// // isNil checks if the data is nil or its equivalent (e.g., empty string, nil).
// func isNil[T any](data T) bool {
// 	// Use reflection to determine if data is nil or an empty value
// 	value := reflect.ValueOf(data)
// 	// If it's a pointer, slice, or map, it can be nil
// 	if value.Kind() == reflect.Ptr || value.Kind() == reflect.Slice || value.Kind() == reflect.Map {
// 		return value.IsNil()
// 	}

// 	// If it's a string, check if it's empty
// 	if value.Kind() == reflect.String {
// 		return value.Len() == 0
// 	}

// 	// For other types (e.g., int, bool, etc.), check if their zero value is being used
// 	return value.IsZero()
// }

// NewGRPCError converts a CustomError into a gRPC error with structured details.
func NewGRPCError(customErr *CustomError) error {
	if customErr == nil {
		return status.Error(codes.Internal, "Unknown error")
	}

	appMessage := GetAppErrorMessage(customErr.ErrorCode)
	fullMessage := customErr.Error()
	if appMessage != "" {
		fullMessage = fmt.Sprintf("%s: %s: %s", customErr.ErrorCode, appMessage, customErr.Error())
	}

	log.Printf("[GRPC ERROR] Code: %s | MessageCode: %s | Message: %s\n", customErr.code, customErr.ErrorCode, fullMessage)

	st := status.New(customErr.code, fullMessage)

	detail := &proto.ErrorResponse{
		Success:   false,
		ErrorCode: customErr.ErrorCode,
		Message:   fullMessage,
	}
	detailProto, err := anypb.New(detail)
	if err != nil {
		return status.Error(codes.Internal, "Failed to wrap gRPC error details")
	}

	stWithDetails, err := st.WithDetails(detailProto)
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
