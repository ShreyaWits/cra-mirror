package tracer

import (
	"github.com/gofiber/fiber/v2"
	// "go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

func TracingMiddleware(tracer trace.Tracer) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx, span := tracer.Start(c.UserContext(), c.Path())
		defer span.End()

		c.SetUserContext(ctx)

		return c.Next()
	}
}
