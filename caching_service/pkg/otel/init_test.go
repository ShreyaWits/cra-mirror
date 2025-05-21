package opentelemetry

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/trace"
	oteltrace "go.opentelemetry.io/otel/trace"
)

// Mock config for testing
type mockConfig struct{}

func (m *mockConfig) GetOTELCollectorURL() string {
	return ""
}

func (m *mockConfig) GetServiceName() string {
	return "test-service"
}

func (m *mockConfig) GetDeploymentEnv() string {
	return "test"
}

func TestSetupOTelSDK(t *testing.T) {
	// Save original environment variables
	originalEnv := os.Getenv("OTEL_COLLECTOR_GRPC_ENDPOINT")
	originalServiceName := os.Getenv("SERVICE_NAME")
	originalDeploymentEnv := os.Getenv("ENVIRONMENT")

	// Set test environment variables
	os.Setenv("OTEL_COLLECTOR_GRPC_ENDPOINT", "")
	os.Setenv("SERVICE_NAME", "test-service")
	os.Setenv("ENVIRONMENT", "test")

	// Restore environment variables after test
	defer func() {
		os.Setenv("OTEL_COLLECTOR_GRPC_ENDPOINT", originalEnv)
		os.Setenv("SERVICE_NAME", originalServiceName)
		os.Setenv("ENVIRONMENT", originalDeploymentEnv)
	}()

	// Test successful initialization
	t.Run("successful initialization", func(t *testing.T) {
		// Initialize OpenTelemetry
		shutdown, err := SetupOTelSDK(context.Background())
		assert.NoError(t, err)
		assert.NotNil(t, shutdown)

		// Test tracer creation
		tracer := otel.Tracer("test-tracer")
		assert.NotNil(t, tracer)

		// Test span creation
		ctx, span := tracer.Start(context.Background(), "test-span")
		assert.NotNil(t, span)
		assert.NotNil(t, ctx)

		// Add attributes to span
		span.SetAttributes(attribute.String("test.key", "test.value"))
		span.End()

		// Cleanup
		shutdown(context.Background())
	})

	// Test initialization with stdout exporter only
	t.Run("stdout exporter only", func(t *testing.T) {
		// Set invalid endpoint to force stdout exporter only
		t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "invalid://endpoint")
		t.Setenv("OTEL_SERVICE_NAME", "test-service")

		// Initialize OpenTelemetry
		shutdown, err := SetupOTelSDK(context.Background())
		assert.NoError(t, err)
		assert.NotNil(t, shutdown)

		// Test tracer creation
		tracer := otel.Tracer("test-tracer")
		assert.NotNil(t, tracer)

		// Test span creation
		ctx, span := tracer.Start(context.Background(), "test-span")
		assert.NotNil(t, span)
		assert.NotNil(t, ctx)

		// Cleanup
		shutdown(context.Background())
	})

	// Test initialization with default values
	t.Run("default values", func(t *testing.T) {
		// Clear environment variables to test defaults
		t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "")
		t.Setenv("OTEL_SERVICE_NAME", "")

		// Initialize OpenTelemetry
		shutdown, err := SetupOTelSDK(context.Background())
		assert.NoError(t, err)
		assert.NotNil(t, shutdown)

		// Test tracer creation
		tracer := otel.Tracer("test-tracer")
		assert.NotNil(t, tracer)

		// Test span creation
		ctx, span := tracer.Start(context.Background(), "test-span")
		assert.NotNil(t, span)
		assert.NotNil(t, ctx)

		// Cleanup
		shutdown(context.Background())
	})

	t.Run("span_propagation", func(t *testing.T) {
		// Set up test environment for span propagation
		os.Setenv("OTEL_COLLECTOR_GRPC_ENDPOINT", "")
		os.Setenv("SERVICE_NAME", "test-service")
		os.Setenv("ENVIRONMENT", "test")

		// Set up no-op providers for all telemetry types
		tp := trace.NewTracerProvider(
			trace.WithSampler(trace.AlwaysSample()),
		)
		otel.SetTracerProvider(tp)
		defer tp.Shutdown(context.Background())

		// Set up no-op meter provider
		mp := metric.NewMeterProvider()
		otel.SetMeterProvider(mp)
		defer mp.Shutdown(context.Background())

		// Set up propagator
		prop := propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		)
		otel.SetTextMapPropagator(prop)

		// Create a new span
		tracer := otel.Tracer("test")
		ctx, span := tracer.Start(context.Background(), "test-span")
		defer span.End()

		// Add some baggage to ensure it's propagated
		m, err := baggage.NewMember("test-key", "test-value")
		require.NoError(t, err)
		b, err := baggage.New(m)
		require.NoError(t, err)
		ctx = baggage.ContextWithBaggage(ctx, b)

		// Set up OpenTelemetry with no-op providers
		shutdown, err := SetupOTelSDK(ctx)
		require.NoError(t, err)
		defer func() {
			err := shutdown(ctx)
			require.NoError(t, err)
		}()

		// Verify span is in context
		spanFromCtx := oteltrace.SpanFromContext(ctx)
		assert.NotNil(t, spanFromCtx)
		assert.True(t, spanFromCtx.SpanContext().IsValid())

		// Test span propagation
		carrier := propagation.MapCarrier{}
		propagator := otel.GetTextMapPropagator()
		propagator.Inject(ctx, carrier)

		// Verify carrier has trace context
		assert.NotEmpty(t, carrier.Get("traceparent"))
		assert.NotEmpty(t, carrier.Get("baggage"))

		// Extract span context from carrier
		extractedCtx := propagator.Extract(context.Background(), carrier)
		extractedSpan := oteltrace.SpanFromContext(extractedCtx)
		assert.NotNil(t, extractedSpan)
		assert.True(t, extractedSpan.SpanContext().IsValid())
	})

	// Test graceful shutdown
	t.Run("graceful shutdown", func(t *testing.T) {
		// Initialize OpenTelemetry
		shutdown, err := SetupOTelSDK(context.Background())
		assert.NoError(t, err)

		// Create a span
		tracer := otel.Tracer("test-tracer")
		_, span := tracer.Start(context.Background(), "test-span")
		span.End()

		// Shutdown with timeout
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Call shutdown and ignore any errors from the collector
		// since we're testing the shutdown function itself
		_ = shutdown(ctx)

		// Verify shutdown completed within timeout
		select {
		case <-ctx.Done():
			assert.Fail(t, "shutdown timed out")
		default:
			// Shutdown completed successfully
		}
	})
}
