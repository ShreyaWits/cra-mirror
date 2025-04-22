package handler

import (
	"nps-config-service/internal/modules/config-manager/services"

	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	Service *services.ConfigService
}

func NewHandler(service *services.ConfigService) *Handler {
	return &Handler{Service: service}
}

func (h *Handler) StoreConfigHandler(c *fiber.Ctx) error {
	environment := c.Params("environment")
	serviceName := c.Params("service")

	var configData map[string]interface{}
	if err := c.BodyParser(&configData); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid config data",
		})
	}

	responseData, responseError := h.Service.StoreConfigService(environment, serviceName, configData)

	if responseError != nil {
		return responseError
	}
	return c.JSON(responseData)
}

func (h *Handler) GetfullConfig(c *fiber.Ctx) error {
	environment := c.Params("environment")
	serviceName := c.Params("service")

	config, err := h.Service.GetConfigService(serviceName, environment)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(config)
}

func (h *Handler) GetByValue(c *fiber.Ctx) error {
	environment := c.Params("environment")
	serviceName := c.Params("service")
	key := c.Params("key")

	value, err := h.Service.GetConfigValueService(serviceName, environment, key)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		key: value,
	})
}

func (h *Handler) GetByMetadata(c *fiber.Ctx) error {
	environment := c.Params("environment")
	serviceName := c.Params("service")

	metadata, err := h.Service.GetConfigMetadataService(serviceName, environment)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(metadata)
}
