package observability

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"encryption_microservice/internal/config"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	// "go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/trace"

	// "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/sdk/resource"

	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
)

// formatEndpoint removes any protocol prefix and ensures the endpoint is in the correct format
func formatEndpoint(url string) string {
	// Remove http:// or https:// if present
	url = strings.TrimPrefix(url, "http://")
	url = strings.TrimPrefix(url, "https://")
	return url
}

// SetupOTelSDK bootstraps the OpenTelemetry pipeline.
// If it does not return an error, make sure to call shutdown for proper cleanup.
func SetupOTelSDK(ctx context.Context, env *config.EnvConfig, cfg *config.Config) (shutdown func(context.Context) error, err error) {
	var shutdownFuncs []func(context.Context) error

	// Print environment info for debugging
	fmt.Printf("Setting up OpenTelemetry SDK with observability URL: %s\n", cfg.OtelCollectorGrpcEndpoint)

	// shutdown calls cleanup functions registered via shutdownFuncs.
	// The errors from the calls are joined.
	// Each registered cleanup will be invoked once.
	shutdown = func(ctx context.Context) error {
		var err error
		for _, fn := range shutdownFuncs {
			err = errors.Join(err, fn(ctx))
		}
		shutdownFuncs = nil
		return err
	}

	// handleErr calls shutdown for cleanup and makes sure that all errors are returned.
	handleErr := func(inErr error) {
		err = errors.Join(inErr, shutdown(ctx))
	}

	// Create a resource with service information
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(config.SERVICE_NAME),
			semconv.ServiceVersionKey.String(config.SERVICE_VERSION),
			semconv.DeploymentEnvironmentKey.String(env.Environment),
		),
	)
	if err != nil {
		handleErr(fmt.Errorf("failed to create resource: %w", err))
		return
	}

	// Set up propagator.
	prop := newPropagator()
	otel.SetTextMapPropagator(prop)

	// Set up trace provider.
	tracerProvider, err := newTracerProvider(ctx, env, res, cfg)
	if err != nil {
		fmt.Printf("Warning: Failed to create tracer provider: %v\n", err)
		// Fallback to a noop tracer provider
		noopTracerProvider := trace.NewTracerProvider()
		otel.SetTracerProvider(noopTracerProvider)
	} else {
		shutdownFuncs = append(shutdownFuncs, tracerProvider.Shutdown)
		otel.SetTracerProvider(tracerProvider)
		fmt.Println("Tracer provider successfully initialized")
	}

	// Set up meter provider.
	meterProvider, err := newMeterProvider(env, res, cfg)
	if err != nil {
		fmt.Printf("Warning: Failed to create meter provider: %v\n", err)
		// Fallback to a noop meter provider
		noopMeterProvider := metric.NewMeterProvider()
		otel.SetMeterProvider(noopMeterProvider)
	} else {
		shutdownFuncs = append(shutdownFuncs, meterProvider.Shutdown)
		otel.SetMeterProvider(meterProvider)
		fmt.Println("Meter provider successfully initialized")
	}

	// Set up logger provider.
	loggerProvider, err := newLoggerProvider(ctx, env, res, cfg)
	if err != nil {
		fmt.Printf("Warning: Failed to create logger provider: %v\n", err)
		// Fallback to a noop logger provider
		noopLoggerProvider := log.NewLoggerProvider()
		global.SetLoggerProvider(noopLoggerProvider)
	} else {
		shutdownFuncs = append(shutdownFuncs, loggerProvider.Shutdown)
		global.SetLoggerProvider(loggerProvider)
		fmt.Println("Logger provider successfully initialized")
	}

	return
}

func newPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
}

