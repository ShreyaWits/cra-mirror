package services

import (
	"fmt"
	"thirdparty_service/internal/config"
	"thirdparty_service/internal/modules/execute/dtos"
	"thirdparty_service/internal/modules/execute/repositories"
	"thirdparty_service/pkg/sendgrid"
	twilio_sms "thirdparty_service/pkg/twilio"
	"thirdparty_service/pkg/whatsapp"
)

type Service interface {
	VerifyAadhaar(string) (bool, string, string, error)
	VerifyPAN(string) (bool, string, error)
	SendSMS(string, string) (string, error)
	InitiatePayment(string, float64) (string, string, error)
	SendTwilioSms(payload *dtos.TwilioSmsRequest) error
	SendWhatsAppMessage(payload *dtos.SendWhatsAppMessageRequest) error
	SendEmailBySendGrid(payload *dtos.SendGridEmailRequest) error
}

type service struct {
	repo repositories.Repository
}

func New(repo repositories.Repository) Service {
	return &service{repo: repo}
}

func (s *service) VerifyAadhaar(aadhaar string) (bool, string, string, error) {
	return s.repo.VerifyAadhaar(aadhaar)
}

func (s *service) VerifyPAN(pan string) (bool, string, error) {
	return s.repo.VerifyPAN(pan)
}

func (s *service) SendSMS(phone, message string) (string, error) {
	return s.repo.SendSMS(phone, message)
}
func (s *service) SendTwilioSms(payload *dtos.TwilioSmsRequest) error {
	accountSid := config.AppConfig.TwilioAccountSID
	authToken := config.AppConfig.TwilioAuthToken
	formNumber := config.AppConfig.TwilioFormNumber
	client := twilio_sms.NewTwilioClient(accountSid, authToken)
	phoneNumber := fmt.Sprintf("%s %s", payload.CountryCode, payload.Phone)

	if err := client.SendSMS(phoneNumber, formNumber, payload.Message); err != nil {
		return err
	}
	return nil
}

func (s *service) SendEmailBySendGrid(payload *dtos.SendGridEmailRequest) error {
	apiKey := config.AppConfig.SendGridApiKey
	fromEmail := config.AppConfig.SendGridFromEmail
	name := config.AppConfig.SendGridFromName

	contentType := "text/plain"

	switch payload.Type {
	case "TEXT":
		contentType = "text/plain"
	case "HTML":
		contentType = "text/html"
	}

	// new SendGrid client
	client := sendgrid.NewSendGridClient(apiKey)

	err := client.SendEmail(sendgrid.Email{
		From: sendgrid.Contact{
			Email: fromEmail,
			Name:  name,
		},
		ReplyTo: sendgrid.Contact{
			Email: fromEmail,
			Name:  name,
		},
		Content: []sendgrid.Content{
			{
				Type:  contentType,
				Value: payload.Body,
			},
		},
		Personalizations: []sendgrid.Personalization{
			{
				To:      []sendgrid.Contact{{Email: payload.To}},
				Subject: payload.Subject,
			},
		},
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *service) SendWhatsAppMessage(payload *dtos.SendWhatsAppMessageRequest) error {
	accountSid := config.AppConfig.SendWhatsAppMessageSID
	authToken := config.AppConfig.SendWhatsAppMessageToken
	fromNumber := config.AppConfig.SendWhatsAppMessageFromNumber // Twilio sandbox or registered number

	if accountSid == "" || authToken == "" {
		return fmt.Errorf("twilio credentials are not set in environment")
	}

	client, err := whatsapp.NewWhatsAppClient(accountSid, authToken, fromNumber)
	if err != nil {
		return fmt.Errorf("failed to create WhatsApp client: %w", err)
	}

	to := payload.Phone // `SendWhatsAppMessage` prepends "whatsapp:" internally
	message := payload.Message

	err = client.SendWhatsAppMessage(to, message)
	if err != nil {
		return fmt.Errorf("failed to send WhatsApp message: %w", err)
	}

	return nil
}


func (s *service) InitiatePayment(userID string, amount float64) (string, string, error) {
	return s.repo.InitiatePayment(userID, amount)
}
