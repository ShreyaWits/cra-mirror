package httpclient_test

import (
	"context"
	httpclient "messaging_service/pkg/http"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name      string
		timeout   time.Duration
		wantError bool
	}{
		{
			name:      "valid timeout",
			timeout:   5 * time.Second,
			wantError: false,
		},
		{
			name:      "zero timeout",
			timeout:   0,
			wantError: true,
		},
		{
			name:      "negative timeout",
			timeout:   -5 * time.Second,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := httpclient.New(tt.timeout)

			if tt.wantError {
				assert.Error(t, err)
				assert.Nil(t, client)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, client)
			}
		})
	}
}

func TestClient_Do(t *testing.T) {
	// Create a test server to handle requests
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check request method
		assert.Equal(t, "GET", r.Method)

		// Check headers were set correctly
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "Bearer token123", r.Header.Get("Authorization"))

		// Return a simple response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "success"}`))
	}))
	defer ts.Close()

	// Create a client with a reasonable timeout
	client, err := httpclient.New(5 * time.Second)
	assert.NoError(t, err)
	assert.NotNil(t, client)

	// Create a request with headers and no body
	req := httpclient.Request{
		Method: "GET",
		URL:    ts.URL,
		Headers: map[string]string{
			"Content-Type":  "application/json",
			"Authorization": "Bearer token123",
		},
	}

	ctx := context.Background()
	resp, err := client.Do(ctx, req)

	// Verify response
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Don't forget to close the response body
	defer resp.Body.Close()
}

func TestClient_Do_Error(t *testing.T) {
	// Create a client
	client, err := httpclient.New(5 * time.Second)
	assert.NoError(t, err)

	// Test with invalid URL
	req := httpclient.Request{
		Method: "GET",
		URL:    "http://non-existent-domain.example.com",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	resp, err := client.Do(ctx, req)
	assert.Error(t, err)
	assert.Nil(t, resp)

	// Test with invalid method
	req = httpclient.Request{
		Method: "\n", // Invalid HTTP method
		URL:    "http://example.com",
	}

	resp, err = client.Do(context.Background(), req)
	assert.Error(t, err)
	assert.Nil(t, resp)
}