func newTracerProvider(ctx context.Context, env *config.EnvConfig, res *resource.Resource, cfg *config.Config) (*trace.TracerProvider, error) {
	endpoint := formatEndpoint(cfg.OtelCollectorGrpcEndpoint)
	fmt.Printf("Setting up trace provider with endpoint: %s\n", endpoint)

	traceExporter, err := otlptracegrpc.New(
		ctx,
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithEndpoint(endpoint),
		otlptracegrpc.WithDialOption(grpc.WithTransportCredentials(insecure.NewCredentials())),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create trace exporter: %w", err)
	}

	// Default is 5s. Set to 1s for demonstrative purposes.
	tracerProvider := trace.NewTracerProvider(
		trace.WithBatcher(traceExporter,
			// Default is 5s. Set to 1s for demonstrative purposes.
			trace.WithBatchTimeout(time.Second)),
		trace.WithResource(res),
		trace.WithSampler(trace.AlwaysSample()),
	)
	return tracerProvider, nil
}

func newMeterProvider(env *config.EnvConfig, res *resource.Resource, cfg *config.Config) (*metric.MeterProvider, error) {
	endpoint := formatEndpoint(cfg.OtelCollectorGrpcEndpoint)
	fmt.Printf("Setting up meter provider with endpoint: %s\n", endpoint)

	metricExporter, err := otlpmetricgrpc.New(
		context.Background(),
		otlpmetricgrpc.WithInsecure(),
		otlpmetricgrpc.WithEndpoint(endpoint),
		otlpmetricgrpc.WithDialOption(grpc.WithTransportCredentials(insecure.NewCredentials())),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create metric exporter: %w", err)
	}

	meterProvider := metric.NewMeterProvider(
		metric.WithReader(metric.NewPeriodicReader(metricExporter,
			// Default is 1m. Set to 3s for demonstrative purposes.
			metric.WithInterval(3*time.Second))),
		metric.WithResource(res),
	)
	return meterProvider, nil
}

// StdoutExporter is a log exporter that writes to stdout for debugging
type StdoutExporter struct{}

func (e *StdoutExporter) Export(ctx context.Context, records []log.Record) error {
	for _, rec := range records {
		fmt.Fprintf(os.Stdout, "[OTLP LOG] [%s] %s: %s\n",
			rec.Timestamp().Format(time.RFC3339),
			rec.SeverityText(),
			rec.Body().AsString(),
		)
	}
	return nil
}

func (e *StdoutExporter) Shutdown(ctx context.Context) error {
	return nil
}

func (e *StdoutExporter) ForceFlush(ctx context.Context) error {
	return nil
}

func newLoggerProvider(ctx context.Context, env *config.EnvConfig, res *resource.Resource, cfg *config.Config) (*log.LoggerProvider, error) {
	endpoint := formatEndpoint(cfg.OtelCollectorGrpcEndpoint)
	fmt.Printf("Setting up logger provider with endpoint: %s\n", endpoint)

	// Create OTLP log exporter for sending to collector
	logExporter, err := otlploggrpc.New(
		ctx,
		otlploggrpc.WithInsecure(),
		otlploggrpc.WithEndpoint(endpoint),
		otlploggrpc.WithDialOption(grpc.WithTransportCredentials(insecure.NewCredentials())),
		otlploggrpc.WithDialOption(grpc.WithBlock()),                // Wait for connection
		otlploggrpc.WithDialOption(grpc.WithTimeout(5*time.Second)), // Timeout after 5 seconds
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP log exporter: %w", err)
	}

	// Create a custom stdout exporter for better debugging
	customStdoutExporter := &StdoutExporter{}

	// Create the log provider with both exporters
	loggerProvider := log.NewLoggerProvider(
		log.WithProcessor(log.NewBatchProcessor(logExporter)),
		log.WithProcessor(log.NewBatchProcessor(customStdoutExporter)),
		log.WithResource(res),
	)

	// Log a message that initialization was successful (using fmt instead of logger)
	fmt.Printf("OpenTelemetry logger provider successfully initialized for service: %s\n", config.SERVICE_NAME)

	return loggerProvider, nil
}
