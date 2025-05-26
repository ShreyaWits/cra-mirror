package services

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"thirdparty_service/internal/config"
	configdto "thirdparty_service/internal/modules/config/dto"
	"thirdparty_service/internal/modules/execute/dtos"
	push_service "thirdparty_service/pkg/push"
	"thirdparty_service/pkg/sendgrid"
	twilio_sms "thirdparty_service/pkg/twilio"
	"thirdparty_service/pkg/whatsapp"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockRepository is a mock implementation of the repositories.Repository interface.
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) VerifyAadhaar(aadhaar string) (bool, string, string, error) {
	args := m.Called(aadhaar)
	return args.Bool(0), args.String(1), args.String(2), args.Error(3)
}

func (m *MockRepository) VerifyPAN(pan string) (bool, string, string, error) {
	args := m.Called(pan)
	return args.Bool(0), args.String(1), args.String(2), args.Error(3)
}

func (m *MockRepository) SendSMS(phone, message string) (string, error) {
	args := m.Called(phone, message)
	return args.String(0), args.Error(1)
}

func (m *MockRepository) InitiatePayment(userID string, amount float64) (string, string, error) {
	args := m.Called(userID, amount)
	return args.String(0), args.String(1), args.Error(2)
}

func (m *MockRepository) SendEmail(to, subject, body string) (string, error) {
	args := m.Called(to, subject, body)
	return args.String(0), args.Error(1)
}

// MockTwilioClient is a mock client for Twilio operations.
type MockTwilioClient struct {
	mock.Mock
}

func (m *MockTwilioClient) SendSMS(to, from, body string) error {
	args := m.Called(to, from, body)
	return args.Error(0)
}

// MockSendGridClient is a mock client for SendGrid operations.
type MockSendGridClient struct {
	mock.Mock
}

func (m *MockSendGridClient) SendEmail(email sendgrid.Email) error {
	args := m.Called(email)
	return args.Error(0)
}

// MockWhatsAppClient is a mock client for WhatsApp operations.
type MockWhatsAppClient struct {
	mock.Mock
}

func (m *MockWhatsAppClient) SendWhatsAppMessage(to, message string) error {
	args := m.Called(to, message)
	return args.Error(0)
}

// MockFCMClient is a mock client for FCM operations.
type MockFCMClient struct {
	mock.Mock
}

func (m *MockFCMClient) SendNotification(ctx context.Context, token, title, body string) error {
	args := m.Called(ctx, token, title, body)
	return args.Error(0)
}

// Helper function to reset config for tests
func resetConfig() {
	config.AppConfig.TwilioAccountSID = ""
	config.AppConfig.TwilioAuthToken = ""
	config.AppConfig.TwilioFormNumber = ""
	config.AppConfig.SendGridApiKey = ""
	config.AppConfig.SendGridFromEmail = ""
	config.AppConfig.SendGridFromName = ""
	config.AppConfig.SendWhatsAppMessageSID = ""
	config.AppConfig.SendWhatsAppMessageToken = ""
	config.AppConfig.SendWhatsAppMessageFromNumber = ""
	config.AppConfig.PushNotificationAccountCreds = ""
	config.AppConfig.PushNotificationProjectID = ""
}

func TestNew(t *testing.T) {
	mockRepo := new(MockRepository)
	// Pass dummy client constructors that return nil of the expected types
	dummyTwilioConstructor := func(accountSid, authToken string) *twilio_sms.TwilioClient { return nil }
	dummySendGridConstructor := func(apiKey string) *sendgrid.SendGridClient { return nil }
	dummyWhatsAppConstructor := func(accountSid, authToken, fromNumber string) (*whatsapp.WhatsAppClient, error) { return nil, nil }
	dummyFCMConstructor := func(ctx context.Context, creds, projectID string) (push_service.FCMClient, error) {
		return push_service.FCMClient{}, nil
	}

	service := NewThirdPartService(
		mockRepo,
		dummyTwilioConstructor,
		dummySendGridConstructor,
		dummyWhatsAppConstructor,
		dummyFCMConstructor,
	)

	assert.NotNil(t, service)
}

