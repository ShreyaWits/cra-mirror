package services

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

type mockTransport struct {
	response *http.Response
	err      error
}

func (m *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.response, m.err
}

func TestLlamaService_ProcessImage(t *testing.T) {
	validBase64 := "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8z/C/HgAGgwJ/lK3Q6wAAAABJRU5ErkJggg=="
	invalidBase64 := "invalid_base64"

	tests := []struct {
		name             string
		base64Image      string
		extractionFields []string
		mockResponse     *http.Response
		mockErr          error
		wantErr          bool
		wantFields       map[string]string
		wantErrMsg       string
	}{
		// Success cases
		{
			name:             "successful image processing with single field",
			base64Image:      validBase64,
			extractionFields: []string{"name"},
			mockResponse: &http.Response{
				StatusCode: http.StatusOK,
				Body: io.NopCloser(bytes.NewBufferString(`{
					"model": "llama3.2-vision",
					"message": {
						"role": "assistant",
						"content": "{\"name\": \"John Doe\"}"
					},
					"done": true
				}`)),
			},
			wantErr:    false,
			wantFields: map[string]string{"name": "John Doe"},
		},
		{
			name:             "successful image processing with multiple fields",
			base64Image:      validBase64,
			extractionFields: []string{"name", "age", "gender"},
			mockResponse: &http.Response{
				StatusCode: http.StatusOK,
				Body: io.NopCloser(bytes.NewBufferString(`{
					"model": "llama3.2-vision",
					"message": {
						"role": "assistant",
						"content": "{\"name\": \"Alice\", \"age\": \"30\", \"gender\": \"female\"}"
					},
					"done": true
				}`)),
			},
			wantErr: false,
			wantFields: map[string]string{
				"name":   "Alice",
				"age":    "30",
				"gender": "female",
			},
		},

		// Error cases
		{
			name:             "invalid base64 image",
			base64Image:      invalidBase64,
			extractionFields: []string{"name"},
			mockResponse:     nil,
			mockErr:          nil,
			wantErr:          true,
			wantErrMsg:       "base64 validation failed",
		},
		{
			name:             "HTTP request failure",
			base64Image:      validBase64,
			extractionFields: []string{"name"},
			mockErr:          errors.New("connection refused"),
			wantErr:          true,
			wantErrMsg:       "HTTP request failed",
		},
		{
			name:             "non-success HTTP status",
			base64Image:      validBase64,
			extractionFields: []string{"name"},
			mockResponse: &http.Response{
				StatusCode: http.StatusBadRequest,
				Body:       io.NopCloser(bytes.NewBufferString("Bad request")),
			},
			wantErr:    true,
			wantErrMsg: "non-success HTTP status",
		},
		{
			name:             "invalid JSON response structure",
			base64Image:      validBase64,
			extractionFields: []string{"name"},
			mockResponse: &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBufferString("invalid json")),
			},
			wantErr:    true,
			wantErrMsg: "failed to unmarshal JSON",
		},
		{
			name:             "invalid content JSON",
			base64Image:      validBase64,
			extractionFields: []string{"name"},
			mockResponse: &http.Response{
				StatusCode: http.StatusOK,
				Body: io.NopCloser(bytes.NewBufferString(`{
					"model": "llama3.2-vision",
					"message": {
						"role": "assistant",
						"content": "invalid json"
					},
					"done": true
				}`)),
			},
			wantErr:    true,
			wantErrMsg: "failed to unmarshal JSON",
		},
		{
			name:             "empty response content",
			base64Image:      validBase64,
			extractionFields: []string{"name"},
			mockResponse: &http.Response{
				StatusCode: http.StatusOK,
				Body: io.NopCloser(bytes.NewBufferString(`{
					"model": "llama3.2-vision",
					"message": {
						"role": "assistant",
						"content": ""
					},
					"done": true
				}`)),
			},
			wantErr:    true,
			wantErrMsg: "failed to unmarshal JSON",
		},
		{
			name:             "partial fields in response",
			base64Image:      validBase64,
			extractionFields: []string{"name", "age"},
			mockResponse: &http.Response{
				StatusCode: http.StatusOK,
				Body: io.NopCloser(bytes.NewBufferString(`{
					"model": "llama3.2-vision",
					"message": {
						"role": "assistant",
						"content": "{\"name\": \"Bob\"}"
					},
					"done": true
				}`)),
			},
			wantErr:    false,
			wantFields: map[string]string{"name": "Bob"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &http.Client{
				Transport: &mockTransport{
					response: tt.mockResponse,
					err:      tt.mockErr,
				},
			}
			modelName := "llama3.2-vision"
			apiURL := "http://localhost:11434/api/chat"

			service := NewLlamaServiceWithClient(modelName, apiURL, mockClient)
			got, err := service.ProcessImage(context.Background(), tt.base64Image, tt.extractionFields)

			if (err != nil) != tt.wantErr {
				t.Errorf("ProcessImage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				if tt.wantErrMsg != "" && !strings.Contains(err.Error(), tt.wantErrMsg) {
					t.Errorf("ProcessImage() error = %v, wantErrMsg containing %q", err, tt.wantErrMsg)
				}
				return
			}

			if !compareMaps(got, tt.wantFields) {
				t.Errorf("ProcessImage() = %v, want %v", got, tt.wantFields)
			}
		})
	}
}

func TestNewLlamaService(t *testing.T) {
	t.Run("default client has timeout", func(t *testing.T) {
		modelName := "llama3.2-vision"
		apiURL := "http://localhost:11434/api/chat"

		service := NewLlamaService(modelName, apiURL)
		if service.client.Timeout != 2*time.Minute {
			t.Errorf("Expected default client timeout of 2 minutes, got %v", service.client.Timeout)
		}
	})

	t.Run("custom client is used when provided", func(t *testing.T) {
		modelName := "llama3.2-vision"
		apiURL := "http://localhost:11434/api/chat"
		customClient := &http.Client{Timeout: 5 * time.Minute}
		service := NewLlamaServiceWithClient(modelName, apiURL, customClient)
		if service.client != customClient {
			t.Error("Expected custom client to be used")
		}
	})
}

func compareMaps(a, b map[string]string) bool {
	if len(b) == 0 {
		return len(a) == 0
	}
	for k, v := range b {
		if a[k] != v {
			return false
		}
	}
	return true
}
