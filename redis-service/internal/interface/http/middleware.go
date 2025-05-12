package http

import (
	"redis-service/internal/app"

	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

func TraceMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := c.Context()
		tracer := app.Di.Tracer
		logger := app.Di.Logger

		spanName := c.Path()
		spanCtx, span := tracer.Start(ctx, spanName)
		defer span.End()

		// Add request attributes
		span.SetAttributes(
			attribute.String("http.method", c.Method()),
			attribute.String("http.url", c.Path()),
			attribute.String("http.user_agent", c.Get("User-Agent")),
		)

		// Log request
		logger.Info("incoming request",
			"method", c.Method(),
			"path", c.Path(),
			"user_agent", c.Get("User-Agent"),
		)

		// Store span context in Fiber context
		c.Locals("spanContext", spanCtx)

		// Continue with the request
		err := c.Next()

		// Add response attributes
		statusCode := c.Response().StatusCode()
		span.SetAttributes(
			attribute.Int("http.status_code", statusCode),
		)

		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			logger.Error("request failed",
				"error", err.Error(),
				"method", c.Method(),
				"path", c.Path(),
				"status_code", statusCode,
			)
		} else {
			logger.Info("request completed",
				"method", c.Method(),
				"path", c.Path(),
				"status_code", statusCode,
			)
		}

		return err
	}
}
