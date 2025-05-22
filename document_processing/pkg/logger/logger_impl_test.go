package logger_test

import (
	"context"
	"io"
	"Document-Processing/pkg/logger"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormatEndpoint(t *testing.T) {
	// This is an exported helper in the package, we'll need to make it public for testing
	// Since it's not exported yet, this test will fail initially
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "no protocol",
			input:    "example.com:8080",
			expected: "example.com:8080",
		},
		{
			name:     "http protocol",
			input:    "http://example.com:8080",
			expected: "example.com:8080",
		},
		{
			name:     "https protocol",
			input:    "https://example.com:8080",
			expected: "example.com:8080",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := logger.FormatEndpoint(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetOutboundIP(t *testing.T) {
	// This is a non-exported function, so we'll test it indirectly through
	// the NewLogger function that uses it

	// Capture stdout to verify logger creation message
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Create a new logger
	log := logger.NewLogger("test-service", false)

	// Close the writer and restore stdout
	w.Close()
	os.Stdout = oldStdout

	// Read the captured output
	var buf strings.Builder
	io.Copy(&buf, r)

	// Verify the output contains a creation message
	output := buf.String()
	assert.Contains(t, output, "Created standard logger for service: test-service")

	// Test that the logger is functional
	assert.NotPanics(t, func() {
		log.Info(context.Background(), "test message")
	})
}

// Test the NewLogger function with different configurations
func TestNewLogger(t *testing.T) {
	tests := []struct {
		name      string
		service   string
		isEnabled bool
	}{
		{
			name:      "disabled OTel",
			service:   "test-service-1",
			isEnabled: false,
		},
		{
			name:      "enabled OTel",
			service:   "test-service-2",
			isEnabled: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// Create logger
			log := logger.NewLogger(tt.service, tt.isEnabled)

			// Restore stdout
			w.Close()
			os.Stdout = oldStdout

			// Read the captured output
			var buf strings.Builder
			io.Copy(&buf, r)
			output := buf.String()

			// Check output contains expected service name
			assert.Contains(t, output, tt.service)

			// Verify we have the correct message based on OTel being enabled or not
			if tt.isEnabled {
				assert.Contains(t, output, "OpenTelemetry-enabled logger")
			} else {
				assert.Contains(t, output, "standard logger")
				assert.Contains(t, output, "OpenTelemetry disabled")
			}

			// Check that the logger implements the Logger interface
			_, ok := log.(logger.Logger)
			assert.True(t, ok)

			// Verify it doesn't panic when used
			assert.NotPanics(t, func() {
				log.Info(context.Background(), "test message")
				log.Error(context.Background(), "test error")
				log.Debug(context.Background(), "test debug")
				log.Warn(context.Background(), "test warning")

				// Test WithFields
				withFields := log.WithFields(map[string]interface{}{
					"key": "value",
				})
				withFields.Info(context.Background(), "test with fields")

				// Test Sync
				log.Sync()
			})
		})
	}
}
