package firebase

import (
	"context"

	"firebase.google.com/go/v4/messaging"
	"github.com/appleboy/go-fcm"
)

type FCMClient interface {
	Send(ctx context.Context, msg ...*messaging.Message) (*messaging.BatchResponse, error)
}

var NewFCMClient = func(ctx context.Context) (FCMClient, error) {
	return fcm.NewClient(ctx)
}

// Function to send push notifications using FCM
func SendPushNotification(token, title, body string) error {
	client, err := NewFCMClient(context.Background())
	if err != nil {
		return err
	}

	msg := &messaging.Message{
		Token: token,
		Android: &messaging.AndroidConfig{
			Notification: &messaging.AndroidNotification{
				Title: title,
				Body:  body,
			},
		},
	}

	_, err = client.Send(context.Background(), msg)
	if err != nil {
		return err
	}
	return nil
}
