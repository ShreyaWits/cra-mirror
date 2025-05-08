package whatsapp

import (
	"errors"
	"log"
	"os"

	"github.com/twilio/twilio-go"
	openapi "github.com/twilio/twilio-go/rest/api/v2010"
)

// TwilioClient interface for mocking
type TwilioClient interface {
	CreateMessage(params *openapi.CreateMessageParams) (*openapi.ApiV2010Message, error)
}

// TwilioClientWrapper wraps the actual Twilio client
type TwilioClientWrapper struct {
	client *twilio.RestClient
}

func (w *TwilioClientWrapper) CreateMessage(params *openapi.CreateMessageParams) (*openapi.ApiV2010Message, error) {
	return w.client.Api.CreateMessage(params)
}

// Default client creation function
var newTwilioClient = func(accountSid, authToken string) TwilioClient {
	client := twilio.NewRestClientWithParams(twilio.ClientParams{
		Username: accountSid,
		Password: authToken,
	})
	return &TwilioClientWrapper{client: client}
}

func SendWhatsAppMessage(to, body string) error {
	if to == "" {
		return errors.New("invalid phone number: recipient cannot be empty")
	}

	if body == "" {
		return errors.New("message body cannot be empty")
	}

	accountSid := os.Getenv("TWILIO_ACCOUNT_SID")
	if accountSid == "" {
		return errors.New("TWILIO_ACCOUNT_SID environment variable is not set")
	}

	authToken := os.Getenv("TWILIO_AUTH_TOKEN")
	if authToken == "" {
		return errors.New("TWILIO_AUTH_TOKEN environment variable is not set")
	}

	client := newTwilioClient(accountSid, authToken)

	params := &openapi.CreateMessageParams{}
	params.SetTo("whatsapp:" + to)
	params.SetFrom("whatsapp:+14155238886") // Twilio sandbox number
	params.SetBody(body)

	_, err := client.CreateMessage(params)
	if err != nil {
		log.Printf("Error sending WhatsApp message: %v", err)
		return err
	}

	log.Println("WhatsApp message sent successfully!")
	return nil
}
