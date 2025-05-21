package dtos

type AuthPayload struct {
	ID   string `json:"id"`
	OTP  string `json:"otp"`
	DbId string `json:"db_id"`
}
