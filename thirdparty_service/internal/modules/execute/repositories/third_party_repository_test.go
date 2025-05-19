package repositories

import (
	"errors"
	"testing"
)

func TestRepository_VerifyAadhaar(t *testing.T) {
	repo := New()

	tests := []struct {
		name          string
		aadhaar       string
		expectedValid bool
		expectedName  string
		expectedDOB   string
		expectedErr   error
	}{
		{
			name:          "Valid Aadhaar",
			aadhaar:       "123412341234",
			expectedValid: true,
			expectedName:  "John Doe",
			expectedDOB:   "1990-01-01",
			expectedErr:   nil,
		},
		{
			name:          "Invalid Aadhaar",
			aadhaar:       "invalid",
			expectedValid: false,
			expectedName:  "",
			expectedDOB:   "",
			expectedErr:   errors.New("invalid Aadhaar"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, name, dob, err := repo.VerifyAadhaar(tt.aadhaar)

			if valid != tt.expectedValid {
				t.Errorf("Expected valid %v, got %v", tt.expectedValid, valid)
			}
			if name != tt.expectedName {
				t.Errorf("Expected name %s, got %s", tt.expectedName, name)
			}
			if dob != tt.expectedDOB {
				t.Errorf("Expected DOB %s, got %s", tt.expectedDOB, dob)
			}
			if (err != nil && tt.expectedErr != nil && err.Error() != tt.expectedErr.Error()) || (err != nil && tt.expectedErr == nil) || (err == nil && tt.expectedErr != nil) {
				t.Errorf("Expected error %v, got %v", tt.expectedErr, err)
			}
		})
	}
}

func TestRepository_VerifyPAN(t *testing.T) {
	repo := New()

	tests := []struct {
		name          string
		pan           string
		expectedValid bool
		expectedName  string
		expectedErr   error
	}{
		{
			name:          "Valid PAN",
			pan:           "ABCDE1234F",
			expectedValid: true,
			expectedName:  "John Doe",
			expectedErr:   nil,
		},
		{
			name:          "Invalid PAN",
			pan:           "invalid",
			expectedValid: false,
			expectedName:  "",
			expectedErr:   errors.New("invalid PAN"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, name, err := repo.VerifyPAN(tt.pan)

			if valid != tt.expectedValid {
				t.Errorf("Expected valid %v, got %v", tt.expectedValid, valid)
			}
			if name != tt.expectedName {
				t.Errorf("Expected name %s, got %s", tt.expectedName, name)
			}
			if (err != nil && tt.expectedErr != nil && err.Error() != tt.expectedErr.Error()) || (err != nil && tt.expectedErr == nil) || (err == nil && tt.expectedErr != nil) {
				t.Errorf("Expected error %v, got %v", tt.expectedErr, err)
			}
		})
	}
}

func TestRepository_SendSMS(t *testing.T) {
	repo := New()

	tests := []struct {
		name           string
		phone          string
		message        string
		expectedStatus string
		expectedErr    error
	}{
		{
			name:           "Send SMS Success",
			phone:          "1234567890",
			message:        "Test message",
			expectedStatus: "sent",
			expectedErr:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, err := repo.SendSMS(tt.phone, tt.message)

			if status != tt.expectedStatus {
				t.Errorf("Expected status %s, got %s", tt.expectedStatus, status)
			}
			if err != tt.expectedErr {
				t.Errorf("Expected error %v, got %v", tt.expectedErr, err)
			}
		})
	}
}

func TestRepository_SendEmail(t *testing.T) {
	repo := New()

	tests := []struct {
		name           string
		to             string
		subject        string
		body           string
		expectedStatus string
		expectedErr    error
	}{
		{
			name:           "Send Email Success",
			to:             "test@example.com",
			subject:        "Test Subject",
			body:           "Test Body",
			expectedStatus: "sent",
			expectedErr:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, err := repo.SendEmail(tt.to, tt.subject, tt.body)

			if status != tt.expectedStatus {
				t.Errorf("Expected status %s, got %s", tt.expectedStatus, status)
			}
			if err != tt.expectedErr {
				t.Errorf("Expected error %v, got %v", tt.expectedErr, err)
			}
		})
	}
}

func TestRepository_InitiatePayment(t *testing.T) {
	repo := New()

	tests := []struct {
		name                string
		userID              string
		amount              float64
		expectedTransactionID string
		expectedStatus      string
		expectedErr         error
	}{
		{
			name:                "Initiate Payment Success",
			userID:              "user123",
			amount:              100.50,
			expectedTransactionID: "txn_user123",
			expectedStatus:      "success",
			expectedErr:         nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transactionID, status, err := repo.InitiatePayment(tt.userID, tt.amount)

			if transactionID != tt.expectedTransactionID {
				t.Errorf("Expected transaction ID %s, got %s", tt.expectedTransactionID, transactionID)
			}
			if status != tt.expectedStatus {
				t.Errorf("Expected status %s, got %s", tt.expectedStatus, status)
			}
			if err != tt.expectedErr {
				t.Errorf("Expected error %v, got %v", tt.expectedErr, err)
			}
		})
	}
}