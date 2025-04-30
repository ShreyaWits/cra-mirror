package app

import (
	"github.com/gofiber/fiber/v2"
)

func SetupRouter() *fiber.App {

	//fiber registeration
	app := fiber.New()

	//modules registration
	RegisterModules(app)
	return app
}
