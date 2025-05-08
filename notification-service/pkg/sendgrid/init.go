package sendgrid

import (
	"fmt"
	"log"

	sendgridSDK "github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

// SendGridResponse is a local representation of SendGrid's internal response struct
type SendGridResponse struct {
	StatusCode int
}

// SendGridClient is an interface for the SendGrid client's Send method
type SendGridClient interface {
	Send(email *mail.SGMailV3) (*SendGridResponse, error)
}

// realSendGridClient wraps the real SendGrid client to satisfy our interface
type realSendGridClient struct {
	client *sendgridSDK.Client
}

func (r *realSendGridClient) Send(email *mail.SGMailV3) (*SendGridResponse, error) {
	resp, err := r.client.Send(email)
	if err != nil {
		return nil, err
	}
	return &SendGridResponse{StatusCode: resp.StatusCode}, nil
}

// InitSendGrid initializes a SendGrid email client with parameters
func InitSendGrid(ApiKey, FromEmail, Recipient, Subject, Message string) *EmailPayload {
	client := &realSendGridClient{client: sendgridSDK.NewSendClient(ApiKey)}
	return &EmailPayload{
		FromEmail:   FromEmail,
		Recipient:   Recipient,
		Subject:     Subject,
		Message:     Message,
		EmailClient: client,
	}
}

// SendEmail sends the email via SendGrid
func (payload *EmailPayload) SendEmail() (bool, error) {
	from := mail.NewEmail("Admin", payload.FromEmail)
	to := mail.NewEmail("User", payload.Recipient)
	message := mail.NewSingleEmail(from, payload.Subject, to, payload.Message, "")

	response, err := payload.EmailClient.Send(message)
	if err != nil {
		log.Printf("Error sending email: %v", err)
		return false, err
	}

	if response.StatusCode >= 200 && response.StatusCode < 300 {
		log.Printf("Email sent successfully with status code: %d\n", response.StatusCode)
		return true, nil
	}

	err = fmt.Errorf("failed to send email, status code: %d", response.StatusCode)
	log.Printf("Error: %v", err)
	return false, err
}
