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
}
