package dtos

import (
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
)

func TestValidateTemplateID(t *testing.T) {
	validate := validator.New()
	RegisterValidations(validate)

	testCases := []struct {
		name       string
		templateID string
		expected   bool
	}{
		{"Valid OTP Verification Template", "otp_verification", true},
		{"Valid NPS Feedback Template", "nps_feedback", true},
		{"Valid Order Placed Template", "order_placed", true},
		{"Invalid Template", "invalid_template", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validate.Var(tc.templateID, "template_id")
			if tc.expected {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestValidatePhone91(t *testing.T) {
	validate := validator.New()
	RegisterValidations(validate)

	testCases := []struct {
		name    string
		phone   string
		isValid bool
	}{
		{"Valid Phone Number", "+919876543210", true},
		{"Missing +91 Prefix", "9876543210", false},
		{"Invalid Phone Number Length", "+9198765432", false},
		{"Invalid Characters", "+91abcdefghij", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validate.Var(tc.phone, "phone")
			if tc.isValid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestValidateChannels(t *testing.T) {
	validate := validator.New()
	RegisterValidations(validate)

	testCases := []struct {
		name     string
		channels []string
		isValid  bool
	}{
		{"Valid Single Channel", []string{"email"}, true},
		{"Valid Multiple Channels", []string{"email", "sms"}, true},
		{"Invalid Channel", []string{"invalid_channel"}, false},
		{"Mixed Valid and Invalid Channels", []string{"email", "invalid_channel"}, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validate.Var(tc.channels, "channel_list")
			if tc.isValid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestValidateRecipient(t *testing.T) {
	validate := validator.New()
	RegisterValidations(validate)

	testCases := []struct {
		name         string
		notification *NotificationRequest
		recipient    Recipient
		expectError  bool
	}{
		{
			name: "Valid Recipient for OTP Verification",
			notification: &NotificationRequest{
				TemplateID: "otp_verification",
				Channels:   []string{"email"},
			},
			recipient: Recipient{
				Email: "test@example.com",
				Data: map[string]string{
					"username": "testuser",
					"otp":      "123456",
				},
			},
			expectError: false,
		},
		{
			name: "Missing Required Data Field",
			notification: &NotificationRequest{
				TemplateID: "otp_verification",
				Channels:   []string{"email"},
			},
			recipient: Recipient{
				Email: "test@example.com",
				Data: map[string]string{
					"username": "testuser",
				},
			},
			expectError: true,
		},
		{
			name: "No Contact Information",
			notification: &NotificationRequest{
				TemplateID: "nps_feedback",
				Channels:   []string{"email"},
			},
			recipient: Recipient{
				Data: map[string]string{
					"username":  "testuser",
					"orderID":   "order123",
					"nps_score": "8",
					"feedback":  "Great service",
				},
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Manually set the context for validation
			validate.RegisterStructValidation(func(sl validator.StructLevel) {
				recipient := sl.Current().Interface().(Recipient)

				if recipient.UserID == "" && recipient.Email == "" && recipient.Phone == "" && recipient.WhatsappNumber == "" {
					sl.ReportError(recipient.UserID, "UserID", "userID", "contact_required", "")
				}

				if tc.notification != nil {
					requiredFields := templateDataRequirements[tc.notification.TemplateID]
					for _, field := range requiredFields {
						if _, ok := recipient.Data[field]; !ok {
							sl.ReportError(recipient.Data, "Data."+field, field, field+"_required", "")
						}
					}
				}
			}, Recipient{})

			err := validate.Struct(tc.recipient)
			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
