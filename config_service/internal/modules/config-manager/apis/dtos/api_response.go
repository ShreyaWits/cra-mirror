package dtos

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
