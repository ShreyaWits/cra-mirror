package whatsapp

import (
	"errors"
	"os"
	"testing"

	openapi "github.com/twilio/twilio-go/rest/api/v2010"
)

// MockTwilioClient is a mock implementation of the Twilio client
type MockTwilioClient struct {
	shouldFail bool
}

func (m *MockTwilioClient) CreateMessage(params *openapi.CreateMessageParams) (*openapi.ApiV2010Message, error) {
	if m.shouldFail {
		if params != nil && params.To != nil && *params.To == "whatsapp:invalid" {
			return nil, errors.New("invalid phone number")
		}
		return nil, errors.New("mock error")
	}
	return &openapi.ApiV2010Message{}, nil
}

func TestSendWhatsAppMessage(t *testing.T) {
	// Save original environment variables
	originalSid := os.Getenv("TWILIO_ACCOUNT_SID")
	originalToken := os.Getenv("TWILIO_AUTH_TOKEN")
	defer func() {
		os.Setenv("TWILIO_ACCOUNT_SID", originalSid)
		os.Setenv("TWILIO_AUTH_TOKEN", originalToken)
	}()

	tests := []struct {
		name        string
		to          string
		body        string
		accountSid  string
		authToken   string
		shouldFail  bool
		wantErr     bool
		errContains string
	}{
		{
			name:       "Valid WhatsApp message",
			to:         "+1234567890",
			body:       "Test message",
			accountSid: "test_sid",
			authToken:  "test_token",
			shouldFail: false,
			wantErr:    false,
		},
		{
			name:        "Missing account SID",
			to:          "+1234567890",
			body:        "Test message",
			accountSid:  "",
			authToken:   "test_token",
			shouldFail:  true,
			wantErr:     true,
			errContains: "TWILIO_ACCOUNT_SID environment variable is not set",
		},
		{
			name:        "Missing auth token",
			to:          "+1234567890",
			body:        "Test message",
			accountSid:  "test_sid",
			authToken:   "",
			shouldFail:  true,
			wantErr:     true,
			errContains: "TWILIO_AUTH_TOKEN environment variable is not set",
		},
		{
			name:        "Invalid phone number",
			to:          "invalid",
			body:        "Test message",
			accountSid:  "test_sid",
			authToken:   "test_token",
			shouldFail:  true,
			wantErr:     true,
			errContains: "invalid phone number",
		},
		{
			name:        "Empty message body",
			to:          "+1234567890",
			body:        "",
			accountSid:  "test_sid",
			authToken:   "test_token",
			shouldFail:  true,
			wantErr:     true,
			errContains: "message body cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables for the test
			os.Setenv("TWILIO_ACCOUNT_SID", tt.accountSid)
			os.Setenv("TWILIO_AUTH_TOKEN", tt.authToken)

			// Create mock client
			mockClient := &MockTwilioClient{shouldFail: tt.shouldFail}

			// Replace the actual client with mock
			originalClient := newTwilioClient
			defer func() {
				newTwilioClient = originalClient
			}()
			newTwilioClient = func(accountSid, authToken string) TwilioClient {
				return mockClient
			}

			err := SendWhatsAppMessage(tt.to, tt.body)
			if (err != nil) != tt.wantErr {
				t.Errorf("SendWhatsAppMessage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errContains != "" && err != nil {
				if err.Error() != tt.errContains {
					t.Errorf("SendWhatsAppMessage() error = %v, expected %v", err, tt.errContains)
				}
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr
}
