package mapper

import (
	"encryption_microservice/internal/modules/encryption/api/dtos"
	errors_pkg "encryption_microservice/pkg/errors"
	"fmt"
	"log"
	"reflect"

	"github.com/gofiber/fiber/v2"
)

// NewResponse creates a success response with data.
func NewResponse[T any](c *fiber.Ctx, message string, code int, messageCode string, data T) error {

	// Log errors if applicable
	if code >= 400 {
		log.Printf("[ERROR] %s | Code: %d\n", message, code)
		return NewErrorResponse(c, message, code, messageCode)
	}

	var requestData any
	// Handle data being nil or empty
	if isNil(data) {
		requestData = nil
	} else {
		requestData = data
	}
	// Create success response
	baseResponse := dtos.BaseResponse{
		Message: message,
		Success: true,
		Code:    code,
		Data:    requestData,
	}

	// Log response
	log.Printf("[DEBUG] Response Body: %+v\n", baseResponse)

	// Send JSON response
	return c.Status(code).JSON(baseResponse)
}

// NewErrorResponse creates a structured error response.
func NewErrorResponse(c *fiber.Ctx, message string, code int, messageCode string) error {

	if messageCode == "" {
		fmt.Printf("No Error Code Define")
	}
	codeMesssage := errors_pkg.GetAppErrorMessage(messageCode)

	if codeMesssage != "" {
		message = codeMesssage + ": " + message
	}

	if codeMesssage == "" {
		message = errors_pkg.GetErrorMessage(code)
	}
	// Log error
	log.Printf("[ERROR] %s | Code: %d | Details: %s\n", message, code, message)

	// Create error response

	baseResponse := dtos.BaseResponse{
		Message:   message,
		Success:   false,
		Code:      code,
		ErrorCode: messageCode,
	}

	return c.Status(code).JSON(baseResponse)
}

// isNil checks if the data is nil or its equivalent (e.g., empty string, nil).
func isNil[T any](data T) bool {
	// Use reflection to determine if data is nil or an empty value
	value := reflect.ValueOf(data)
	// If it's a pointer, slice, or map, it can be nil
	if value.Kind() == reflect.Ptr || value.Kind() == reflect.Slice || value.Kind() == reflect.Map {
		return value.IsNil()
	}

	// If it's a string, check if it's empty
	if value.Kind() == reflect.String {
		return value.Len() == 0
	}

	// For other types (e.g., int, bool, etc.), check if their zero value is being used
	return value.IsZero()
}
