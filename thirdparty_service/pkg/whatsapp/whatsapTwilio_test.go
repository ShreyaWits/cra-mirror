package whatsapp

import (
	"errors"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
)

// --- Helper to replace http.DefaultClient temporarily ---
var originalClient = http.DefaultClient

func mockHTTPClient(mockDo func(req *http.Request) (*http.Response, error)) {
	http.DefaultClient = &http.Client{
		Transport: roundTripFunc(mockDo),
	}
}

type roundTripFunc func(req *http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

// --- Tests ---

func TestNewWhatsAppClient_Success(t *testing.T) {
	client, err := NewWhatsAppClient("sid", "token", "+1234567890")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if client == nil {
		t.Fatal("expected client, got nil")
	}
	if client.FromNumber != "whatsapp:+1234567890" {
		t.Errorf("unexpected FromNumber: %s", client.FromNumber)
	}
}

func TestNewWhatsAppClient_MissingFields(t *testing.T) {
	_, err := NewWhatsAppClient("", "token", "from")
	if err == nil {
		t.Fatal("expected error for empty SID")
	}
	_, err = NewWhatsAppClient("sid", "", "from")
	if err == nil {
		t.Fatal("expected error for empty token")
	}
	_, err = NewWhatsAppClient("sid", "token", "")
	if err == nil {
		t.Fatal("expected error for empty from number")
	}
}

func TestSendWhatsAppMessage_Success(t *testing.T) {
	mockHTTPClient(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 200,
			Body:       ioutil.NopCloser(strings.NewReader("ok")),
		}, nil
	})
	defer func() { http.DefaultClient = originalClient }()

	client := &WhatsAppClient{
		AccountSID: "sid",
		AuthToken:  "token",
		FromNumber: "whatsapp:+1234567890",
	}

	err := client.SendWhatsAppMessage("+919999999999", "Hello!")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestSendWhatsAppMessage_FailureStatus(t *testing.T) {
	mockHTTPClient(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 400,
			Status:     "400 Bad Request",
			Body:       ioutil.NopCloser(strings.NewReader("error")),
		}, nil
	})
	defer func() { http.DefaultClient = originalClient }()

	client := &WhatsAppClient{
		AccountSID: "sid",
		AuthToken:  "token",
		FromNumber: "whatsapp:+1234567890",
	}

	err := client.SendWhatsAppMessage("+919999999999", "Hello!")
	if err == nil || !strings.Contains(err.Error(), "failed to send WhatsApp message") {
		t.Fatalf("expected send failure, got: %v", err)
	}
}

func TestSendWhatsAppMessage_HTTPError(t *testing.T) {
	mockHTTPClient(func(req *http.Request) (*http.Response, error) {
		return nil, errors.New("network error")
	})
	defer func() { http.DefaultClient = originalClient }()

	client := &WhatsAppClient{
		AccountSID: "sid",
		AuthToken:  "token",
		FromNumber: "whatsapp:+1234567890",
	}

	err := client.SendWhatsAppMessage("+919999999999", "Hello!")
	if err == nil || !strings.Contains(err.Error(), "network error") {
		t.Fatalf("expected network error, got: %v", err)
	}
}

func TestSendWhatsAppMessage_EmptyInput(t *testing.T) {
	client := &WhatsAppClient{
		AccountSID: "sid",
		AuthToken:  "token",
		FromNumber: "whatsapp:+1234567890",
	}

	err := client.SendWhatsAppMessage("", "hello")
	if err == nil {
		t.Fatal("expected error for empty to")
	}

	err = client.SendWhatsAppMessage("recipient", "")
	if err == nil {
		t.Fatal("expected error for empty message")
	}
}
