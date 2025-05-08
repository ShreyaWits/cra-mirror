package dto

import "time"

// CreateTemplateRequest represents the request body for creating a template
type CreateTemplateRequest struct {
	Name     string `json:"name" validate:"required"`
	Channel  string `json:"channel" validate:"required,oneof=email sms push"`
	Language string `json:"language" validate:"required,len=2"`
	Content  string `json:"content" validate:"required"`
	IsActive bool   `json:"is_active"`
}

// CreateTemplateResponse represents the response for template creation
type CreateTemplateResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    struct {
		TemplateID string `json:"template_id"`
	} `json:"data"`
}

// GetTemplateRequest represents the request for getting a template
type GetTemplateRequest struct {
	Name     string `query:"name"`
	Channel  string `query:"channel"`
	Language string `query:"language"`
	ID       string `param:"id"`
}

// TemplateResponse represents the template data in responses
type TemplateResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Channel   string    `json:"channel"`
	Language  string    `json:"language"`
	Content   string    `json:"content"`
	Version   int       `json:"version"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// GetTemplateResponse represents the response for getting a template
type GetTemplateResponse struct {
	Success bool             `json:"success"`
	Message string           `json:"message"`
	Data    TemplateResponse `json:"data"`
}

// UpdateTemplateRequest represents the request body for updating a template
type UpdateTemplateRequest struct {
	IsActive bool `json:"is_active"`
}

// UpdateTemplateResponse represents the response for template update
type UpdateTemplateResponse struct {
	Success bool             `json:"success"`
	Message string           `json:"message"`
	Data    TemplateResponse `json:"data"`
}

// DeleteTemplateRequest represents the request for deleting a template
type DeleteTemplateRequest struct {
	ID string `param:"id" validate:"required"`
}

// DeleteTemplateResponse represents the response for template deletion
type DeleteTemplateResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// InterpolateTemplateRequest represents the request for template interpolation
type InterpolateTemplateRequest struct {
	TemplateID string            `json:"template_id" validate:"required"`
	Data       map[string]string `json:"data" validate:"required"`
}

// InterpolateTemplateResponse represents the response for template interpolation
type InterpolateTemplateResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    struct {
		RenderedBody string `json:"rendered_body"`
	} `json:"data"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Success bool   `json:"success"`
	ErrorMessage string `json:"message"`
	ErrorCode   string `json:"error_code"`
	Data

}
