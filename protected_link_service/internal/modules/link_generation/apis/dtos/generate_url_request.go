package apiDtos

type GenerateUrlRequest struct {
	UserID      string `json:"user_id" validate:"required"`
	Name        string `json:"name" validate:"required"`
	RequestType string `json:"request_type" validate:"required"`
	ModelType   string `json:"model_type" validate:"required,oneof=jwt hybrid"`
	Email       string `json:"email" validate:"required,email"`
	ExpireIn    string `json:"expire_in" validate:"required"`
	OtpRequired bool   `json:"otp_required"`
	Phone       string `json:"phone" validate:"required"`
	ChannelType string `json:"channel_type" validate:"required"`
	Data        JSONB  `json:"data" validate:"required"`
}
