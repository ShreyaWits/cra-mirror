package push_service

import (
	"context"
	"testing"
	"fmt"
	"firebase.google.com/go/v4/messaging"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/assert"
)

type mockFCMClient struct {
	mock.Mock
}

func (m *mockFCMClient) Send(ctx context.Context, msg *messaging.Message) (string, error) {
	args := m.Called(ctx, msg)
	return args.String(0), args.Error(1)
}

func TestSendPushNotification(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() (FCMClient, string, string, string, string)
		wantErr  bool
	}{
		{
			name: "successful notification",
			setup: func() (FCMClient, string, string, string, string) {
				mockClient := &mockFCMClient{}
				mockClient.On("Send", context.Background(), mock.AnythingOfType("*messaging.Message")).
					Return("test-message-id", nil).
					Run(func(args mock.Arguments) {
						msg := args.Get(1).(*messaging.Message)
						assert.Equal(t, "test-token", msg.Token)
						assert.Equal(t, "Test Title", msg.Android.Notification.Title)
					})
				return mockClient, "valid-server-key", "test-token", "Test Title", "Test Body"
			},
			wantErr: false,
		},
		{
			name: "invalid token",
			setup: func() (FCMClient, string, string, string, string) {
				mockClient := &mockFCMClient{}
				mockClient.On("Send", context.Background(), mock.AnythingOfType("*messaging.Message")).
					Return("", fmt.Errorf("invalid token error")). // Return an error for invalid token
					Run(func(args mock.Arguments) {
						msg := args.Get(1).(*messaging.Message)
						assert.Equal(t, "invalid-token", msg.Token)
						assert.Equal(t, "Test Title", msg.Android.Notification.Title)
					})
				return mockClient, "valid-server-key", "invalid-token", "Test Title", "Test Body"
			},
			wantErr: true,
		},
		{
			name: "client error",
			setup: func() (FCMClient, string, string, string, string) {
				mockClient := &mockFCMClient{}
				mockClient.On("Send", context.Background(), mock.AnythingOfType("*messaging.Message")).
					Return("", fmt.Errorf("client error")).
					Run(func(args mock.Arguments) {
						msg := args.Get(1).(*messaging.Message)
						assert.Equal(t, "test-token", msg.Token)
						assert.Equal(t, "Test Title", msg.Android.Notification.Title)
					})
				return mockClient, "valid-server-key", "test-token", "Test Title", "Test Body"
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, serverKey, token, title, body := tt.setup()
			err := sendPushNotification(client, serverKey, token, title, body)
			if (err != nil) != tt.wantErr {
				t.Errorf("sendPushNotification() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}