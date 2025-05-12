package services

import "thirdparty_service/internal/repositories"

type Service interface {
	VerifyAadhaar(string) (bool, string, string, error)
	VerifyPAN(string) (bool, string, error)
	SendSMS(string, string) (string, error)
	SendEmail(string, string, string) (string, error)
	InitiatePayment(string, float64) (string, string, error)
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

func (s *service) SendEmail(to, subject, body string) (string, error) {
	return s.repo.SendEmail(to, subject, body)
}

func (s *service) InitiatePayment(userID string, amount float64) (string, string, error) {
	return s.repo.InitiatePayment(userID, amount)
}
