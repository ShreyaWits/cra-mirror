package models

type VerifyOTPRequest struct {
	UserID        string `json:"user_id" validate:"required"`
	VerficationId string `json:"verification_id" validate:"required"`
	OTP           string `json:"otp" validate:"required"`
}
