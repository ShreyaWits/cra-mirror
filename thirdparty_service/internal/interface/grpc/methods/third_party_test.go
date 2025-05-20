package methods

import (
	"context"
	"errors" // Added errors import
	"testing"

	"thirdparty_service/internal/modules/execute/dtos"
	"thirdparty_service/pkg/codes"
	protos "thirdparty_service/proto"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockService is a mock implementation of services.Service
type MockService struct {
	mock.Mock
}

func (m *MockService) VerifyAadhaar(aadhaar string) (bool, string, string, error) {
	args := m.Called(aadhaar)
	return args.Bool(0), args.String(1), args.String(2), args.Error(3)
}

// Corrected VerifyPAN signature to match service and repository
func (m *MockService) VerifyPAN(pan string) (bool, string, string, error) {
	args := m.Called(pan)
	return args.Bool(0), args.String(1), args.String(2), args.Error(3)
}

func (m *MockService) SendSMS(phone, message string) (string, error) {
	args := m.Called(phone, message)
	return args.String(0), args.Error(1)
}

func (m *MockService) InitiatePayment(userID string, amount float64) (string, string, error) {
	args := m.Called(userID, amount)
	return args.String(0), args.String(1), args.Error(2)
}

func (m *MockService) SendTwilioSms(payload *dtos.TwilioSmsRequest) error {
	args := m.Called(payload)
	return args.Error(0)
}

func (m *MockService) SendWhatsAppMessage(payload *dtos.SendWhatsAppMessageRequest) error {
	args := m.Called(payload)
	return args.Error(0)
}

func (m *MockService) SendEmailBySendGrid(payload *dtos.SendGridEmailRequest) error {
	args := m.Called(payload)
	return args.Error(0)
}

func (m *MockService) SendPushNotification(payload *dtos.PushNotificationRequest) error {
	args := m.Called(payload)
	return args.Error(0)
}

func TestInvokeTwilioSms(t *testing.T) {
	mockSvc := new(MockService)
	server := InitMethods(mockSvc)

	tests := []struct {
		name          string
		request       *protos.InvokeTwilioRequest
		mockSetup     func()
		expectedCode  string
		expectedMsg   string
		expectedError bool
	}{
		{
			name: "successful SMS send",
			request: &protos.InvokeTwilioRequest{
				Phone:       "+11234567890", // E.164 format
				Message:     "Test message",
				CountryCode: "+1",
			},
			mockSetup: func() {
				mockSvc.On("SendTwilioSms", mock.AnythingOfType("*dtos.TwilioSmsRequest")).Return(nil)
			},
			expectedCode:  codes.TS0001,
			expectedMsg:   codes.SuccessMessage(codes.TS0001),
			expectedError: false,
		},
		{
			name: "invalid phone number",
			request: &protos.InvokeTwilioRequest{
				Phone:       "", // Invalid phone
				Message:     "Test message",
				CountryCode: "+1",
			},
			mockSetup:     func() {},
			expectedCode:  codes.TS1001,
			expectedMsg:   codes.ErrorMessage(codes.TS1001),
			expectedError: false,
		},
		{
			name: "service error",
			request: &protos.InvokeTwilioRequest{
				Phone:       "+11234567890",
				Message:     "Test message",
				CountryCode: "+1",
			},
			mockSetup: func() {
				mockSvc.ExpectedCalls = nil // Reset previous calls
				mockSvc.On("SendTwilioSms", mock.AnythingOfType("*dtos.TwilioSmsRequest")).Return(errors.New("mock service error"))
			},
			expectedCode:  codes.TS1008,
			expectedMsg:   codes.ErrorMessage(codes.TS1008),
			expectedError: false,
		},
		{
			name: "invalid country code",
			request: &protos.InvokeTwilioRequest{
				Phone:       "+11234567890",
				Message:     "Test message",
				CountryCode: "", // Invalid country code
			},
			mockSetup:     func() {},
			expectedCode:  codes.TS1003,
			expectedMsg:   codes.ErrorMessage(codes.TS1003),
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			response, err := server.InvokeTwilioSms(context.Background(), tt.request)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedCode, response.Code)
			assert.Equal(t, tt.expectedMsg, response.Message)
		})
	}
}