func TestVerifyAadhaar(t *testing.T) {
	mockRepo := new(MockRepository)
	// Use the updated New function with dummy client constructors for methods that don't use them
	service := NewThirdPartService(mockRepo, nil, nil, nil, nil)

	expectedBool := true
	expectedString1 := "details1"
	expectedString2 := "details2"
	expectedErr := errors.New("aadhaar verification error")

	mockRepo.On("VerifyAadhaar", "123456789012").Return(expectedBool, expectedString1, expectedString2, nil).Once()
	mockRepo.On("VerifyAadhaar", "invalid").Return(false, "", "", expectedErr).Once()

	// Test success case
	b, s1, s2, err := service.VerifyAadhaar("123456789012")
	assert.Equal(t, expectedBool, b)
	assert.Equal(t, expectedString1, s1)
	assert.Equal(t, expectedString2, s2)
	assert.NoError(t, err)

	// Test error case
	b, s1, s2, err = service.VerifyAadhaar("invalid")
	assert.False(t, b)
	assert.Empty(t, s1)
	assert.Empty(t, s2)
	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)

	mockRepo.AssertExpectations(t)
}

func TestVerifyPAN(t *testing.T) {
	mockRepo := new(MockRepository)
	// Use the updated New function with dummy client constructors for methods that don't use them
	service := NewThirdPartService(mockRepo, nil, nil, nil, nil)

	expectedBool := true
	expectedString := "details"
	expectedErr := errors.New("pan verification error")

	mockRepo.On("VerifyPAN", "ABCDE1234F").Return(expectedBool, expectedString, interface{}(""), nil).Once()
	mockRepo.On("VerifyPAN", "invalid").Return(false, "", interface{}(""), expectedErr).Once()

	// Test success case
	b, s, _, err := service.VerifyPAN("ABCDE1234F")
	assert.Equal(t, expectedBool, b)
	assert.Equal(t, expectedString, s)
	assert.NoError(t, err)

	// Test error case
	b, s, _, err = service.VerifyPAN("invalid")
	assert.False(t, b)
	assert.Empty(t, s)
	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)

	mockRepo.AssertExpectations(t)
}

func TestSendSMS(t *testing.T) {
	mockRepo := new(MockRepository)
	// Use the updated New function with dummy client constructors for methods that don't use them
	service := NewThirdPartService(mockRepo, nil, nil, nil, nil)

	expectedString := "sms_id_123"
	expectedErr := errors.New("send sms error")

	mockRepo.On("SendSMS", "1234567890", "Hello").Return(expectedString, nil).Once()
	mockRepo.On("SendSMS", "invalid", "Error message").Return("", expectedErr).Once()

	// Test success case
	s, err := service.SendSMS("1234567890", "Hello")
	assert.Equal(t, expectedString, s)
	assert.NoError(t, err)

	// Test error case
	s, err = service.SendSMS("invalid", "Error message")
	assert.Empty(t, s)
	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)

	mockRepo.AssertExpectations(t)
}

func TestInitiatePayment(t *testing.T) {
	mockRepo := new(MockRepository)
	// Use the updated New function with dummy client constructors for methods that don't use them
	service := NewThirdPartService(mockRepo, nil, nil, nil, nil)

	expectedString1 := "payment_id_abc"
	expectedString2 := "transaction_id_xyz"
	expectedErr := errors.New("initiate payment error")

	mockRepo.On("InitiatePayment", "user123", 100.50).Return(expectedString1, expectedString2, nil).Once()
	mockRepo.On("InitiatePayment", "user456", 50.00).Return("", "", expectedErr).Once()

	// Test success case
	s1, s2, err := service.InitiatePayment("user123", 100.50)
	assert.Equal(t, expectedString1, s1)
	assert.Equal(t, expectedString2, s2)
	assert.NoError(t, err)

	// Test error case
	s1, s2, err = service.InitiatePayment("user456", 50.00)
	assert.Empty(t, s1)
	assert.Empty(t, s2)
	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)

	mockRepo.AssertExpectations(t)
}

