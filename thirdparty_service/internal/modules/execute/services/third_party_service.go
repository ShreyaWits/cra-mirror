package services

import (
	"context"
	"fmt"
	"thirdparty_service/internal/config"
	"thirdparty_service/internal/modules/execute/dtos"
	"thirdparty_service/internal/modules/execute/repositories"
	push_service "thirdparty_service/pkg/push"
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
	SendPushNotification(payload *dtos.PushNotificationRequest) error
}

type service struct {
	repo                      repositories.Repository
	twilioClientConstructor   func(accountSid, authToken string) *twilio_sms.TwilioClient
	sendGridClientConstructor func(apiKey string) *sendgrid.SendGridClient
	whatsAppClientConstructor func(accountSid, authToken, fromNumber string) (*whatsapp.WhatsAppClient, error)
	fcmClientConstructor      func(ctx context.Context, creds, projectID string) (push_service.FCMClient, error)
}

func New(
	repo repositories.Repository,
	twilioClientConstructor func(accountSid, authToken string) *twilio_sms.TwilioClient,
	sendGridClientConstructor func(apiKey string) *sendgrid.SendGridClient,
	whatsAppClientConstructor func(accountSid, authToken, fromNumber string) (*whatsapp.WhatsAppClient, error),
	fcmClientConstructor func(ctx context.Context, creds, projectID string) (push_service.FCMClient, error),
) Service {
	return &service{
		repo:                      repo,
		twilioClientConstructor:   twilioClientConstructor,
		sendGridClientConstructor: sendGridClientConstructor,
		whatsAppClientConstructor: whatsAppClientConstructor,
		fcmClientConstructor:      fcmClientConstructor,
	}
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
	if accountSid == "" || authToken == "" || formNumber == "" {
		return fmt.Errorf("twilio credentials are not set in environment")
	}
	client := s.twilioClientConstructor(accountSid, authToken)
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

	if apiKey == "" {
		return fmt.Errorf("SendGrid API key is not set")
	}
	if fromEmail == "" {
		return fmt.Errorf("SendGrid from email is not set")
	}
	if name == "" {
		return fmt.Errorf("SendGrid from name is not set")
	}

	contentType := "text/plain"

	switch payload.Type {
	case "TEXT":
		contentType = "text/plain"
	case "HTML":
		contentType = "text/html"
	}

	// new SendGrid client
	client := s.sendGridClientConstructor(apiKey)

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

	if accountSid == "" || authToken == "" || fromNumber == "" {
		return fmt.Errorf("twilio credentials are not set in environment")
	}

	client, err := s.whatsAppClientConstructor(accountSid, authToken, fromNumber)
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

func (s *service) SendPushNotification(payload *dtos.PushNotificationRequest) error {
	if payload == nil || payload.ToToken == "" || payload.Title == "" || payload.Body == "" {
		return fmt.Errorf("invalid push notification payload")
	}

	creds := config.AppConfig.PushNotificationAccountCreds
	projectID := config.AppConfig.PushNotificationProjectID

	if creds == "" {
		return fmt.Errorf("FCM account credentials are not set")
	}
	if projectID == "" {
		return fmt.Errorf("FCM project ID is not set")
	}

	ctx := context.Background()
	client, err := s.fcmClientConstructor(ctx, creds, projectID)
	if err != nil {
		return fmt.Errorf("failed to create FCM client: %w", err)
	}

	err = client.SendNotification(ctx, payload.ToToken, payload.Title, payload.Body)
	if err != nil {
		return fmt.Errorf("failed to send push notification: %w", err)
	}

	return nil
}

func (s *service) InitiatePayment(userID string, amount float64) (string, string, error) {
	return s.repo.InitiatePayment(userID, amount)
}
