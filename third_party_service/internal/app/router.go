package app

import (
	"third_party_service/internal/modules/third_party/api/handlers"
	"third_party_service/internal/modules/third_party/di"
	"third_party_service/internal/modules/third_party/services"

	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(router fiber.Router) error {
	v1 := router.Group("/api/v1")

	if err := setupHealthCheckRoutes(v1); err != nil {
		return err
	}

	if err := setupHealthCheckRoutes(v1); err != nil {
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

func setupThirdPartyRoutes(router fiber.Router) {
	thirdPartyPrefix := router.Group("/third-party")

	var thirdPartyService *services.ThirdPartyService
	if err := di.Container.Invoke(func(u *services.ThirdPartyService) {
		thirdPartyService = u
	}); err != nil {
		// Log the error or handle it as needed
		panic(err) // or use a logger to log the error
	}

	userHandlers := handlers.NewThirdPartyHandler(thirdPartyService)

	thirdPartyPrefix.Post("/send-email",
		// middleware.ValidatorMiddleware[schemas.CreateUserRequest](),
		userHandlers.HandlerSendEmail,
	)

}
