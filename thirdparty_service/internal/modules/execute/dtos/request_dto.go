package dtos

type TwilioSmsRequest struct {
	Phone       string `json:"phone" validate:"required,numeric" error_code:"TS1001"`
	Message     string `json:"message" validate:"required" error_code:"TS1002"`
	CountryCode string `json:"country_code" validate:"required" error_code:"TS1003"`
}

type SendGridEmailRequest struct {
	Subject string `json:"from" validate:"required,email" error_code:"TS1004"`
	To      string `json:"to" validate:"required,email" error_code:"TS1005"`
	Body    string `json:"body" validate:"required" error_code:"TS1006"`
	Type    string `json:"type" validate:"required,oneof=TEXT HTML" error_code:"TS1007"`
}

type SendWhatsAppMessageRequest struct {
	Phone       string `json:"phone" validate:"required,e164"`          // e.g., +919876543210
	Message     string `json:"message" validate:"required"`             // message body
	CountryCode string `json:"country_code" validate:"required,len=2"`  // e.g., "IN"
}

type VerifyAadharRequest struct {
	AadharNumber string `json:"number" validate:"required,min=12,max=12" error_code:"TS1010"`
}
type VerifyPANRequest struct {
	PANNumber string `json:"number" validate:"required,min=10,max=10" error_code:"TS1011"`
}
