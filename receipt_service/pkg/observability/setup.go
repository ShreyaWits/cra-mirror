package observability

import (
	"context"
	"os"
	"time"

	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"google.golang.org/grpc"
)

// NewTracerProvider exports the tracer provider setup
func NewTracerProvider(ctx context.Context, url string) (*trace.TracerProvider, error) {
	return newTracerProvider(ctx, url)
}

// NewMeterProvider exports the meter provider setup
func NewMeterProvider(url string) (*metric.MeterProvider, error) {
	return newMeterProvider(url)
}

// NewLoggerProvider exports the logger provider setup
func NewLoggerProvider(url string) (*log.LoggerProvider, error) {
	return newLoggerProvider(url)
}

func getEnvOrDefault(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func newTracerProvider(ctx context.Context, tracerProviderUrl string) (*trace.TracerProvider, error) {
	traceExporter, err := otlptracegrpc.New(
		ctx,
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithEndpoint(tracerProviderUrl),
	)
	if err != nil {
		return nil, err
	}
	serviceName := getEnvOrDefault("OTEL_SERVICE_NAME", "nps-reciept-service")
	serviceVersion := getEnvOrDefault("OTEL_SERVICE_VERSION", "0.1.0")
	environment := getEnvOrDefault("OTEL_ENVIRONMENT", "development")

	tracerProvider := trace.NewTracerProvider(
		trace.WithBatcher(traceExporter, trace.WithBatchTimeout(time.Second)),
		trace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(serviceName),
			semconv.ServiceVersionKey.String(serviceVersion),
			semconv.DeploymentEnvironmentKey.String(environment),
		)),
	)
	return tracerProvider, nil
}

func newMeterProvider(metricProviderUrl string) (*metric.MeterProvider, error) {
	metricExporter, err := otlpmetricgrpc.New(
		context.Background(),
		otlpmetricgrpc.WithInsecure(),
		otlpmetricgrpc.WithEndpoint(metricProviderUrl),
	)
	if err != nil {
		return nil, err
	}
	serviceName := getEnvOrDefault("OTEL_SERVICE_NAME", "my-service")
	serviceVersion := getEnvOrDefault("OTEL_SERVICE_VERSION", "0.1.0")
	environment := getEnvOrDefault("OTEL_ENVIRONMENT", "development")

	meterProvider := metric.NewMeterProvider(
		metric.WithReader(metric.NewPeriodicReader(metricExporter, metric.WithInterval(3*time.Second))),
		metric.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(serviceName+"-metrics"),
			semconv.ServiceVersionKey.String(serviceVersion),
			semconv.DeploymentEnvironmentKey.String(environment),
		)),
	)
	return meterProvider, nil
}

func newLoggerProvider(url string) (*log.LoggerProvider, error) {
	logExporter, err := otlploggrpc.New(
		context.Background(),
		otlploggrpc.WithInsecure(),
		otlploggrpc.WithEndpoint(url),
		otlploggrpc.WithDialOption(grpc.WithBlock()),
	)
	if err != nil {
		return nil, err
	}
	serviceName := getEnvOrDefault("OTEL_SERVICE_NAME", "my-service")
	serviceVersion := getEnvOrDefault("OTEL_SERVICE_VERSION", "0.1.0")
	environment := getEnvOrDefault("OTEL_ENVIRONMENT", "development")

	loggerProvider := log.NewLoggerProvider(
		log.WithProcessor(log.NewBatchProcessor(logExporter)),
		log.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(serviceName),
			semconv.ServiceVersionKey.String(serviceVersion),
			semconv.DeploymentEnvironmentKey.String(environment),
		)),
	)
	return loggerProvider, nil
}
