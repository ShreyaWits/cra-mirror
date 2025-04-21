package handler

import (
	"nps-config-service/internal/config-manager/services"

	"github.com/gofiber/fiber/v2"
)

func StoreConfigHandler(c *fiber.Ctx) error {
	environment := c.Params("environment")
	serviceName := c.Params("service")

	var configData map[string]interface{}
	if err := c.BodyParser(&configData); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid config data",
		})
	}

	responseData, responseError := services.StoreConfigService(environment, serviceName, configData)

	if responseError != nil {
		return responseError
	}
	return c.JSON(responseData)
}

func GetfullConfig(c *fiber.Ctx) error {
	environment := c.Params("environment")
	serviceName := c.Params("service")

	config, err := services.GetConfigService(serviceName, environment)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(config)
}

func GetByValue(c *fiber.Ctx) error {
	environment := c.Params("environment")
	serviceName := c.Params("service")
	key := c.Params("key")

	value, err := services.GetConfigValueService(serviceName, environment, key)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		key: value,
	})
}

func GetByMetadata(c *fiber.Ctx) error {
	environment := c.Params("environment")
	serviceName := c.Params("service")

	metadata, err := services.GetConfigMetadataService(serviceName, environment)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(metadata)
}
