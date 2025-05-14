package sendgrid

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// SendGridClient encapsulates the API key and HTTP client
type SendGridClient struct {
	APIKey     string
	HTTPClient *http.Client
}

// NewSendGridClient creates a new SendGrid client with the given API key
func NewSendGridClient(apiKey string) *SendGridClient {
	return &SendGridClient{
		APIKey:     apiKey,
		HTTPClient: &http.Client{},
	}
}

// Email represents the structure of a SendGrid email
type Email struct {
	Personalizations []Personalization `json:"personalizations"`
	From             Contact           `json:"from"`
	ReplyTo          Contact           `json:"reply_to"`
	Content          []Content         `json:"content"`
}

type Personalization struct {
	To      []Contact `json:"to"`
	Subject string    `json:"subject"`
}

type Contact struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

type Content struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

// SendEmail sends an email using SendGrid API
func (c *SendGridClient) SendEmail(email Email) error {
	url := "https://api.sendgrid.com/v3/mail/send"

	payload, err := json.Marshal(email)
	if err != nil {
		return fmt.Errorf("failed to marshal email payload: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("sendgrid API error: status %d", resp.StatusCode)
	}

	return nil
}
