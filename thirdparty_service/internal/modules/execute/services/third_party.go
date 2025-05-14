package services

import (
	"fmt"
	"thirdparty_service/internal/config"
	"thirdparty_service/internal/modules/execute/dtos"
	"thirdparty_service/internal/modules/execute/repositories"
	"thirdparty_service/pkg/sendgrid"
	twilio_sms "thirdparty_service/pkg/twilio"
)

type Service interface {
	VerifyAadhaar(string) (bool, string, string, error)
	VerifyPAN(string) (bool, string, error)
	SendSMS(string, string) (string, error)
	InitiatePayment(string, float64) (string, string, error)
	SendTwilioSms(payload *dtos.TwilioSmsRequest) error
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
				Type:  "text/plain",
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

func (s *service) InitiatePayment(userID string, amount float64) (string, string, error) {
	return s.repo.InitiatePayment(userID, amount)
}
