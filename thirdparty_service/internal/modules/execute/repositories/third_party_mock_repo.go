package repositories

import (
	"github.com/stretchr/testify/mock"
)

// MockRepo satisfies the repositories.Repository interface used by the service.
type MockRepo struct {
	mock.Mock
}

func (m *MockRepo) VerifyAadhaar(aadhaar string) (bool, string, string, error) {
	args := m.Called(aadhaar)
	return args.Bool(0), args.String(1), args.String(2), args.Error(3)
}

func (m *MockRepo) VerifyPAN(pan string) (bool, string, error) {
	args := m.Called(pan)
	return args.Bool(0), args.String(1), args.Error(2)
}

func (m *MockRepo) SendSMS(phone, message string) (string, error) {
	args := m.Called(phone, message)
	return args.String(0), args.Error(1)
}

func (m *MockRepo) SendEmail(to, subject, body string) (string, error) {
	args := m.Called(to, subject, body)
	return args.String(0), args.Error(1)
}

func (m *MockRepo) InitiatePayment(userID string, amount float64) (string, string, error) {
	args := m.Called(userID, amount)
	return args.String(0), args.String(1), args.Error(2)
}