func TestSendTwilioSms(t *testing.T) {
	mockRepo := new(MockRepository) // Not used in this test, but needed for service creation

	// To test the client interaction, we need a mock constructor that returns a mock client
	mockClientForInteraction := new(MockTwilioClient)
	mockTwilioClientConstructorForInteraction := func(accountSid, authToken string) *twilio_sms.TwilioClient {
		// Create a new TwilioClient that uses our mock's SendSMS function
		client := twilio_sms.NewTwilioClient(accountSid, authToken)
		client.HTTPClient = &http.Client{
			Transport: &mockTransport{
				roundTripFunc: func(req *http.Request) (*http.Response, error) {
					// Call the mock's SendSMS function with the request parameters
					err := mockClientForInteraction.SendSMS(req.FormValue("To"), req.FormValue("From"), req.FormValue("Body"))
					if err != nil {
						return &http.Response{StatusCode: 400}, nil
					}
					return &http.Response{StatusCode: 200}, nil
				},
			},
		}
		return client
	}
	serviceForInteraction := NewThirdPartService(mockRepo, mockTwilioClientConstructorForInteraction, nil, nil, nil)

	originalConfig := config.AppConfig
	defer func() { config.AppConfig = originalConfig }()

	// Set dynamic config for the test
	testDynamicConfig := &configdto.ConfigResponse{
		TwilioAccountSID: "test_sid",
		TwilioAuthToken:  "test_token",
		TwilioFormNumber: "test_from",
	}
	config.AppConfig.SetEnv(testDynamicConfig)

	payload := &dtos.TwilioSmsRequest{
		CountryCode: "+1",
		Phone:       "1234567890",
		Message:     "Test message",
	}

	// Test success case with client interaction
	mockClientForInteraction.On("SendSMS", "+1 1234567890", "test_from", "Test message").Return(nil).Once()
	err := serviceForInteraction.SendTwilioSms(payload)
	assert.NoError(t, err)
	mockClientForInteraction.AssertExpectations(t)

	// Test error case with client interaction
	expectedErr := errors.New("twilio send error")
	mockClientForInteraction.On("SendSMS", "+1 1234567890", "test_from", "Test message").Return(expectedErr).Once()
	err = serviceForInteraction.SendTwilioSms(payload)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "twilio API error")
	mockClientForInteraction.AssertExpectations(t)

	// Dummy constructor that should never be called (for missing credentials cases)
	dummyTwilioConstructor := func(accountSid, authToken string) *twilio_sms.TwilioClient {
		panic("should not be called when credentials are missing")
	}
	serviceWithDummy := NewThirdPartService(mockRepo, dummyTwilioConstructor, nil, nil, nil)

	// Test error case - Missing credentials
	resetConfig()
	err = serviceWithDummy.SendTwilioSms(payload)
	assert.Error(t, err)
	assert.Equal(t, "twilio credentials are not set in environment", err.Error())

	// Test error case - Missing account SID
	config.AppConfig.TwilioAuthToken = "test_token"
	config.AppConfig.TwilioFormNumber = "test_from"
	err = serviceWithDummy.SendTwilioSms(payload)
	assert.Error(t, err)
	assert.Equal(t, "twilio credentials are not set in environment", err.Error())

	// Test error case - Missing auth token
	resetConfig()
	config.AppConfig.TwilioAccountSID = "test_sid"
	config.AppConfig.TwilioFormNumber = "test_from"
	err = serviceWithDummy.SendTwilioSms(payload)
	assert.Error(t, err)
	assert.Equal(t, "twilio credentials are not set in environment", err.Error())

	// Test error case - Missing from number
	resetConfig()
	config.AppConfig.TwilioAccountSID = "test_sid"
	config.AppConfig.TwilioAuthToken = "test_token"
	err = serviceWithDummy.SendTwilioSms(payload)
	assert.Error(t, err)
	assert.Equal(t, "twilio credentials are not set in environment", err.Error())
}

// mockTransport implements http.RoundTripper for mocking HTTP requests
type mockTransport struct {
	roundTripFunc func(*http.Request) (*http.Response, error)
}

func (m *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.roundTripFunc(req)
}

