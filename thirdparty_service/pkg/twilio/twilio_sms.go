package twilio_sms

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// TwilioClient encapsulates the credentials needed to interact with Twilio API
type TwilioClient struct {
	AccountSID string
	AuthToken  string
	HTTPClient *http.Client
}

// NewTwilioClient creates a new instance of TwilioClient
func NewTwilioClient(accountSID, authToken string) *TwilioClient {
	return &TwilioClient{
		AccountSID: accountSID,
		AuthToken:  authToken,
		HTTPClient: &http.Client{},
	}
}

// SendSMS sends an SMS using Twilio's REST API
func (c *TwilioClient) SendSMS(from, to, body string) error {
	endpoint := fmt.Sprintf("https://api.twilio.com/2010-04-01/Accounts/%s/Messages.json", c.AccountSID)

	form := url.Values{}
	form.Set("To", to)
	form.Set("From", from)
	form.Set("Body", body)

	req, err := http.NewRequest("POST", endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.SetBasicAuth(c.AccountSID, c.AuthToken)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("twilio API error: status %d", resp.StatusCode)
	}

	return nil
}
