package twilio

import (
	"errors"
	"testing"

	openapi "github.com/twilio/twilio-go/rest/api/v2010"
)

// MockTwilioClient is a mock implementation of the Twilio client
type MockTwilioClient struct {
	shouldFail bool
}

func (m *MockTwilioClient) CreateMessage(params *openapi.CreateMessageParams) (*openapi.ApiV2010Message, error) {
	if m.shouldFail {
		return nil, errors.New("mock error")
	}
	return &openapi.ApiV2010Message{}, nil
}

func TestNewTwilioClient(t *testing.T) {
	tests := []struct {
		name       string
		accountSID string
		authToken  string
		fromPhone  string
		wantErr    bool
	}{
		{
			name:       "Valid client creation",
			accountSID: "test_sid",
			authToken:  "test_token",
			fromPhone:  "+1234567890",
			wantErr:    false,
		},
		{
			name:       "Empty account SID",
			accountSID: "",
			authToken:  "test_token",
			fromPhone:  "+1234567890",
			wantErr:    true,
		},
		{
			name:       "Empty auth token",
			accountSID: "test_sid",
			authToken:  "",
			fromPhone:  "+1234567890",
			wantErr:    true,
		},
		{
			name:       "Empty from phone",
			accountSID: "test_sid",
			authToken:  "test_token",
			fromPhone:  "",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewTwilioClient(tt.accountSID, tt.authToken, tt.fromPhone)
			if client == nil && !tt.wantErr {
				t.Errorf("NewTwilioClient() returned nil, want non-nil")
			}
			if client != nil && tt.wantErr {
				t.Errorf("NewTwilioClient() returned non-nil, want nil")
			}
		})
	}
}

func TestSendSMS(t *testing.T) {
	tests := []struct {
		name       string
		client     *TwilioSMSClient
		to         string
		message    string
		shouldFail bool
		wantErr    bool
	}{
		{
			name: "Valid SMS",
			client: &TwilioSMSClient{
				from: "+1234567890",
			},
			to:         "+0987654321",
			message:    "Test message",
			shouldFail: false,
			wantErr:    false,
		},
		{
			name: "Invalid recipient number",
			client: &TwilioSMSClient{
				from: "+1234567890",
			},
			to:         "invalid",
			message:    "Test message",
			shouldFail: true,
			wantErr:    true,
		},
		{
			name: "Empty message",
			client: &TwilioSMSClient{
				from: "+1234567890",
			},
			to:         "+0987654321",
			message:    "",
			shouldFail: true,
			wantErr:    true,
		},
		{
			name:       "Nil client",
			client:     nil,
			to:         "+0987654321",
			message:    "Test message",
			shouldFail: true,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.client != nil {
				// Create mock client
				mockClient := &MockTwilioClient{shouldFail: tt.shouldFail}

				// Replace the actual client with mock
				originalClient := tt.client.c
				defer func() {
					tt.client.c = originalClient
				}()
				tt.client.c = mockClient
			}

			err := tt.client.SendSMS(tt.to, tt.message)
			if (err != nil) != tt.wantErr {
				t.Errorf("SendSMS() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTwilioSMSClient_Integration(t *testing.T) {
	// This test requires actual Twilio credentials to be set in environment variables
	// It's commented out by default to prevent accidental API calls
	/*
		accountSID := os.Getenv("TWILIO_ACCOUNT_SID")
		authToken := os.Getenv("TWILIO_AUTH_TOKEN")
		fromPhone := os.Getenv("TWILIO_FROM_PHONE")
		toPhone := os.Getenv("TWILIO_TEST_PHONE")

		if accountSID == "" || authToken == "" || fromPhone == "" || toPhone == "" {
			t.Skip("Skipping integration test: missing Twilio credentials")
		}

		client := NewTwilioClient(accountSID, authToken, fromPhone)
		if client == nil {
			t.Fatal("Failed to create Twilio client")
		}

		err := client.SendSMS(toPhone, "Integration test message")
		if err != nil {
			t.Errorf("SendSMS() error = %v, want nil", err)
		}
	*/
}