func TestSendEmailBySendGrid(t *testing.T) {
	mockRepo := new(MockRepository) // Not used in this test, but needed for service creation

	// To test the client interaction, we need a mock constructor that returns a mock client
	mockClientForInteraction := new(MockSendGridClient)
	mockSendGridClientConstructorForInteraction := func(apiKey string) *sendgrid.SendGridClient {
		// Create a new SendGridClient that uses our mock's SendEmail function
		client := sendgrid.NewSendGridClient(apiKey)
		client.HTTPClient = &http.Client{
			Transport: &mockTransport{
				roundTripFunc: func(req *http.Request) (*http.Response, error) {
					// Parse the request body to get the email
					var email sendgrid.Email
					if err := json.NewDecoder(req.Body).Decode(&email); err != nil {
						return &http.Response{StatusCode: 400}, nil
					}
					// Call the mock's SendEmail function
					err := mockClientForInteraction.SendEmail(email)
					if err != nil {
						return &http.Response{StatusCode: 400}, nil
					}
					return &http.Response{StatusCode: 200}, nil
				},
			},
		}
		return client
	}
	serviceForInteraction := NewThirdPartService(mockRepo, nil, mockSendGridClientConstructorForInteraction, nil, nil)

	// Save and restore original config
	originalConfig := config.AppConfig
	defer func() { config.AppConfig = originalConfig }()

	config.AppConfig.SendGridApiKey = "test_key"
	config.AppConfig.SendGridFromEmail = "from@example.com"
	config.AppConfig.SendGridFromName = "Test Sender"

	payloadText := &dtos.SendGridEmailRequest{
		To:      "to@example.com",
		Subject: "Test Subject Text",
		Body:    "Test Body Text",
		Type:    "TEXT",
	}

	// Test success case - TEXT type with client interaction
	expectedEmailText := sendgrid.Email{
		From:    sendgrid.Contact{Email: "from@example.com", Name: "Test Sender"},
		ReplyTo: sendgrid.Contact{Email: "from@example.com", Name: "Test Sender"},
		Content: []sendgrid.Content{{Type: "text/plain", Value: "Test Body Text"}},
		Personalizations: []sendgrid.Personalization{{
			To: []sendgrid.Contact{{Email: "to@example.com"}}, Subject: "Test Subject Text",
		}},
	}
	mockClientForInteraction.On("SendEmail", expectedEmailText).Return(nil).Once()
	err := serviceForInteraction.SendEmailBySendGrid(payloadText)
	assert.NoError(t, err)
	mockClientForInteraction.AssertExpectations(t)

	// Test success case - HTML type with client interaction
	payloadHTML := &dtos.SendGridEmailRequest{
		To:      "to@example.com",
		Subject: "Test Subject HTML",
		Body:    "<h1>Test Body HTML</h1>",
		Type:    "HTML",
	}
	expectedEmailHTML := sendgrid.Email{
		From:    sendgrid.Contact{Email: "from@example.com", Name: "Test Sender"},
		ReplyTo: sendgrid.Contact{Email: "from@example.com", Name: "Test Sender"},
		Content: []sendgrid.Content{{Type: "text/html", Value: "<h1>Test Body HTML</h1>"}},
		Personalizations: []sendgrid.Personalization{{
			To: []sendgrid.Contact{{Email: "to@example.com"}}, Subject: "Test Subject HTML",
		}},
	}
	mockClientForInteraction.On("SendEmail", expectedEmailHTML).Return(nil).Once()
	err = serviceForInteraction.SendEmailBySendGrid(payloadHTML)
	assert.NoError(t, err)
	mockClientForInteraction.AssertExpectations(t)


	// Test error case with client interaction
	expectedErr := errors.New("sendgrid send error")
	mockClientForInteraction.On("SendEmail", mock.AnythingOfType("sendgrid.Email")).Return(expectedErr).Once()
	err = serviceForInteraction.SendEmailBySendGrid(payloadText)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "sendgrid API error")

	// Dummy constructor that should never be called (for missing credentials cases)
	dummySendGridConstructor := func(apiKey string) *sendgrid.SendGridClient {
		panic("should not be called when credentials are missing")
	}
	serviceWithDummy := NewThirdPartService(mockRepo, nil, dummySendGridConstructor, nil, nil)

	// Test error case - Missing credentials
	// Ensure config is clean for this test
	config.AppConfig.SendGridApiKey = ""
	config.AppConfig.SendGridFromEmail = ""
	config.AppConfig.SendGridFromName = ""
	err = serviceWithDummy.SendEmailBySendGrid(payloadText)
	assert.Error(t, err)
	assert.Equal(t, "SendGrid API key is not set", err.Error())

	// Test error case - Missing from email
	config.AppConfig.SendGridApiKey = "test_key"
	config.AppConfig.SendGridFromName = "Test Sender"
	err = serviceWithDummy.SendEmailBySendGrid(payloadText)
	assert.Error(t, err)
	assert.Equal(t, "SendGrid from email is not set", err.Error())

	// Test error case - Missing from name
	resetConfig()
	config.AppConfig.SendGridApiKey = "test_key"
	config.AppConfig.SendGridFromEmail = "from@example.com"
	err = serviceWithDummy.SendEmailBySendGrid(payloadText)
	assert.Error(t, err)
	assert.Equal(t, "SendGrid from name is not set", err.Error())
}

