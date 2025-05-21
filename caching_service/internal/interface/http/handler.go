package http

import (
	"redis-service/internal/app"

	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel/attribute"
)

func HealthCheck(c *fiber.Ctx) error {
	tracer := app.Di.Tracer
	logger := app.Di.Logger

	ctx := c.Context()
	_, span := tracer.Start(ctx, "HealthCheck")
	defer span.End()

	span.SetAttributes(
		attribute.String("health.check", "true"),
	)

	logger.Info("health check requested")

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status": "ok",
	})
}
