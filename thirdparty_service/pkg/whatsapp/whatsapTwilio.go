package whatsapp

import (
	"bytes"
	"errors"
	"log"
	"net/http"
	"net/url"
)

// WhatsAppClient defines a Twilio WhatsApp API client
type WhatsAppClient struct {
	AccountSID string
	AuthToken  string
	FromNumber string
}

// NewClient creates a new instance of WhatsAppClient
func NewWhatsAppClient(accountSID, authToken, fromNumber string) (*WhatsAppClient, error) {
	if accountSID == "" || authToken == "" || fromNumber == "" {
		return nil, errors.New("account SID, auth token, and from number must not be empty")
	}
	return &WhatsAppClient{
		AccountSID: accountSID,
		AuthToken:  authToken,
		FromNumber: "whatsapp:" + fromNumber,
	}, nil
}

// SendMessage sends a WhatsApp message using Twilio's REST API
func (c *WhatsAppClient) SendWhatsAppMessage(to, body string) error {
	if to == "" || body == "" {
		return errors.New("recipient and message body must not be empty")
	}

	endpoint := "https://api.twilio.com/2010-04-01/Accounts/" + c.AccountSID + "/Messages.json"

	form := url.Values{}
	form.Set("To", "whatsapp:"+to)
	form.Set("From", c.FromNumber)
	form.Set("Body", body)

	req, err := http.NewRequest("POST", endpoint, bytes.NewBufferString(form.Encode()))
	if err != nil {
		return err
	}

	req.SetBasicAuth(c.AccountSID, c.AuthToken)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		log.Println("✅ WhatsApp message sent successfully!")
		return nil
	}

	log.Printf("❌ Failed to send WhatsApp message. Status: %s\n", resp.Status)
	return errors.New("failed to send WhatsApp message")
}
