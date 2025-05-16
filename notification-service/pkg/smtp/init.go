package smtp

import (
	"fmt"
	"log"

	"gopkg.in/gomail.v2"
)

// EmailConfig holds the configuration for SMTP server
type EmailConfig struct {
	SMTPHost     string // SMTP server hostname
	SMTPPort     int    // SMTP server port (587 for TLS, 465 for SSL)
	SMTPUser     string // SMTP user (e.g., email address)
	SMTPPassword string // SMTP password (e.g., email password or API key)
}

// Dialer interface for mocking
type Dialer interface {
	DialAndSend(msg ...*gomail.Message) error
}

// Default dialer function
var newDialer = func(host string, port int, username, password string) Dialer {
	return gomail.NewDialer(host, port, username, password)
}

// SendEmail is a function to send emails using SMTP
func SMTP(config *EmailConfig, toEmail, subject, body string) error {
	if toEmail == "" {
		return fmt.Errorf("invalid recipient: email address cannot be empty")
	}

	// Set up the SMTP server and email details
	mailer := gomail.NewMessage()
	mailer.SetHeader("From", config.SMTPUser)
	mailer.SetHeader("To", toEmail)
	mailer.SetHeader("Subject", subject)
	mailer.SetBody("text/plain", body)

	// Create dialer and send the email
	dialer := newDialer(config.SMTPHost, config.SMTPPort, config.SMTPUser, config.SMTPPassword)
	if err := dialer.DialAndSend(mailer); err != nil {
		log.Printf("Error sending email to %s: %v", toEmail, err)
		return fmt.Errorf("could not send email: %v", err)
	}

	log.Printf("Email sent to %s successfully", toEmail)
	return nil
}
