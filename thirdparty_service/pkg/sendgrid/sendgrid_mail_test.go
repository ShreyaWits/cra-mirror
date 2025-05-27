package sendgrid

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
)

// mockRoundTripper to mock HTTP responses
type mockRoundTripper struct {
	mockResponse *http.Response
	mockError    error
}

func (m *mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.mockResponse, m.mockError
}

func sampleEmail() Email {
	return Email{
		Personalizations: []Personalization{
			{
				To: []Contact{{Email: "to@example.com"}},
				Subject: "subject",
			},
		},
		From: Contact{Email: "from@example.com"},
		ReplyTo: Contact{Email: "replyto@example.com"},
		Content: []Content{{Type: "text/plain", Value: "hello"}},
	}
}

func TestSendEmail_Success(t *testing.T) {
	email := sampleEmail()

	client := &SendGridClient{
		APIKey: "dummy",
		HTTPClient: &http.Client{
			Transport: &mockRoundTripper{
				mockResponse: &http.Response{
					StatusCode: 202,
					Body:       io.NopCloser(strings.NewReader("Accepted")),
				},
			},
		},
	}

	err := client.SendEmail(email)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestSendEmail_APIFailure(t *testing.T) {
	email := sampleEmail()

	client := &SendGridClient{
		APIKey: "dummy",
		HTTPClient: &http.Client{
			Transport: &mockRoundTripper{
				mockResponse: &http.Response{
					StatusCode: 400,
					Body:       io.NopCloser(strings.NewReader("Bad Request")),
				},
			},
		},
	}

	err := client.SendEmail(email)
	if err == nil || !strings.Contains(err.Error(), "sendgrid API error") {
		t.Errorf("expected API error, got %v", err)
	}
}

func TestSendEmail_RequestError(t *testing.T) {
	email := sampleEmail()

	client := &SendGridClient{
		APIKey: "dummy",
		HTTPClient: &http.Client{
			Transport: &mockRoundTripper{
				mockError: errors.New("network error"),
			},
		},
	}

	err := client.SendEmail(email)
	if err == nil || !strings.Contains(err.Error(), "failed to send request") {
		t.Errorf("expected request error, got %v", err)
	}
}

func TestSendEmail_RequestCreationError(t *testing.T) {
    email := sampleEmail()

    client := &SendGridClient{
        APIKey: "key",
        HTTPClient: &http.Client{
            Transport: &mockRoundTripper{
                mockResponse: nil,
                mockError:    errors.New("should not reach http"),
            },
        },
        newRequestFunc: func(method, url string, body io.Reader) (*http.Request, error) {
            return nil, errors.New("failed to create request")
        },
    }

    err := client.SendEmail(email)
    if err == nil || !strings.Contains(err.Error(), "failed to create request") {
        t.Fatalf("expected request creation error, got %v", err)
    }
}
