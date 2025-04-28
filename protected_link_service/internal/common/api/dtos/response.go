package commonDtos

import "time"

type ApiResponseDto struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   interface{} `json:"error,omitempty"`
}

type ValidationErrorDetail struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type BaseResponseDto struct {
	ID        string     `json:"id"`
	CreatedAt *time.Time `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
}
