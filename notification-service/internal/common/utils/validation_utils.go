package utils

// import (
// 	"notification-service/internal/common/models"
// 	"github.com/go-playground/validator/v10"
// )

// ValidationError represents a single validation error with field and message
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// SuccessResponse represents a structured success response for gRPC
type SuccessResponse struct {
	Status     string      `json:"status"`
	StatusCode int         `json:"status_code"`
	Message    string      `json:"message"`
	Data       interface{} `json:"data"`
}

// ErrorResponse represents a structured error response for gRPC
type ErrorResponse struct {
	Status     string      `json:"status"`
	StatusCode int         `json:"status_code"`
	Message    string      `json:"message"`
	Error      interface{} `json:"error"`
}

// SendSuccessResponse formats the success response
func SendSuccessResponse(statusCode int, message string, data interface{}) *SuccessResponse {
	return &SuccessResponse{
		Status:     "success",
		StatusCode: statusCode,
		Message:    message,
		Data:       data,
	}
}

// SendErrorResponse formats the error response
func SendErrorResponse(statusCode int, message string, err interface{}) *ErrorResponse {
	return &ErrorResponse{
		Status:     "error",
		StatusCode: statusCode,
		Message:    message,
		Error:      err,
	}
}