func TestInvokeSendGridEmail(t *testing.T) {
	mockSvc := new(MockService)
	server := InitMethods(mockSvc)

	tests := []struct {
		name          string
		request       *protos.InvokeSendGridRequest
		mockSetup     func()
		expectedCode  string
		expectedMsg   string
		expectedError bool
	}{
		{
			name: "successful email send",
			request: &protos.InvokeSendGridRequest{
				Subject: "subject@example.com", // Valid email for Subject
				To:      "test@example.com",
				Body:    "Test email body",
				Type:    "TEXT",
			},
			mockSetup: func() {
				mockSvc.On("SendEmailBySendGrid", mock.AnythingOfType("*dtos.SendGridEmailRequest")).Return(nil)
			},
			expectedCode:  codes.TS0001,
			expectedMsg:   codes.SuccessMessage(codes.TS0001),
			expectedError: false,
		},
		{
			name: "invalid email",
			request: &protos.InvokeSendGridRequest{
				Subject: "subject@example.com", // Valid email for Subject
				To:      "invalid-email",
				Body:    "Test email body",
				Type:    "TEXT",
			},
			mockSetup:     func() {},
			expectedCode:  codes.TS1005,
			expectedMsg:   codes.ErrorMessage(codes.TS1005),
			expectedError: false,
		},
		{
			name: "service error",
			request: &protos.InvokeSendGridRequest{
				Subject: "subject@example.com",
				To:      "test@example.com",
				Body:    "Test email body",
				Type:    "TEXT",
			},
			mockSetup: func() {
				mockSvc.ExpectedCalls = nil
				mockSvc.On("SendEmailBySendGrid", mock.AnythingOfType("*dtos.SendGridEmailRequest")).Return(errors.New("mock service error"))
			},
			expectedCode:  codes.TS1007,
			expectedMsg:   codes.ErrorMessage(codes.TS1007),
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			response, err := server.InvokeSendGridEmail(context.Background(), tt.request)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedCode, response.Code)
			assert.Equal(t, tt.expectedMsg, response.Message)
		})
	}
}

func TestInvokeWhatsAppMessage(t *testing.T) {
	mockSvc := new(MockService)
	server := InitMethods(mockSvc)

	tests := []struct {
		name          string
		request       *protos.InvokeWhatsAppRequest
		mockSetup     func()
		expectedCode  string
		expectedMsg   string
		expectedError bool
	}{
		{
			name: "successful WhatsApp message",
			request: &protos.InvokeWhatsAppRequest{
				Phone:       "+11234567890", // E.164 format
				Message:     "Test message",
				CountryCode: "US", // 2 characters without plus sign
			},
			mockSetup: func() {
				mockSvc.On("SendWhatsAppMessage", mock.AnythingOfType("*dtos.SendWhatsAppMessageRequest")).Return(nil)
			},
			expectedCode:  codes.TS0001,
			expectedMsg:   codes.SuccessMessage(codes.TS0001),
			expectedError: false,
		},
		{
			name: "invalid phone number",
			request: &protos.InvokeWhatsAppRequest{
				Phone:       "", // Invalid phone
				Message:     "Test message",
				CountryCode: "US",
			},
			mockSetup:     func() {},
			expectedCode:  "VALIDATION_ERROR",
			expectedMsg:   "",
			expectedError: false,
		},
		{
			name: "service error",
			request: &protos.InvokeWhatsAppRequest{
				Phone:       "+11234567890",
				Message:     "Test message",
				CountryCode: "US",
			},
			mockSetup: func() {
				mockSvc.ExpectedCalls = nil
				mockSvc.On("SendWhatsAppMessage", mock.AnythingOfType("*dtos.SendWhatsAppMessageRequest")).Return(errors.New("mock service error"))
			},
			expectedCode:  codes.TS1012,
			expectedMsg:   codes.ErrorMessage(codes.TS1012),
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			response, err := server.InvokeWhatsAppMessage(context.Background(), tt.request)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedCode, response.Code)
			assert.Equal(t, tt.expectedMsg, response.Message)
		})
	}
}

