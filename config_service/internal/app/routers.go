package app

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func SetupRouter() *fiber.App {

	//fiber registeration
	app := fiber.New()
	var ConfigDefault = recover.Config{
		Next:              nil,
		EnableStackTrace:  true,
		StackTraceHandler: recover.ConfigDefault.StackTraceHandler,
	}
	app.Use(recover.New(ConfigDefault))
	//modules registration
	RegisterModules(app)
	return app
}
