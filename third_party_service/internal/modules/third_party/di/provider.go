package di

import (
	"fmt"
	"third_party_service/internal/common/middleware"
	"third_party_service/internal/modules/third_party/repositories"
	"third_party_service/internal/modules/third_party/services"

	"github.com/gofiber/fiber/v2"

	"go.uber.org/dig"
)

var (
	Container = dig.New()
)

// InitializeApp initializes the application with all dependencies
func InitializeApp() (*fiber.App, error) {
	app := fiber.New(fiber.Config{ErrorHandler: middleware.FiberErrorHandler})

	// initialize application dependencies
	if err := initializeDeps(); err != nil {
		return nil, fmt.Errorf("failed to initialize app dependencies: %w", err)
	}

	return app, nil
}

func initializeDeps() error {

	thirdPartyRepositoryRes := repositories.NewThirdPartyRepository()
	if err := Container.Provide(func() repositories.ThirdPartyRepository {
		return thirdPartyRepositoryRes
	}); err != nil {
		return err
	}

	userService := services.NewThirdPartyService(thirdPartyRepositoryRes)
	if err := Container.Provide(func() *services.ThirdPartyService {
		return userService
	}); err != nil {
		return err
	}

	return nil
}
