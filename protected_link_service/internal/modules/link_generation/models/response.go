package models

// BaseResponse is the common response structure.
type BaseResponse struct {
	Message string `json:"message" example:"Success"`
	Success bool   `json:"success" example:"true"`
	Code    int    `json:"code,omitempty" example:"200"` // Optional error code
	Data    any    `json:"data,omitempty"`               // Optional data, example omitted since it's dynamic
}

type ErrorResponseSwagger struct {
	Message string `json:"message" example:"Invalid Request"`
	Success bool   `json:"success" example:"false"`
	Code    int    `json:"code,omitempty" example:"400"` // Optional error code
}

// GenerateUrlResponse extends BaseResponse for URL generation.
type GenerateUrlResponse struct {
	RedirectURL string `json:"redirectUrl,omitempty" example:"https://example.com/auth?token=abc123"`
}

// GenerateUrlInfoResponse holds user-specific URL information.
type GenerateUrlInfoResponse struct {
	UserID      string `json:"user_id"`
	RequestType string `json:"request_type"`
}

// ErrorResponse extends BaseResponse for error cases.
type ErrorResponse struct {
	ErrorMessage string `json:"error_details,omitempty"`
}

type FieldErrorResponseDTO struct {
	ErrorCode    string `json:"errorCode"`
	ErrorMessage string `json:"errorMsg"`
	Field        string `json:"field"`
}

type GetUrlResponse struct {
	RequestType string      `json:"request_type" validate:"required" example:"hybrid"`
	Data        interface{} `json:"data" validate:"required"`
}

type OtpResponse struct {
	RequestId   string `json:"requestId" validate:"required" exmple:"df34fd6c-cc78-4afa-9ad4-f0dd64ecddfe"`
	OTPRequired bool   `json:"otpRequired" validate:"required" example:"true"`
}
