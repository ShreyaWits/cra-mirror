package config

import (
	"context"
	"log"
	"os"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func InitTelemetry() func() {
	ctx := context.Background()

	// Get collector endpoints from environment variables
	tracesEndpoint := os.Getenv("OTEL_EXPORTER_OTLP_TRACES_ENDPOINT")
	metricsEndpoint := os.Getenv("OTEL_EXPORTER_OTLP_METRICS_ENDPOINT")

	if tracesEndpoint == "" || metricsEndpoint == "" {
		log.Printf("Warning: OpenTelemetry endpoints not configured. Observability will be limited.")
		return func() {}
	}

	// Create gRPC connection to collector for metrics
	conn, err := grpc.DialContext(ctx, metricsEndpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		log.Printf("Failed to create gRPC connection to collector: %v", err)
		return func() {}
	}

	// Create OTLP exporter
	exporter, err := otlpmetricgrpc.New(ctx,
		otlpmetricgrpc.WithGRPCConn(conn),
	)
	if err != nil {
		log.Printf("Failed to create OTLP exporter: %v", err)
		return func() {}
	}

	// Get service name and version from environment
	serviceName := os.Getenv("OTEL_SERVICE_NAME")
	if serviceName == "" {
		serviceName = "nps-receipt-service"
	}

	serviceVersion := os.Getenv("OTEL_SERVICE_VERSION")
	if serviceVersion == "" {
		serviceVersion = "1.0.0"
	}

	// Create resource with service information
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
			semconv.ServiceVersion(serviceVersion),
		),
	)
	if err != nil {
		log.Printf("Failed to create resource: %v", err)
		return func() {}
	}

	// Get metrics interval from environment or use default
	metricsInterval := 10 * time.Second
	if intervalStr := os.Getenv("OTEL_METRICS_INTERVAL"); intervalStr != "" {
		if interval, err := time.ParseDuration(intervalStr); err == nil {
			metricsInterval = interval
		}
	}

	// Create MeterProvider
	mp := metric.NewMeterProvider(
		metric.WithResource(res),
		metric.WithReader(
			metric.NewPeriodicReader(
				exporter,
				metric.WithInterval(metricsInterval),
			),
		),
	)

	// Set global MeterProvider
	otel.SetMeterProvider(mp)

	// Return cleanup function
	return func() {
		if err := mp.Shutdown(ctx); err != nil {
			log.Printf("Failed to shutdown meter provider: %v", err)
		}
		if err := conn.Close(); err != nil {
			log.Printf("Failed to close gRPC connection: %v", err)
		}
	}
}
