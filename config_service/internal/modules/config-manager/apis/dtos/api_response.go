package dtos

import "fmt"

type ApiResponseDto struct {
	Success bool              `json:"success"`
	Error   *ErrorResponseDto `json:"error,omitempty"`
	Data    interface{}       `json:"data,omitempty"`
}

type ErrorResponseDto struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type ServiceErrorResponse struct {
	ErrorCode    string `json:"data,omitempty"`
	ErrorMessage string `json:"message,omitempty"`
	StatusCode   int16  `json:"code,omitempty"`
}

// Error implements error.
func (s *ServiceErrorResponse) Error() string {
	return fmt.Sprintf("error: %s, message: %s, code: %d", s.ErrorCode, s.ErrorMessage, s.StatusCode)
}
