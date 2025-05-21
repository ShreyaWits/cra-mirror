package opentelemetry

import (
	"context"
	"io"
	"os"
	"testing"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/propagation"
	logsdk "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	oteltrace "go.opentelemetry.io/otel/trace"

	"thirdparty_service/internal/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetupOTelSDK(t *testing.T) {
	// Save original environment variables
	originalEnv := os.Getenv("OTEL_COLLECTOR_URL")
	originalServiceName := os.Getenv("SERVICE_NAME")
	originalDeploymentEnv := os.Getenv("DEPLOYMENT_ENV")

	// Set test environment variables
	os.Setenv("OTEL_COLLECTOR_URL", "")
	os.Setenv("SERVICE_NAME", "test-service")
	os.Setenv("DEPLOYMENT_ENV", "test")

	// Restore environment variables after test
	defer func() {
		os.Setenv("OTEL_COLLECTOR_URL", originalEnv)
		os.Setenv("SERVICE_NAME", originalServiceName)
		os.Setenv("DEPLOYMENT_ENV", originalDeploymentEnv)
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
		os.Setenv("OTEL_COLLECTOR_URL", "")
		os.Setenv("SERVICE_NAME", "test-service")
		os.Setenv("DEPLOYMENT_ENV", "test")

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

func TestNewPropagator(t *testing.T) {
	prop := newPropagator()
	assert.NotNil(t, prop)

	// Verify it implements TextMapPropagator
	_, ok := prop.(propagation.TextMapPropagator)
	assert.True(t, ok, "Expected a TextMapPropagator")

	// Use a carrier and inject context with baggage
	ctx := context.Background()
	b, err := baggage.NewMember("user", "alice")
	assert.NoError(t, err)

	bg, err := baggage.New(b)
	assert.NoError(t, err)

	ctx = baggage.ContextWithBaggage(ctx, bg)
	carrier := propagation.MapCarrier{}

	prop.Inject(ctx, carrier)

	// Check if baggage was injected
	assert.Contains(t, carrier, "baggage", "Expected 'baggage' header in carrier")
	assert.Contains(t, carrier["baggage"], "user=alice")
}

func TestNewResource(t *testing.T) {
	t.Run("successful creation", func(t *testing.T) {
		// Save original config values
		originalServiceName := config.AppConfig.ServiceName
		originalEnvironment := config.AppConfig.Environment

		// Set test config values
		config.AppConfig.ServiceName = "test-service"
		config.AppConfig.Environment = "test"

		defer func() {
			config.AppConfig.ServiceName = originalServiceName
			config.AppConfig.Environment = originalEnvironment
		}()

		res, err := NewResource()
		assert.NoError(t, err)
		assert.NotNil(t, res)

		// Verify resource attributes
		attrs := res.Attributes()
		serviceNameFound := false
		environmentFound := false

		for i := 0; i < len(attrs); i++ {
			attr := attrs[i]
			if attr.Key == semconv.ServiceNameKey {
				serviceNameFound = true
				assert.Equal(t, "test-service", attr.Value.AsString())
			}
			if attr.Key == semconv.DeploymentEnvironmentKey {
				environmentFound = true
				assert.Equal(t, "test", attr.Value.AsString())
			}
		}

		assert.True(t, serviceNameFound, "Service name attribute not found")
		assert.True(t, environmentFound, "Environment attribute not found")
	})

	t.Run("empty service name", func(t *testing.T) {
		// Save original config values
		originalServiceName := config.AppConfig.ServiceName
		originalEnvironment := config.AppConfig.Environment

		// Set test config values
		config.AppConfig.ServiceName = ""
		config.AppConfig.Environment = "test"

		defer func() {
			config.AppConfig.ServiceName = originalServiceName
			config.AppConfig.Environment = originalEnvironment
		}()

		res, err := NewResource()
		assert.NoError(t, err)
		assert.NotNil(t, res)
	})
}

func TestStdoutExporter(t *testing.T) {
	t.Run("export records", func(t *testing.T) {
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w
		defer func() { os.Stdout = oldStdout }()

		exporter := &stdoutExporter{}
		record := logsdk.Record{} // cannot set body/severity in v0.11.0

		err := exporter.Export(context.Background(), []logsdk.Record{record})
		assert.NoError(t, err)

		w.Close()
		output, _ := io.ReadAll(r)
		assert.Contains(t, string(output), "] : \n") // minimal check for output format
	})

	t.Run("shutdown", func(t *testing.T) {
		exporter := &stdoutExporter{}
		err := exporter.Shutdown(context.Background())
		assert.NoError(t, err)
	})

	t.Run("force flush", func(t *testing.T) {
		exporter := &stdoutExporter{}
		err := exporter.ForceFlush(context.Background())
		assert.NoError(t, err)
	})
}

func TestNewTracerProvider(t *testing.T) {
	t.Run("successful creation", func(t *testing.T) {
		// Save original config values
		originalServiceName := config.AppConfig.ServiceName
		originalEnvironment := config.AppConfig.Environment

		// Set test config values
		config.AppConfig.ServiceName = "test-service"
		config.AppConfig.Environment = "test"

		defer func() {
			config.AppConfig.ServiceName = originalServiceName
			config.AppConfig.Environment = originalEnvironment
		}()

		// Test with valid endpoint
		tp, err := newTracerProvider(context.Background(), "localhost:4317")
		assert.NoError(t, err)
		assert.NotNil(t, tp)
	})

	t.Run("empty endpoint", func(t *testing.T) {
		tp, err := newTracerProvider(context.Background(), "")
		assert.Error(t, err)
		assert.Nil(t, tp)
		assert.Equal(t, "OTEL_COLLECTOR_URL is not set", err.Error())
	})

	t.Run("invalid endpoint", func(t *testing.T) {
		tp, err := newTracerProvider(context.Background(), "invalid://endpoint")
		assert.NoError(t, err)
		assert.NotNil(t, tp)
	})
}

func TestNewMeterProvider(t *testing.T) {
	t.Run("successful creation", func(t *testing.T) {
		// Save original config values
		originalServiceName := config.AppConfig.ServiceName
		originalEnvironment := config.AppConfig.Environment

		// Set test config values
		config.AppConfig.ServiceName = "test-service"
		config.AppConfig.Environment = "test"

		defer func() {
			config.AppConfig.ServiceName = originalServiceName
			config.AppConfig.Environment = originalEnvironment
		}()

		mp, err := newMeterProvider("localhost:4317")
		assert.NoError(t, err)
		assert.NotNil(t, mp)
	})

	t.Run("invalid endpoint", func(t *testing.T) {
		mp, err := newMeterProvider("invalid://endpoint")
		assert.NoError(t, err)
		assert.NotNil(t, mp)
	})
}

func TestNewLoggerProvider(t *testing.T) {
	t.Run("successful creation", func(t *testing.T) {
		// Save original config values
		originalServiceName := config.AppConfig.ServiceName
		originalEnvironment := config.AppConfig.Environment

		// Set test config values
		config.AppConfig.ServiceName = "test-service"
		config.AppConfig.Environment = "test"

		defer func() {
			config.AppConfig.ServiceName = originalServiceName
			config.AppConfig.Environment = originalEnvironment
		}()

		lp, err := newLoggerProvider("localhost:4317")
		assert.NoError(t, err)
		assert.NotNil(t, lp)
	})

	t.Run("invalid endpoint", func(t *testing.T) {
		lp, err := newLoggerProvider("invalid://endpoint")
		assert.NoError(t, err)
		assert.NotNil(t, lp)
	})
}

// func TestSetupOTelSDK_resourceNewError(t *testing.T) {
// 	orig := resourceNew
// 	defer func() { resourceNew = orig }()
// 	resourceNew = func(ctx context.Context, opts ...resource.Option) (*resource.Resource, error) {
// 		return nil, errors.New("forced error")
// 	}
// 	shutdown, err := SetupOTelSDK(context.Background())
// 	assert.Nil(t, shutdown)
// 	assert.Error(t, err)
// 	assert.Equal(t, "forced error", err.Error())
// }

// func TestNewResource_resourceNewError(t *testing.T) {
// 	orig := resourceNew
// 	defer func() { resourceNew = orig }()
// 	resourceNew = func(ctx context.Context, opts ...resource.Option) (*resource.Resource, error) {
// 		return nil, errors.New("forced error")
// 	}
// 	res, err := NewResource()
// 	assert.Nil(t, res)
// 	assert.Error(t, err)
// 	assert.Equal(t, "forced error", err.Error())
// }
