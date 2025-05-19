package twilio_sms

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"testing"
)

// mockRoundTripper implements http.RoundTripper
type mockRoundTripper struct {
	mockResponse *http.Response
	mockError    error
	captureReq   **http.Request // optional: capture request for assertions
}

func (m *mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if m.captureReq != nil {
		*m.captureReq = req
	}
	return m.mockResponse, m.mockError
}

// helper to create mock client
func newMockClient(resp *http.Response, err error, captureReq **http.Request) *http.Client {
	return &http.Client{
		Transport: &mockRoundTripper{
			mockResponse: resp,
			mockError:    err,
			captureReq:   captureReq,
		},
	}
}

func TestSendSMS_Success(t *testing.T) {
	var captured *http.Request

	client := &TwilioClient{
		AccountSID: "AC123456",
		AuthToken:  "auth_token",
		HTTPClient: newMockClient(&http.Response{
			StatusCode: 201,
			Body:       io.NopCloser(bytes.NewBufferString(`{"sid":"SM123"}`)),
		}, nil, &captured),
	}

	err := client.SendSMS("+1234567890", "+1987654321", "Hello from Twilio")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if captured != nil {
		bodyBytes, _ := io.ReadAll(captured.Body)
		bodyStr := string(bodyBytes)

		expectedFields := []string{
			"To=%2B1987654321",
			"From=%2B1234567890",
			"Body=Hello+from+Twilio",
		}
		for _, f := range expectedFields {
			if !strings.Contains(bodyStr, f) {
				t.Errorf("expected form field %s in body, got %s", f, bodyStr)
			}
		}
	}

}

func TestSendSMS_TwilioAPIError(t *testing.T) {
	client := &TwilioClient{
		AccountSID: "AC123456",
		AuthToken:  "auth_token",
		HTTPClient: newMockClient(&http.Response{
			StatusCode: 400,
			Body:       io.NopCloser(strings.NewReader("Bad Request")),
		}, nil, nil),
	}

	err := client.SendSMS("+from", "+to", "body")
	if err == nil || !strings.Contains(err.Error(), "twilio API error") {
		t.Errorf("expected Twilio API error, got %v", err)
	}
}

func TestSendSMS_RequestCreationError(t *testing.T) {
	client := &TwilioClient{
		AccountSID: "AC123456",
		AuthToken:  "auth_token",
		HTTPClient: &http.Client{},
	}

	// Inject invalid phone number that leads to an invalid URL
	// because the URL will become invalid like https://api.twilio.com/.../Messages.json?%xx
	invalidSID := string([]byte{0x7f}) // invalid URL byte
	client.AccountSID = invalidSID

	err := client.SendSMS("+from", "+to", "body")
	if err == nil || !strings.Contains(err.Error(), "failed to create request") {
		t.Errorf("expected request creation error, got %v", err)
	}
}

func TestSendSMS_NetworkError(t *testing.T) {
	client := &TwilioClient{
		AccountSID: "AC123456",
		AuthToken:  "auth_token",
		HTTPClient: newMockClient(nil, http.ErrHandlerTimeout, nil),
	}

	err := client.SendSMS("+123", "+456", "test")
	if err == nil || !strings.Contains(err.Error(), "failed to send request") {
		t.Errorf("expected network error, got %v", err)
	}
}

func TestSendSMS_ResponseBodyNil(t *testing.T) {
	client := &TwilioClient{
		AccountSID: "AC123456",
		AuthToken:  "auth_token",
		HTTPClient: newMockClient(&http.Response{
			StatusCode: 200,
			Body:       nil, // simulate nil body
		}, nil, nil),
	}

	err := client.SendSMS("+123", "+456", "test")
	if err != nil {
		t.Errorf("expected no error with nil body, got %v", err)
	}
}
