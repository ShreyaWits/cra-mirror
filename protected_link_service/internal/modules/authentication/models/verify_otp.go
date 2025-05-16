package models

type VerifyOTPRequest struct {
	UserID         string `json:"user_id" validate:"required,numeric"`
	OTP            string `json:"otp" validate:"required,numeric"`
	VerificationID string `json:"verification_id" validate:"required"`
}

type VerifyOTPResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   interface{} `json:"error,omitempty"`
}