func TestInvokePushNotification(t *testing.T) {
	mockSvc := new(MockService)
	server := InitMethods(mockSvc)

	tests := []struct {
		name          string
		request       *protos.PushNotificationRequest
		mockSetup     func()
		expectedCode  string
		expectedMsg   string
		expectedError bool
	}{
		{
			name: "successful push notification",
			request: &protos.PushNotificationRequest{
				ToToken: "test_token",
				Title:   "Test Title",
				Body:    "Test Body",
			},
			mockSetup: func() {
				mockSvc.On("SendPushNotification", mock.AnythingOfType("*dtos.PushNotificationRequest")).Return(nil)
			},
			expectedCode:  codes.TS0001,
			expectedMsg:   codes.SuccessMessage(codes.TS0001),
			expectedError: false,
		},
		{
			name: "invalid token",
			request: &protos.PushNotificationRequest{
				ToToken: "", // Invalid token
				Title:   "Test Title",
				Body:    "Test Body",
			},
			mockSetup:     func() {},
			expectedCode:  codes.TS0001,
			expectedMsg:   codes.SuccessMessage(codes.TS0001),
			expectedError: false,
		},
		{
			name: "service error",
			request: &protos.PushNotificationRequest{
				ToToken: "test_token",
				Title:   "Test Title",
				Body:    "Test Body",
			},
			mockSetup: func() {
				mockSvc.ExpectedCalls = nil
				mockSvc.On("SendPushNotification", mock.AnythingOfType("*dtos.PushNotificationRequest")).Return(errors.New("mock service error"))
			},
			expectedCode:  codes.TS1013,
			expectedMsg:   codes.ErrorMessage(codes.TS1013),
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			response, err := server.InvokePushNotification(context.Background(), tt.request)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedCode, response.Status)
			assert.Equal(t, tt.expectedMsg, response.Message)
		})
	}
}

func TestVerifyAadhaar(t *testing.T) {
	mockSvc := new(MockService)
	server := InitMethods(mockSvc)

	tests := []struct {
		name             string
		request          *protos.AadharVerifyRequest
		mockSetup        func()
		expectedCode     string
		expectedMsg      string
		expectedVerified bool
		expectedData     *protos.AadharResult
	}{
		{
			name: "successful verification",
			request: &protos.AadharVerifyRequest{
				Number: "123412341234",
			},
			mockSetup: func() {
				mockSvc.On("VerifyAadhaar", "123412341234").Return(true, "Rajesh Kumar", "1985-04-12", nil)
			},
			expectedCode:     codes.TS0001,
			expectedMsg:      codes.SuccessMessage(codes.TS0001),
			expectedVerified: true,
			expectedData: &protos.AadharResult{
				Name: "Rajesh Kumar",
				Dob:  "1985-04-12",
			},
		},
		{
			name: "invalid aadhaar number",
			request: &protos.AadharVerifyRequest{
				Number: "", // Invalid aadhaar
			},
			mockSetup:        func() {},
			expectedCode:     codes.TS1010,
			expectedMsg:      codes.ErrorMessage(codes.TS1010),
			expectedVerified: false,
			expectedData:     nil,
		},
		{
			name: "verification failed - unknown aadhaar",
			request: &protos.AadharVerifyRequest{
				Number: "999999999999",
			},
			mockSetup: func() {
				mockSvc.On("VerifyAadhaar", "999999999999").Return(false, "", "", nil)
			},
			expectedCode:     codes.TS1010,
			expectedMsg:      codes.ErrorMessage(codes.TS1010),
			expectedVerified: false,
			expectedData:     nil,
		},
		{
			name: "service error",
			request: &protos.AadharVerifyRequest{
				Number: "123412341234",
			},
			mockSetup: func() {
				mockSvc.ExpectedCalls = nil
				mockSvc.On("VerifyAadhaar", "123412341234").Return(false, "", "", errors.New("mock service error"))
			},
			expectedCode:     codes.TS1010,
			expectedMsg:      codes.ErrorMessage(codes.TS1010),
			expectedVerified: false,
			expectedData:     nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			response, err := server.VerifyAadhaar(context.Background(), tt.request)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedCode, response.Code)
			assert.Equal(t, tt.expectedMsg, response.Message)
			assert.Equal(t, tt.expectedVerified, response.Verified)
			if tt.expectedData != nil {
				assert.NotNil(t, response.Data)
				assert.Equal(t, tt.expectedData.Name, response.Data.Name)
				assert.Equal(t, tt.expectedData.Dob, response.Data.Dob)
			} else {
				assert.Nil(t, response.Data)
			}
		})
	}
}

