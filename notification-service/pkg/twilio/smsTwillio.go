package twilio

import (
	"errors"
	"log"

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

type TwilioSMSClient struct {
	c    TwilioClient
	from string
}

func NewTwilioClient(accountSID, authToken, FromPhone string) *TwilioSMSClient {
	if accountSID == "" || authToken == "" || FromPhone == "" {
		return nil
	}

	client := twilio.NewRestClientWithParams(twilio.ClientParams{
		Username: accountSID,
		Password: authToken,
	})

	return &TwilioSMSClient{
		c:    &TwilioClientWrapper{client: client},
		from: FromPhone,
	}
}

func (tc *TwilioSMSClient) SendSMS(to, message string) error {
	if tc == nil {
		return errors.New("client is nil")
	}

	if to == "" {
		return errors.New("recipient phone number cannot be empty")
	}

	if message == "" {
		return errors.New("message cannot be empty")
	}

	params := &openapi.CreateMessageParams{}
	params.SetTo(to)
	params.SetFrom(tc.from)
	params.SetBody(message)

	_, err := tc.c.CreateMessage(params)
	if err != nil {
		log.Printf("Failed to send SMS: %v", err)
		return err
	}

	log.Printf("SMS sent successfully to %s", to)
	return nil
}
