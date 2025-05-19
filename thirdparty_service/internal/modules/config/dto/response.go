package dto

type ConfigResponse struct {
	HTTPListenAddress string `validate:"required" json:"HTTP_LISTEN_ADDRESS"`
	HTTPListenPort    int    `validate:"required,min=1,max=65535" json:"HTTP_LISTEN_PORT"`

	GRPCListenAddress string `validate:"required" json:"GRPC_LISTEN_ADDRESS"`
	GRPCListenPort    int    `validate:"required,min=1,max=65535" json:"GRPC_LISTEN_PORT"`

	DatabaseHost     string `validate:"required" json:"DATABASE_HOST"`
	DatabasePort     int    `validate:"required,min=1,max=65535" json:"DATABASE_PORT"`
	DatabaseUser     string `validate:"required" json:"DATABASE_USER"`
	DatabasePassword string `validate:"required" json:"DATABASE_PASSWORD"`
	DatabaseName     string `validate:"required" json:"DATABASE_NAME"`

	RedisHost string `validate:"required" json:"REDIS_HOST"`
	RedisPort int    `validate:"required,min=1,max=65535" json:"REDIS_PORT"`

	TwilioAccountSID string `validate:"required" json:"TWILIO_ACCOUNT_SID"`
	TwilioAuthToken  string `validate:"required" json:"TWILIO_AUTH_TOKEN"`
	TwilioFormNumber string `validate:"required" json:"TWILIO_FORM_NUMBER"`

	SendGridApiKey                string `validate:"required" json:"SENDGRID_API_KEY"`
	SendGridFromEmail             string `validate:"required" json:"SENDGRID_FROM_EMAIL"`
	SendGridFromName              string `validate:"required" json:"SENDGRID_FROM_NAME"`
	SendWhatsAppMessageSID        string `validate:"required" json:"SEND_WHATSAPP_MESSAGE_SID"`
	SendWhatsAppMessageToken      string `validate:"required" json:"SEND_WHATSAPP_MESSAGE_TOKEN"`
	SendWhatsAppMessageFromNumber string `validate:"required" json:"SEND_WHATSAPP_MESSAGE_FROM_NUMBER"`
}
