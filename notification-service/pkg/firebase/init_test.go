package firebase

import (
	"context"
	"errors"
	"testing"

	"firebase.google.com/go/v4/messaging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock client
type MockFCMClient struct {
	mock.Mock
}

func (m *MockFCMClient) Send(ctx context.Context, msg ...*messaging.Message) (*messaging.BatchResponse, error) {
	args := m.Called(ctx, msg[0])
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*messaging.BatchResponse), args.Error(1)
}

func TestSendPushNotification_Success(t *testing.T) {
	mockClient := new(MockFCMClient)
	NewFCMClient = func(ctx context.Context) (FCMClient, error) {
		return mockClient, nil
	}

	expectedToken := "test-token"
	expectedTitle := "Hello"
	expectedBody := "Test Body"

	mockClient.On("Send", mock.Anything, mock.MatchedBy(func(msg *messaging.Message) bool {
		return msg.Token == expectedToken &&
			msg.Android.Notification.Title == expectedTitle &&
			msg.Android.Notification.Body == expectedBody
	})).Return(&messaging.BatchResponse{SuccessCount: 1}, nil)

	err := SendPushNotification(expectedToken, expectedTitle, expectedBody)
	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}

func TestSendPushNotification_ClientInitError(t *testing.T) {
	NewFCMClient = func(ctx context.Context) (FCMClient, error) {
		return nil, errors.New("init failed")
	}

	err := SendPushNotification("token", "title", "body")
	assert.EqualError(t, err, "init failed")
}

func TestSendPushNotification_SendFails(t *testing.T) {
	mockClient := new(MockFCMClient)
	NewFCMClient = func(ctx context.Context) (FCMClient, error) {
		return mockClient, nil
	}

	mockClient.On("Send", mock.Anything, mock.Anything).Return((*messaging.BatchResponse)(nil), errors.New("send error"))

	err := SendPushNotification("token", "title", "body")
	assert.EqualError(t, err, "send error")
}
