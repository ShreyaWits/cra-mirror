package push_service

import (
	"context"

	"firebase.google.com/go/v4/messaging"
)

type FCMClient interface {
	Send(ctx context.Context, msg *messaging.Message) (string, error)
}

// Function to send push notifications using FCM
func sendPushNotification(client FCMClient, _, token, title, body string) error {

	// Construct the FCM message
	msg := &messaging.Message{
		Token: token,
		Android: &messaging.AndroidConfig{
			Notification: &messaging.AndroidNotification{
				Title: title,
				Body:  body,
			},
		},
	}
	// Send the message
	_, err := client.Send(context.Background(), msg)
	return err
}
