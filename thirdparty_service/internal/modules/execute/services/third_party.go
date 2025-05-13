package services

import (
	"thirdparty_service/internal/config"
	"thirdparty_service/internal/modules/execute/repositories"
	twilio_sms "thirdparty_service/pkg/twilio"
)

type Service interface {
	VerifyAadhaar(string) (bool, string, string, error)
	VerifyPAN(string) (bool, string, error)
	SendSMS(string, string) (string, error)
	SendEmail(string, string, string) (string, error)
	InitiatePayment(string, float64) (string, string, error)
	SendTwilioSms(string, string) error
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
func (s *service) SendTwilioSms(phone, message string) error {
	accountSid := config.AppConfig.TwilioAccountSID
	authToken := config.AppConfig.TwilioAuthToken
	formNumber := config.AppConfig.TwilioFormNumber
	client := twilio_sms.NewTwilioClient(accountSid, authToken)
	err := client.SendSMS(phone, formNumber, message)
	if err != nil {
		return err
	}
	return nil
}

func (s *service) SendEmail(to, subject, body string) (string, error) {
	return s.repo.SendEmail(to, subject, body)
}

func (s *service) InitiatePayment(userID string, amount float64) (string, string, error) {
	return s.repo.InitiatePayment(userID, amount)
}
