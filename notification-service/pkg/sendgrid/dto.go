package sendgrid

type EmailPayload struct {
	ApiKey      string `json:"api_key" validate:"required"`
	FromEmail   string `json:"from_email" validate:"required"`
	Recipient   string `json:"recipient" validate:"required"`
	Subject     string `json:"subject" validate:"required"`
	Message     string `json:"message" validate:"required"`
	EmailClient SendGridClient
}
