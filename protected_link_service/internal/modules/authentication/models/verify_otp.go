package models

type VerifyOTPRequest struct {
	UserID string `json:"user_id"`
	OTP    string `json:"otp"`
}
