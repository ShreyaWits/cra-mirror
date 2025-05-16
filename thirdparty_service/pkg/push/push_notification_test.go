package push_service

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"testing"
)

// mockHTTPClient to simulate http.Client
type mockHTTPClient struct {
	doFunc func(req *http.Request) (*http.Response, error)
}

func (m *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return m.doFunc(req)
}

// A minimal fake JWT JSON (note: will cause error on NewFCMClient because private key invalid)
const fakeCredsJSON = `{
	"type": "service_account",
	"project_id": "fake-project",
	"private_key_id": "fakekeyid",
	"private_key": "-----BEGIN PRIVATE KEY-----\nfake\n-----END PRIVATE KEY-----\n",
	"client_email": "fake@fake-project.iam.gserviceaccount.com",
	"client_id": "1234567890",
	"auth_uri": "https://accounts.google.com/o/oauth2/auth",
	"token_uri": "https://oauth2.googleapis.com/token",
	"auth_provider_x509_cert_url": "https://www.googleapis.com/oauth2/v1/certs",
	"client_x509_cert_url": "https://www.googleapis.com/robot/v1/metadata/x509/fake"
}`

// Test NewFCMClientWithHTTPClient and SendNotification with mocks
func TestFCMClient_SendNotification(t *testing.T) {
	ctx := context.Background()

	mockClient := &mockHTTPClient{
		doFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(bytes.NewBufferString(`{}`)),
			}, nil
		},
	}

	client := NewFCMClientWithHTTPClient("fake-project", mockClient, "fake-token")

	// success case
	err := client.SendNotification(ctx, "target-token", "title", "body")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	// missing params
	for _, tc := range []struct {
		toToken string
		title   string
		body    string
	}{
		{"", "title", "body"},
		{"token", "", "body"},
		{"token", "title", ""},
	} {
		err := client.SendNotification(ctx, tc.toToken, tc.title, tc.body)
		if err == nil {
			t.Errorf("expected error for missing parameters, got nil")
		}
	}

	// network error
	client.client = &mockHTTPClient{
		doFunc: func(req *http.Request) (*http.Response, error) {
			return nil, errors.New("network error")
		},
	}
	err = client.SendNotification(ctx, "token", "title", "body")
	if err == nil || err.Error() != "failed to send FCM request: network error" {
		t.Errorf("expected network error, got %v", err)
	}

	// HTTP error status
	client.client = &mockHTTPClient{
		doFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 400,
				Body:       io.NopCloser(bytes.NewBufferString("bad request")),
			}, nil
		},
	}
	err = client.SendNotification(ctx, "token", "title", "body")
	if err == nil || err.Error() != "FCM error: status 400 - bad request" {
		t.Errorf("expected FCM error, got %v", err)
	}
}

// Test NewFCMClient errors for empty creds/project and invalid creds
func TestNewFCMClient(t *testing.T) {
	ctx := context.Background()

	// empty creds
	_, err := NewFCMClient(ctx, "", "project")
	if err == nil {
		t.Errorf("expected error for empty creds, got nil")
	}

	// empty project
	_, err = NewFCMClient(ctx, fakeCredsJSON, "")
	if err == nil {
		t.Errorf("expected error for empty projectID, got nil")
	}

	// invalid creds JSON
	_, err = NewFCMClient(ctx, "invalid-json", "project")
	if err == nil {
		t.Errorf("expected error for invalid creds JSON, got nil")
	}

	// valid creds JSON but invalid private key (will fail to retrieve token)
	_, err = NewFCMClient(ctx, fakeCredsJSON, "fake-project")
	if err == nil {
		t.Errorf("expected error for invalid private key in creds, got nil")
	}
}
