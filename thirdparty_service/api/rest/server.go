package rest

import (
	"thirdparty_service/api/rest/v1/middleware"

	"github.com/gofiber/fiber/v2"
)

// InitializeApp initializes the application with all dependencies
func InitializeServer() (*fiber.App, error) {
	app := fiber.New(fiber.Config{ErrorHandler: middleware.FiberErrorHandler})

	// initialize application dependencies
	// if err := di.Container.Invoke(func(db *gorm.DB) {
	// 	// Run database migrations
	// 	err := models.RunMigrations(db)
	// 	if err != nil {
	// 		fmt.Printf("failed to run migrations: %v\n", err)
	// 		return
	// 	}
	// }); err != nil {
	// 	return nil, fmt.Errorf("failed to invoke db dependency: %v", err)
	// }

	return app, nil
}