func TestVerifyPAN(t *testing.T) {
	mockSvc := new(MockService)
	server := InitMethods(mockSvc)

	tests := []struct {
		name             string
		request          *protos.PANVerifyRequest
		mockSetup        func()
		expectedCode     string
		expectedMsg      string
		expectedVerified bool
		expectedData     *protos.PanResult
	}{
		{
			name: "successful verification",
			request: &protos.PANVerifyRequest{
				Number: "ABCDE1234F",
			},
			mockSetup: func() {
				mockSvc.On("VerifyPAN", "ABCDE1234F").Return(true, "Rajesh Kumar", "Individual", nil)
			},
			expectedCode:     codes.TS0001,
			expectedMsg:      codes.SuccessMessage(codes.TS0001),
			expectedVerified: true,
			expectedData: &protos.PanResult{
				Name:    "Rajesh Kumar",
				Dob:     "", // The service layer doesn't return Dob for PAN
				PanType: "Individual",
			},
		},
		{
			name: "invalid PAN number",
			request: &protos.PANVerifyRequest{
				Number: "", // Invalid PAN
			},
			mockSetup:        func() {},
			expectedCode:     codes.TS1011,
			expectedMsg:      codes.ErrorMessage(codes.TS1011),
			expectedVerified: false,
			expectedData:     nil,
		},
		{
			name: "verification failed - unknown PAN",
			request: &protos.PANVerifyRequest{
				Number: "XXXXX9999X",
			},
			mockSetup: func() {
				mockSvc.On("VerifyPAN", "XXXXX9999X").Return(false, "", "", nil)
			},
			expectedCode:     codes.TS1011,
			expectedMsg:      codes.ErrorMessage(codes.TS1011),
			expectedVerified: false,
			expectedData:     nil,
		},
		{
			name: "service error",
			request: &protos.PANVerifyRequest{
				Number: "ABCDE1234F",
			},
			mockSetup: func() {
				mockSvc.ExpectedCalls = nil
				mockSvc.On("VerifyPAN", "ABCDE1234F").Return(false, "", "", errors.New("mock service error"))
			},
			expectedCode:     codes.TS1011,
			expectedMsg:      codes.ErrorMessage(codes.TS1011),
			expectedVerified: false,
			expectedData:     nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			response, err := server.VerifyPAN(context.Background(), tt.request)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedCode, response.Code)
			assert.Equal(t, tt.expectedMsg, response.Message)
			assert.Equal(t, tt.expectedVerified, response.Verified)
			if tt.expectedData != nil {
				assert.NotNil(t, response.Data)
				assert.Equal(t, tt.expectedData.Name, response.Data.Name)
				assert.Equal(t, tt.expectedData.Dob, response.Data.Dob)
				assert.Equal(t, tt.expectedData.PanType, response.Data.PanType)
			} else {
				assert.Nil(t, response.Data)
			}
		})
	}
}
