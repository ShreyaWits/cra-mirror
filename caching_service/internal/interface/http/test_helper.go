package http

import (
	"io"
	"log/slog"

	"redis-service/internal/app"

	tracenoop "go.opentelemetry.io/otel/trace/noop"
)

// setupTestContainer initializes the test container with no-op tracer and logger
func setupTestContainer() {
	// Create a no-op tracer
	tracer := tracenoop.NewTracerProvider().Tracer("test")
	// Create a new logger that discards output
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	app.Di = &app.Container{
		Tracer: tracer,
		Logger: logger,
	}
}
