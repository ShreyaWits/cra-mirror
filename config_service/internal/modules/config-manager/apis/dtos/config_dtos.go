package dtos

// SuccessResponse defines the structure for success responses
type SuccessResponse struct {
	Success    bool        `json:"success"`
	StatusCode int         `json:"status_code"`
	Message    string      `json:"message"`
	Data       interface{} `json:"data,omitempty"` // Use omitempty to omit if data is nil
}

// ErrorResponse defines the structure for error responses
type ErrorResponse struct {
	Success    bool        `json:"success"`
	StatusCode int         `json:"status_code"`
	Message    string      `json:"message"`
	Error      interface{} `json:"error,omitempty"` // Use omitempty to omit if error is nil
}
