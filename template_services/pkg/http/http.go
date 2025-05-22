package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"
)

// HTTPClient defines an interface for making HTTP requests.
type HTTPClient interface {
	Do(ctx context.Context, method, url string, headers map[string]string, body interface{}) (*http.Response, error)
}

// client implements HTTPClient.
type client struct {
	httpClient *http.Client
}

// New creates a new HTTP client with a timeout.
func New(timeout time.Duration) HTTPClient {
	return &client{
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// Do sends an HTTP request and returns the response.
func (c *client) Do(ctx context.Context, method, url string, headers map[string]string, body interface{}) (*http.Response, error) {
	var bodyReader io.Reader

	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, err
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return c.httpClient.Do(req)
}
