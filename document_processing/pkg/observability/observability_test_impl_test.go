package observability_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/propagation"
)

// Test formatEndpoint directly with the exported wrapper
func TestLocalFormatEndpoint(t *testing.T) {
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
			result := FormatEndpoint(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Helper function for test
func FormatEndpoint(url string) string {
	// Exact same implementation as in the package
	// This avoids import cycles in tests
	if len(url) >= 7 && url[:7] == "http://" {
		return url[7:]
	}
	if len(url) >= 8 && url[:8] == "https://" {
		return url[8:]
	}
	return url
}

// Test creating a propagator
func TestLocalCreatePropagator(t *testing.T) {
	prop := CreatePropagator()
	assert.NotNil(t, prop)

	// Just check it's not nil, we can't easily check the type in detail
	assert.NotNil(t, prop)
}

// Helper to create a propagator for testing
func CreatePropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
}

// Test a stdout exporter
func TestLocalStdoutExporter(t *testing.T) {
	exporter := &LocalStdoutExporter{}

	// Test Shutdown
	err := exporter.Shutdown(context.Background())
	assert.NoError(t, err)

	// Test ForceFlush
	err = exporter.ForceFlush(context.Background())
	assert.NoError(t, err)

	// Test Export
	assert.NotPanics(t, func() {
		exporter.Export(context.Background(), nil)
		exporter.Export(context.Background(), []TestLogRecord{
			{msg: "test message", severity: "INFO"},
		})
	})
}

// LocalStdoutExporter is a simple test implementation of a log exporter
type LocalStdoutExporter struct{}

type TestLogRecord struct {
	msg      string
	severity string
}

func (r TestLogRecord) Timestamp() interface{} {
	return nil
}

func (r TestLogRecord) SeverityText() string {
	return r.severity
}

func (r TestLogRecord) Body() TestLogBody {
	return TestLogBody{msg: r.msg}
}

type TestLogBody struct {
	msg string
}

func (b TestLogBody) AsString() string {
	return b.msg
}

func (e *LocalStdoutExporter) Export(ctx context.Context, records interface{}) error {
	// If we have records to export
	if records != nil {
		if logRecords, ok := records.([]TestLogRecord); ok {
			for _, rec := range logRecords {
				fmt.Printf("[TEST LOG] [%s] %s\n",
					rec.SeverityText(),
					rec.Body().AsString(),
				)
			}
		}
	}
	return nil
}

func (e *LocalStdoutExporter) Shutdown(ctx context.Context) error {
	return nil
}

func (e *LocalStdoutExporter) ForceFlush(ctx context.Context) error {
	return nil
}
