package httpclient

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

// HTTP method constants
const (
	GET    = "GET"
	POST   = "POST"
	PUT    = "PUT"
	DELETE = "DELETE"
	PATCH  = "PATCH"
)

// Request represents an HTTP request
type Request struct {
	Method  string
	URL     string
	Headers map[string]string
	Body    []byte
}

// HTTPClient interface defines the methods for making HTTP requests
type HTTPClient interface {
	Do(ctx context.Context, req Request) (*http.Response, error)
}

type client struct {
	httpClient *http.Client
}

// New creates a new HTTP client with the specified timeout
func New(timeout time.Duration) HTTPClient {
	return &client{
		httpClient: &http.Client{Timeout: timeout},
	}
}

// Do executes an HTTP request and returns the response
func (c *client) Do(ctx context.Context, req Request) (*http.Response, error) {
	var bodyReader io.Reader

	// Only create a buffer if body is not nil
	if req.Body != nil && len(req.Body) > 0 {
		bodyReader = bytes.NewBuffer(req.Body)
	}

	httpReq, err := http.NewRequestWithContext(ctx, req.Method, req.URL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("request creation failed: %w", err)
	}

	// Set headers
	for k, v := range req.Headers {
		httpReq.Header.Set(k, v)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("http request failed: %w", err)
	}

	return resp, nil
}
