package dtos

import (
	"regexp"
	"time"

	"github.com/go-playground/validator/v10"
)

// Allowed values for channels
var validChannels = map[string]bool{
	"email":    true,
	"sms":      true,
	"push":     true,
	"whatsapp": true,
}

var templateDataRequirements = map[string][]string{
	"otp_verification": {"username", "otp"},
	"nps_feedback":     {"username", "orderID", "nps_score", "feedback"},
	"order_placed":     {"username", "otp"},
}

func validateTemplateID(fl validator.FieldLevel) bool {
	templateID := fl.Field().String()
	_, exists := templateDataRequirements[templateID]
	return exists
}

func validatePhone91(fl validator.FieldLevel) bool {
	phone := fl.Field().String()
	regex := regexp.MustCompile(`^\+91\d{10}$`)
	return regex.MatchString(phone)
}

func validateChannels(fl validator.FieldLevel) bool {
	channels, ok := fl.Field().Interface().([]string)
	if !ok {
		return false
	}
	for _, channel := range channels {
		if !validChannels[channel] {
			return false
		}
	}
	return true
}

func validateRecipient(sl validator.StructLevel) {
	recipient := sl.Current().Interface().(Recipient)
	notification := sl.Parent().Interface().(NotificationRequest)

	if recipient.UserID == "" && recipient.Email == "" && recipient.Phone == "" && recipient.WhatsappNumber == "" {
		sl.ReportError(recipient.UserID, "UserID", "userID", "contact_required", "")
	}

	requiredFields := templateDataRequirements[notification.TemplateID]
	for _, field := range requiredFields {
		if _, ok := recipient.Data[field]; !ok {
			sl.ReportError(recipient.Data, "Data."+field, field, field+"_required", "")
		}
	}
}

func validateAtLeastOneContact(fl validator.StructLevel) {
	recipient := fl.Current().Interface().(Recipient)
	if recipient.UserID == "" && recipient.Email == "" && recipient.Phone == "" && recipient.WhatsappNumber == "" {
		fl.ReportError(recipient.UserID, "UserID", "userID", "contact_required", "")
	}
}

func RegisterValidations(validate *validator.Validate) {
	validate.RegisterValidation("phone", validatePhone91)
	validate.RegisterValidation("channel_list", validateChannels)
	validate.RegisterValidation("template_id", validateTemplateID)
	validate.RegisterStructValidation(validateAtLeastOneContact, Recipient{})
	validate.RegisterStructValidation(validateRecipient, Recipient{})
}

// DTOs

type Recipient struct {
	UserID         string            `json:"userID,omitempty"`
	Email          string            `json:"email,omitempty" validate:"omitempty,email"`
	Phone          string            `json:"phone,omitempty" validate:"omitempty,phone"`
	WhatsappNumber string            `json:"whatsapp_number,omitempty"`
	Data           map[string]string `json:"data" validate:"required,min=1"`
}

type NotificationRequest struct {
	TrackingID string            `json:"trackingId" validate:"required"`
	TemplateID string            `json:"templateID" validate:"required,template_id"`
	Channels   []string          `json:"channels" validate:"required,min=1,channel_list"`
	Meta       map[string]string `json:"meta,omitempty"`
	Tags       []string          `json:"tags,omitempty"`
	Recipients []Recipient       `json:"recipients" validate:"required,min=1,dive"`
}

type NotificationResponse struct {
	NotificationID string    `json:"notificationId"`
	RecipientEmail string    `json:"recipientEmail,omitempty"`
	RecipientPhone string    `json:"recipientPhone,omitempty"`
	Provider       string    `json:"provider"`
	Timestamp      time.Time `json:"timestamp"`
}

type NotificationByIdResponse struct {
	NotificationID string    `json:"notificationId"`
	RecipientEmail string    `json:"recipientEmail,omitempty"`
	RecipientPhone string    `json:"recipientPhone,omitempty"`
	MsgSts         string    `json:"msgSts"`
	EmailSts       string    `json:"emailSts"`
	Provider       string    `json:"provider"`
	Timestamp      time.Time `json:"timestamp"`
}

type NotificationTemporalPayload struct {
	TrackingID       string            `json:"trackingId"`
	Recipient        Recipient         `json:"recipient"`
	Channel          string            `json:"channel"`
	PrimaryService   string            `json:"primaryService"`
	FallbackService  string            `json:"fallbackService"`
	Meta             map[string]string `json:"meta"`
	Tags             []string          `json:"tags,omitempty"`
	TemplateID       string            `json:"templateID"`
	RetryCount       int               `json:"retryCount"`
	ExecutionTimeout int               `json:"executionTimeout"`
}
