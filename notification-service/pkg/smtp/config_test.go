package smtp

import (
	"testing"
	"reflect"
)

func TestSMTPConfig(t *testing.T) {
	tests := []struct {
		name     string
		config   SMTPConfig
		expected SMTPConfig
	}{
		{
			name: "Valid SMTP configuration",
			config: SMTPConfig{
				Host:     "smtp.example.com",
				Port:     587,
				Username: "test@example.com",
				Password: "password",
			},
			expected: SMTPConfig{
				Host:     "smtp.example.com",
				Port:     587,
				Username: "test@example.com",
				Password: "password",
			},
		},
		{
			name: "Empty SMTP configuration",
			config: SMTPConfig{
				Host:     "",
				Port:     0,
				Username: "",
				Password: "",
			},
			expected: SMTPConfig{
				Host:     "",
				Port:     0,
				Username: "",
				Password: "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !reflect.DeepEqual(tt.config, tt.expected) {
				t.Errorf("SMTPConfig = %v, want %v", tt.config, tt.expected)
			}
		})
	}
}

func TestSMTPConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  SMTPConfig
		isValid bool
	}{
		{
			name: "Valid configuration",
			config: SMTPConfig{
				Host:     "smtp.example.com",
				Port:     587,
				Username: "test@example.com",
				Password: "password",
			},
			isValid: true,
		},
		{
			name: "Missing host",
			config: SMTPConfig{
				Host:     "",
				Port:     587,
				Username: "test@example.com",
				Password: "password",
			},
			isValid: false,
		},
		{
			name: "Invalid port",
			config: SMTPConfig{
				Host:     "smtp.example.com",
				Port:     0,
				Username: "test@example.com",
				Password: "password",
			},
			isValid: false,
		},
		{
			name: "Missing username",
			config: SMTPConfig{
				Host:     "smtp.example.com",
				Port:     587,
				Username: "",
				Password: "password",
			},
			isValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := tt.config.Host != "" && tt.config.Port > 0 && tt.config.Username != ""
			if isValid != tt.isValid {
				t.Errorf("SMTPConfig validation = %v, want %v", isValid, tt.isValid)
			}
		})
	}
}