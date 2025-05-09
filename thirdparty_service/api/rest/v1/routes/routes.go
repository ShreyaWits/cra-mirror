package routes

import (
	"thirdparty_service/api/rest/v1/handlers"
	"thirdparty_service/api/rest/v1/middleware"
	"thirdparty_service/api/rest/v1/schemas"
	"thirdparty_service/internal/di"
	"thirdparty_service/internal/services"

	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(router fiber.Router) error {
	v1 := router.Group("/api/v1")

	if err := setupHealthCheckRoutes(v1); err != nil {
		return err
	}

	if err := setupUserRoutes(v1); err != nil {
		return err
	}
	if err := setupWebhookRoutes(v1); err != nil {
		return err
	}

	return nil
}

func setupHealthCheckRoutes(router fiber.Router) error {
	router.Get(
		"/healthz",
		handlers.HandleHealthCheck,
	)
	return nil
}

func setupUserRoutes(router fiber.Router) error {
	userPrefix := router.Group("/user")

	var userService *services.UserService
	if err := di.Container.Invoke(func(u *services.UserService) {
		userService = u
	}); err != nil {
		return err
	}

	userHandlers := handlers.NewUserHandler(userService)

	userPrefix.Post("/",
		middleware.ValidatorMiddleware[schemas.CreateUserRequest](),
		userHandlers.HandleCreateUser,
	)

	return nil
}

func setupWebhookRoutes(router fiber.Router) error {
	userPrefix := router.Group("/webhook")

	var webhookService services.WebhookService
	if err := di.Container.Invoke(func(s services.WebhookService) {
		webhookService = s
	}); err != nil {
		return err
	}

	webhookHandlers := handlers.NewWebhookHandler(webhookService)

	userPrefix.Post("/", webhookHandlers.HandleWebhook)

	return nil
}
