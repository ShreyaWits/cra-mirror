package smtp

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"gopkg.in/gomail.v2"
)

// MockDialer is a mock implementation of gomail.Dialer
type MockDialer struct {
	shouldFail bool
}

func (m *MockDialer) DialAndSend(msg ...*gomail.Message) error {
	if m.shouldFail {
		return fmt.Errorf("mock error")
	}
	return nil
}

func TestSMTP(t *testing.T) {
	// Save original dialer
	originalDialer := newDialer
	defer func() {
		newDialer = originalDialer
	}()

	tests := []struct {
		name        string
		config      *EmailConfig
		toEmail     string
		subject     string
		body        string
		mockDialer  *MockDialer
		wantErr     bool
		errContains string
	}{
		{
			name: "Valid email configuration",
			config: &EmailConfig{
				SMTPHost:     "smtp.example.com",
				SMTPPort:     587,
				SMTPUser:     "test@example.com",
				SMTPPassword: "password",
			},
			toEmail:    "recipient@example.com",
			subject:    "Test Subject",
			body:       "Test Body",
			mockDialer: &MockDialer{shouldFail: false},
			wantErr:    false,
		},
		{
			name: "Invalid SMTP host",
			config: &EmailConfig{
				SMTPHost:     "invalid-host",
				SMTPPort:     587,
				SMTPUser:     "test@example.com",
				SMTPPassword: "password",
			},
			toEmail:     "recipient@example.com",
			subject:     "Test Subject",
			body:        "Test Body",
			mockDialer:  &MockDialer{shouldFail: true},
			wantErr:     true,
			errContains: "could not send email",
		},
		{
			name: "Invalid port",
			config: &EmailConfig{
				SMTPHost:     "smtp.example.com",
				SMTPPort:     0,
				SMTPUser:     "test@example.com",
				SMTPPassword: "password",
			},
			toEmail:     "recipient@example.com",
			subject:     "Test Subject",
			body:        "Test Body",
			mockDialer:  &MockDialer{shouldFail: true},
			wantErr:     true,
			errContains: "could not send email",
		},
		{
			name: "Empty recipient",
			config: &EmailConfig{
				SMTPHost:     "smtp.example.com",
				SMTPPort:     587,
				SMTPUser:     "test@example.com",
				SMTPPassword: "password",
			},
			toEmail:     "",
			subject:     "Test Subject",
			body:        "Test Body",
			mockDialer:  &MockDialer{shouldFail: true},
			wantErr:     true,
			errContains: "invalid recipient",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up mock dialer
			newDialer = func(host string, port int, username, password string) Dialer {
				return tt.mockDialer
			}

			err := SMTP(tt.config, tt.toEmail, tt.subject, tt.body)
			if (err != nil) != tt.wantErr {
				t.Errorf("SMTP() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && !strings.Contains(err.Error(), tt.errContains) {
				t.Errorf("SMTP() error = %v, should contain %v", err, tt.errContains)
			}
		})
	}
}

func TestSMTPWithEnvVars(t *testing.T) {
	// Save original environment variables
	originalHost := os.Getenv("SMTP_HOST")
	originalPort := os.Getenv("SMTP_PORT")
	originalUser := os.Getenv("SMTP_USER")
	originalPass := os.Getenv("SMTP_PASSWORD")
	defer func() {
		os.Setenv("SMTP_HOST", originalHost)
		os.Setenv("SMTP_PORT", originalPort)
		os.Setenv("SMTP_USER", originalUser)
		os.Setenv("SMTP_PASSWORD", originalPass)
	}()

	// Set up test environment variables
	os.Setenv("SMTP_HOST", "smtp.example.com")
	os.Setenv("SMTP_PORT", "587")
	os.Setenv("SMTP_USER", "test@example.com")
	os.Setenv("SMTP_PASSWORD", "password")

	config := &EmailConfig{
		SMTPHost:     os.Getenv("SMTP_HOST"),
		SMTPPort:     587,
		SMTPUser:     os.Getenv("SMTP_USER"),
		SMTPPassword: os.Getenv("SMTP_PASSWORD"),
	}

	// Set up mock dialer
	originalDialer := newDialer
	defer func() {
		newDialer = originalDialer
	}()
	newDialer = func(host string, port int, username, password string) Dialer {
		return &MockDialer{shouldFail: false}
	}

	err := SMTP(config, "recipient@example.com", "Test Subject", "Test Body")
	if err != nil {
		t.Errorf("SMTP() with env vars error = %v, wantErr false", err)
	}
}
