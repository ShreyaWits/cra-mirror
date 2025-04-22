package apiDtos

type GenerateUrlRequest struct {
	UserID      string `json:"user_id" validate:"required" example:"1234"`
	RequestType string `json:"request_type" validate:"required" example:"auth"`
	ModelType   string `json:"model_type" validate:"required" example:"hybrid"`
	OtpRequired bool   `json:"otp_required" default:"false" example:"true"`
	ExpireIn    string `json:"expire_in" default:"24h" example:"3"`
	Email       string `json:"email" validate:"required" example:"Tt7b2@example.com"`
	Phone       string `json:"phone" validate:"required" example:"1234567890"`
	ChannelType string `json:"channel_type" validate:"required" example:"email"`
	Data        JSONB  `json:"data" validate:"required" swaggertype:"object" `
}