func TestSendWhatsAppMessage(t *testing.T) {
	mockRepo := new(MockRepository) // Not used in this test, but needed for service creation

	// To test the client interaction, we need a mock constructor that returns a mock client
	mockClientForInteraction := new(MockWhatsAppClient)
	mockWhatsAppClientConstructorForInteraction := func(accountSid, authToken, fromNumber string) (*whatsapp.WhatsAppClient, error) {
		// Create a new WhatsAppClient that uses our mock's SendWhatsAppMessage function
		client, err := whatsapp.NewWhatsAppClient(accountSid, authToken, fromNumber)
		if err != nil {
			return nil, err
		}
		// Replace the default HTTP client with our mock
		http.DefaultClient = &http.Client{
			Transport: &mockTransport{
				roundTripFunc: func(req *http.Request) (*http.Response, error) {
					// Call the mock's SendWhatsAppMessage function with the request parameters
					err := mockClientForInteraction.SendWhatsAppMessage(req.FormValue("To"), req.FormValue("Body"))
					if err != nil {
						return &http.Response{StatusCode: 400}, nil
					}
					return &http.Response{StatusCode: 200}, nil
				},
			},
		}
		return client, nil
	}
	serviceForInteraction := NewThirdPartService(mockRepo, nil, nil, mockWhatsAppClientConstructorForInteraction, nil)

	// Save and restore original config
	originalConfig := config.AppConfig
	defer func() { config.AppConfig = originalConfig }()

	config.AppConfig.SendWhatsAppMessageSID = "test_sid"
	config.AppConfig.SendWhatsAppMessageToken = "test_token"
	config.AppConfig.SendWhatsAppMessageFromNumber = "test_from"

	payload := &dtos.SendWhatsAppMessageRequest{
		Phone:   "1234567890",
		Message: "Test WhatsApp message",
	}

	// Test success case with client interaction
	mockClientForInteraction.On("SendWhatsAppMessage", "whatsapp:1234567890", "Test WhatsApp message").Return(nil).Once()
	err := serviceForInteraction.SendWhatsAppMessage(payload)
	assert.NoError(t, err)
	mockClientForInteraction.AssertExpectations(t)

	// Test error case - SendWhatsAppMessage error with client interaction
	expectedSendErr := errors.New("whatsapp send error")
	mockClientForInteraction.On("SendWhatsAppMessage", "whatsapp:1234567890", "Test WhatsApp message").Return(expectedSendErr).Once()
	err = serviceForInteraction.SendWhatsAppMessage(payload)
	assert.Error(t, err)
	assert.Equal(t, "failed to send WhatsApp message: failed to send WhatsApp message", err.Error())
	mockClientForInteraction.AssertExpectations(t)

	// Test error case - NewWhatsAppClient error
	mockWhatsAppClientConstructorWithError := func(accountSid, authToken, fromNumber string) (*whatsapp.WhatsAppClient, error) {
		return nil, errors.New("client creation error")
	}
	serviceWithError := NewThirdPartService(mockRepo, nil, nil, mockWhatsAppClientConstructorWithError, nil)
	err = serviceWithError.SendWhatsAppMessage(payload)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create WhatsApp client")
	assert.Contains(t, err.Error(), "client creation error")

	// Test error case - Missing credentials
	resetConfig()
	err = serviceWithError.SendWhatsAppMessage(payload)
	assert.Error(t, err)
	assert.Equal(t, "twilio credentials are not set in environment", err.Error())

	// Test error case - Missing from number
	resetConfig()
	config.AppConfig.SendWhatsAppMessageSID = "test_sid"
	config.AppConfig.SendWhatsAppMessageToken = "test_token"
	err = serviceWithError.SendWhatsAppMessage(payload)
	assert.Error(t, err)
	assert.Equal(t, "twilio credentials are not set in environment", err.Error())
}

