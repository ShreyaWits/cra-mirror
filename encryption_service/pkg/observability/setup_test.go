package observability

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormatEndpoint(t *testing.T) {
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
			result := formatEndpoint(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestStdoutExporter(t *testing.T) {
	exporter := &StdoutExporter{}

	// Test that methods don't panic
	assert.NotPanics(t, func() {
		err := exporter.Export(context.Background(), nil)
		assert.NoError(t, err)

		err = exporter.Shutdown(context.Background())
		assert.NoError(t, err)

		err = exporter.ForceFlush(context.Background())
		assert.NoError(t, err)
	})
}
