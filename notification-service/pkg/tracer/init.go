package tracer

import (
	"context"
	"fmt"
	"os"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
	"github.com/sirupsen/logrus"
)

func InitTracer(serviceName string) (func(context.Context) error, error) {
	logrus.Info("Initializing Jaeger tracer...")

	jagerURL := os.Getenv("JAGER_URL")

	exporter, err := jaeger.New(jaeger.WithCollectorEndpoint(
		jaeger.WithEndpoint(jagerURL),
	))
	if err != nil {
		logrus.Errorf("Failed to create Jaeger exporter: %v", err)
		return nil, fmt.Errorf("creating Jaeger exporter: %w", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(serviceName),
		)),
	)

	otel.SetTracerProvider(tp)
	logrus.Info("Jaeger tracer initialized successfully!")

	return tp.Shutdown, nil
}
