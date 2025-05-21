package observability

import (
	"context"

	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/log"
)

// TestNewPropagator is a test wrapper for the private newPropagator function
func TestNewPropagator() propagation.TextMapPropagator {
	return newPropagator()
}

// NewStdoutExporterForTest creates a new StdoutExporter for testing
func NewStdoutExporterForTest() *StdoutExporter {
	return &StdoutExporter{}
}

// TestResetGlobalProviders resets global providers to their default state
// This is useful for cleaning up after tests that modify global state
func TestResetGlobalProviders() {
	// Create empty no-op providers for cleanup
	noopLoggerProvider := log.NewLoggerProvider()
	global.SetLoggerProvider(noopLoggerProvider)
}

// TestFormatEndpointWrapper provides test access to the private formatEndpoint function
func TestFormatEndpointWrapper(url string) string {
	return formatEndpoint(url)
}

// TestCreateEmptyResource creates an empty resource for testing
func TestCreateEmptyResource(ctx context.Context) (interface{}, error) {
	res, err := createEmptyResource(ctx)
	return res, err
}

// Helper to create an empty resource for testing
func createEmptyResource(ctx context.Context) (interface{}, error) {
	return nil, nil
}
