package sendgrid

import (
	"errors"
	"net/http"
	"testing"

	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockSendGridClient is a mock implementation of the SendGridClient interface
type MockSendGridClient struct {
	mock.Mock
}

func (m *MockSendGridClient) Send(email *mail.SGMailV3) (*SendGridResponse, error) {
	args := m.Called(email)
	response, ok := args.Get(0).(*SendGridResponse)
	if !ok && args.Get(0) != nil {
		return nil, errors.New("mock returned unexpected type")
	}
	return response, args.Error(1)
}

func TestInitSendGrid(t *testing.T) {
	apiKey := "test-api-key"
	fromEmail := "sender@example.com"
	recipient := "recipient@example.com"
	subject := "Test Subject"
	message := "Test Message"

	payload := InitSendGrid(apiKey, fromEmail, recipient, subject, message)

	assert.NotNil(t, payload)
	assert.Equal(t, fromEmail, payload.FromEmail)
	assert.Equal(t, recipient, payload.Recipient)
	assert.Equal(t, subject, payload.Subject)
	assert.Equal(t, message, payload.Message)
	assert.NotNil(t, payload.EmailClient)
}

func TestEmailPayload_SendEmail_Success(t *testing.T) {
	mockClient := new(MockSendGridClient)
	payload := &EmailPayload{
		FromEmail:   "sender@example.com",
		Recipient:   "recipient@example.com",
		Subject:     "Test Subject",
		Message:     "Test Message",
		EmailClient: mockClient,
	}

	mockResponse := &SendGridResponse{StatusCode: http.StatusOK}
	mockClient.On("Send", mock.AnythingOfType("*mail.SGMailV3")).Return(mockResponse, nil)

	success, err := payload.SendEmail()

	assert.True(t, success)
	assert.Nil(t, err)
	mockClient.AssertExpectations(t)
}

func TestEmailPayload_SendEmail_ClientError(t *testing.T) {
	mockClient := new(MockSendGridClient)
	payload := &EmailPayload{
		FromEmail:   "sender@example.com",
		Recipient:   "recipient@example.com",
		Subject:     "Test Subject",
		Message:     "Test Message",
		EmailClient: mockClient,
	}

	expectedErr := errors.New("sendgrid client error")
	mockClient.On("Send", mock.AnythingOfType("*mail.SGMailV3")).Return((*SendGridResponse)(nil), expectedErr)

	success, err := payload.SendEmail()

	assert.False(t, success)
	assert.EqualError(t, err, expectedErr.Error())
	mockClient.AssertExpectations(t)
}

func TestEmailPayload_SendEmail_APIError(t *testing.T) {
	mockClient := new(MockSendGridClient)
	payload := &EmailPayload{
		FromEmail:   "sender@example.com",
		Recipient:   "recipient@example.com",
		Subject:     "Test Subject",
		Message:     "Test Message",
		EmailClient: mockClient,
	}

	mockResponse := &SendGridResponse{StatusCode: http.StatusBadRequest}
	mockClient.On("Send", mock.AnythingOfType("*mail.SGMailV3")).Return(mockResponse, nil)

	success, err := payload.SendEmail()

	assert.False(t, success)
	assert.EqualError(t, err, "failed to send email, status code: 400")
	mockClient.AssertExpectations(t)
}
