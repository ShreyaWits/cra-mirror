package sendgrid

import (
	"testing"

	"github.com/go-playground/validator/v10"
)

func TestEmailPayload_Validation(t *testing.T) {
	validate := validator.New()

	tests := []struct {
		name    string
		payload EmailPayload
		wantErr bool
	}{
		{
			name: "valid payload",
			payload: EmailPayload{
				ApiKey:    "test-api-key",
				FromEmail: "sender@example.com",
				Recipient: "recipient@example.com",
				Subject:   "Test Subject",
				Message:   "Test Message",
			},
			wantErr: false,
		},
		{
			name: "missing ApiKey",
			payload: EmailPayload{
				FromEmail: "sender@example.com",
				Recipient: "recipient@example.com",
				Subject:   "Test Subject",
				Message:   "Test Message",
			},
			wantErr: true,
		},
		{
			name: "missing FromEmail",
			payload: EmailPayload{
				ApiKey:    "test-api-key",
				Recipient: "recipient@example.com",
				Subject:   "Test Subject",
				Message:   "Test Message",
			},
			wantErr: true,
		},
		{
			name: "missing Recipient",
			payload: EmailPayload{
				ApiKey:    "test-api-key",
				FromEmail: "sender@example.com",
				Subject:   "Test Subject",
				Message:   "Test Message",
			},
			wantErr: true,
		},
		{
			name: "missing Subject",
			payload: EmailPayload{
				ApiKey:    "test-api-key",
				FromEmail: "sender@example.com",
				Recipient: "recipient@example.com",
				Message:   "Test Message",
			},
			wantErr: true,
		},
		{
			name: "missing Message",
			payload: EmailPayload{
				ApiKey:    "test-api-key",
				FromEmail: "sender@example.com",
				Recipient: "recipient@example.com",
				Subject:   "Test Subject",
			},
			wantErr: true,
		},
		{
			name:    "empty payload",
			payload: EmailPayload{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validate.Struct(tt.payload)
			if (err != nil) != tt.wantErr {
				t.Errorf("EmailPayload validation error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}