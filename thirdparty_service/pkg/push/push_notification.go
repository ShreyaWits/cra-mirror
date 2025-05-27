package push_service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"golang.org/x/oauth2/google"
)

type httpDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

type FCMClient struct {
	projectID string
	client    httpDoer
	token     string
	endpoint  string
}

// NewFCMClient creates a new FCMClient by loading credentials JSON and project ID
func NewFCMClient(ctx context.Context, credsData string, projectID string) (*FCMClient, error) {
	if credsData == "" || projectID == "" {
		return nil, fmt.Errorf("credentials and projectID must be provided")
	}

	credsBytes := []byte(credsData)

	conf, err := google.JWTConfigFromJSON(credsBytes, "https://www.googleapis.com/auth/firebase.messaging")
	if err != nil {
		return nil, fmt.Errorf("failed to parse JWT config: %w", err)
	}

	httpClient := conf.Client(ctx)

	token, err := conf.TokenSource(ctx).Token()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve token: %w", err)
	}

	return &FCMClient{
		projectID: projectID,
		client:    httpClient,
		token:     token.AccessToken,
		endpoint:  fmt.Sprintf("https://fcm.googleapis.com/v1/projects/%s/messages:send", projectID),
	}, nil
}

// NewFCMClientWithHTTPClient creates an FCMClient with given http client and token (for testing)
func NewFCMClientWithHTTPClient(projectID string, client httpDoer, token string) *FCMClient {
	return &FCMClient{
		projectID: projectID,
		client:    client,
		token:     token,
		endpoint:  fmt.Sprintf("https://fcm.googleapis.com/v1/projects/%s/messages:send", projectID),
	}
}

// SendNotification sends a push notification using the FCM v1 API
func (f *FCMClient) SendNotification(ctx context.Context, toToken, title, body string) error {
	if toToken == "" || title == "" || body == "" {
		return fmt.Errorf("toToken, title, and body are required")
	}

	payload := map[string]interface{}{
		"message": map[string]interface{}{
			"token": toToken,
			"notification": map[string]string{
				"title": title,
				"body":  body,
			},
		},
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", f.endpoint, bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+f.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := f.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send FCM request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("FCM error: status %d - %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}
