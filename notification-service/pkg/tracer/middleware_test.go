package tracer

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func TestTracingMiddleware(t *testing.T) {
	// Setup in-memory span exporter
	exporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	defer func() { _ = tp.Shutdown(context.Background()) }()
	otel.SetTracerProvider(tp)

	tracer := otel.Tracer("test_tracer")
	app := fiber.New()

	// Create a manual span before the request
	_, preReqSpan := tracer.Start(context.Background(), "simple_test_span")
	preReqSpan.End()

	// Use middleware
	app.Use(TracingMiddleware(tracer))

	// Add test route
	app.Get("/test", func(c *fiber.Ctx) error {
		ctx := c.UserContext() // Get the context from Fiber
		_, span := tracer.Start(ctx, "handler_span")
		defer span.End()
	
		return c.SendString("Hello")
	})

	// Trigger the test request
	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)

	// Force flush spans
	require.NoError(t, tp.ForceFlush(context.Background()))

	// Fetch spans
	spans := exporter.GetSpans()

	// Log span details
	for i, s := range spans {
		t.Logf("Span %d: Name=%s, TraceID=%s, SpanID=%s, ParentSpanID=%s", i, s.Name,
			s.SpanContext.TraceID().String(),
			s.SpanContext.SpanID().String(),
			s.Parent.SpanID().String())
	}

	// We expect 3 spans:
	// 1. simple_test_span (manually created)
	// 2. span created by middleware
	// 3. span created by handler
	require.Len(t, spans, 3, "Expected 3 spans, got %d", len(spans))

	// Validate span names
	var names []string
	for _, s := range spans {
		names = append(names, s.Name)
	}
	require.Contains(t, names, "simple_test_span")
require.Contains(t, names, "/test")
require.Contains(t, names, "handler_span")

}
