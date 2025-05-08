package di

import (
	"messaging_service/internal/config"
	"messaging_service/internal/messaging_service/handler"
	"messaging_service/pkg/logger"
)

// Container holds all the dependencies for the application
type Container struct {
	// Config
	Config *config.Config

	// Key Manager
	MessagingHandler *handler.MessagingHandler
}

// NewContainer creates a new dependency injection container
func NewContainer() (*Container, error) {
	container := &Container{}
	cfg, err := config.LoadConfig()
	// Load configuration
	if err != nil {
		return nil, err
	}
	container.Config = cfg

	logger.InitLogger()
	// Messaging service server instance
	msgHandler := handler.MessagingHandler{}
	container.MessagingHandler = &msgHandler

	return container, nil
}

// // SetupRoutes sets up all the routes for the application
// func (c *Container) SetupRoutes(app *fiber.App) {

// 	routes.SetupAPIRoutes(app, c.EncryptionHandler)
// }
