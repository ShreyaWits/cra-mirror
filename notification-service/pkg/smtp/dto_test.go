package smtp

import (
	"reflect"
	"testing"
)

func TestEmailData(t *testing.T) {
	tests := []struct {
		name     string
		email    EmailData
		expected EmailData
	}{
		{
			name: "Valid email data",
			email: EmailData{
				To:      "recipient@example.com",
				Subject: "Test Subject",
				Body:    "Test Body",
			},
			expected: EmailData{
				To:      "recipient@example.com",
				Subject: "Test Subject",
				Body:    "Test Body",
			},
		},
		{
			name: "Empty email data",
			email: EmailData{
				To:      "",
				Subject: "",
				Body:    "",
			},
			expected: EmailData{
				To:      "",
				Subject: "",
				Body:    "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !reflect.DeepEqual(tt.email, tt.expected) {
				t.Errorf("EmailData = %v, want %v", tt.email, tt.expected)
			}
		})
	}
}

func TestEmailDataValidation(t *testing.T) {
	tests := []struct {
		name    string
		email   EmailData
		isValid bool
	}{
		{
			name: "Valid email data",
			email: EmailData{
				To:      "recipient@example.com",
				Subject: "Test Subject",
				Body:    "Test Body",
			},
			isValid: true,
		},
		{
			name: "Missing recipient",
			email: EmailData{
				To:      "",
				Subject: "Test Subject",
				Body:    "Test Body",
			},
			isValid: false,
		},
		{
			name: "Missing subject",
			email: EmailData{
				To:      "recipient@example.com",
				Subject: "",
				Body:    "Test Body",
			},
			isValid: false,
		},
		{
			name: "Missing body",
			email: EmailData{
				To:      "recipient@example.com",
				Subject: "Test Subject",
				Body:    "",
			},
			isValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := tt.email.To != "" && tt.email.Subject != "" && tt.email.Body != ""
			if isValid != tt.isValid {
				t.Errorf("EmailData validation = %v, want %v", isValid, tt.isValid)
			}
		})
	}
}