func TestSendPushNotification(t *testing.T) {
	mockRepo := new(MockRepository) // Not used in this test, but needed for service creation

	mockClient := new(MockFCMClient)
	// Create a mock constructor function that returns the mock client interface
	mockFCMClientConstructor := func(ctx context.Context, creds, projectID string) (push_service.FCMClient, error) {
		client := push_service.NewFCMClientWithHTTPClient(projectID, &http.Client{
			Transport: &mockTransport{
				roundTripFunc: func(req *http.Request) (*http.Response, error) {
					var payload map[string]interface{}
					if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
						return &http.Response{StatusCode: 400}, nil
					}
					message := payload["message"].(map[string]interface{})
					token := message["token"].(string)
					notification := message["notification"].(map[string]interface{})
					title := notification["title"].(string)
					body := notification["body"].(string)
					err := mockClient.SendNotification(req.Context(), token, title, body)
					if err != nil {
						return &http.Response{StatusCode: 400}, nil
					}
					return &http.Response{StatusCode: 200}, nil
				},
			},
		}, "fake-token")
		return *client, nil
	}

	// Use the updated New function with the mock client constructor
	service := NewThirdPartService(mockRepo, nil, nil, nil, mockFCMClientConstructor)

	// Save and restore original config
	originalConfig := config.AppConfig
	defer func() { config.AppConfig = originalConfig }()

	config.AppConfig.PushNotificationAccountCreds = "test_creds"
	config.AppConfig.PushNotificationProjectID = "test_project"

	payload := &dtos.PushNotificationRequest{
		ToToken: "test_token",
		Title:   "Test Title",
		Body:    "Test Body",
	}

	// Test success case
	mockClient.On("SendNotification", mock.Anything, "test_token", "Test Title", "Test Body").Return(nil).Once()
	err := service.SendPushNotification(payload)
	assert.NoError(t, err)
	mockClient.AssertExpectations(t)

	// Test error case - SendNotification error
	expectedSendErr := errors.New("fcm send error")
	mockClient.On("SendNotification", mock.Anything, "test_token", "Test Title", "Test Body").Return(expectedSendErr).Once()
	err = service.SendPushNotification(payload)
	assert.Error(t, err)
	assert.Equal(t, "failed to send push notification: FCM error: status 400 - ", err.Error())
	mockClient.AssertExpectations(t)

	// Test error case - NewFCMClient error
	mockFCMClientConstructorWithError := func(ctx context.Context, creds, projectID string) (push_service.FCMClient, error) {
		return push_service.FCMClient{}, errors.New("client creation error")
	}
	serviceWithError := NewThirdPartService(mockRepo, nil, nil, nil, mockFCMClientConstructorWithError)
	err = serviceWithError.SendPushNotification(payload)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create FCM client")
	assert.Contains(t, err.Error(), "client creation error")

	// Test error case - Missing credentials (creds)
	resetConfig()
	config.AppConfig.PushNotificationProjectID = "test_project" // Keep project ID set
	err = serviceWithError.SendPushNotification(payload)        // Use serviceWithError which has the error constructor
	assert.Error(t, err)
	assert.Equal(t, "FCM account credentials are not set", err.Error())

	// Test error case - Missing credentials (projectID)
	resetConfig()
	config.AppConfig.PushNotificationAccountCreds = "test_creds" // Keep creds set
	err = serviceWithError.SendPushNotification(payload)         // Use serviceWithError
	assert.Error(t, err)
	assert.Equal(t, "FCM project ID is not set", err.Error())
}
