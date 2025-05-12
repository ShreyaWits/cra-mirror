package app

import (
	"log/slog"
	"redis-service/pkg/config"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
)

type Container struct {
	Tracer trace.Tracer
	Logger *slog.Logger
	Meter  metric.Meter
}

var Di *Container

func InitContainer() {
	// Get the global tracer provider
	tracerProvider := otel.GetTracerProvider()
	if tracerProvider == nil {
		// If no tracer provider is set, use no-op provider
		tracerProvider = noop.NewTracerProvider()
		otel.SetTracerProvider(tracerProvider)
	}

	tracer := tracerProvider.Tracer(config.SERVICE_NAME)
	logger := otelslog.NewLogger(config.SERVICE_NAME)
	metric := otel.Meter(config.SERVICE_NAME)
	Di = &Container{
		Tracer: tracer,
		Logger: logger,
		Meter:  metric,
	}
}
