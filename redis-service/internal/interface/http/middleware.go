package http

import (
	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

func TraceMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := c.Context()
		tracer := otel.Tracer("fiber-middleware")

		spanName := c.Path()
		spanCtx, span := tracer.Start(ctx, spanName)
		defer span.End()

		// Add request attributes
		span.SetAttributes(
			attribute.String("http.method", c.Method()),
			attribute.String("http.url", c.Path()),
			attribute.String("http.user_agent", c.Get("User-Agent")),
		)

		// Store span context in Fiber context
		c.Locals("spanContext", spanCtx)

		// Continue with the request
		err := c.Next()

		// Add response attributes
		span.SetAttributes(
			attribute.Int("http.status_code", c.Response().StatusCode()),
		)

		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}

		return err
	}
}
