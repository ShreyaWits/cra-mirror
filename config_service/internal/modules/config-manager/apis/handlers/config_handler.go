package handler

import (
	"fmt"
	common "nps-config-service/internal/common/errors"
	"nps-config-service/internal/modules/config-manager/apis/dtos"
	"nps-config-service/internal/modules/config-manager/services"
	"os"

	"github.com/gofiber/fiber/v2"
)

type ConfigHandler struct {
	Service services.IConfigService
}

func NewConfigHandler(service services.IConfigService) *ConfigHandler {
	return &ConfigHandler{Service: service}
}

func (h *ConfigHandler) AdminHandler(c *fiber.Ctx) error {
	contextData, ok := c.Locals("contextData").(*dtos.AdminDto)
	fmt.Print(contextData)

	if !ok {
		return c.Status(fiber.StatusBadRequest).JSON(common.ThrowError(fiber.StatusBadRequest, "CNF004"))
	}

	// Load the admin secret from .env
	adminSecret := os.Getenv("ADMIN_SECRET")
	if adminSecret == "" {
		return c.Status(fiber.StatusBadRequest).JSON(common.ThrowError(fiber.StatusBadRequest, "CNF005")) // Admin secret not configured
	}

	userName := os.Getenv("USERNAME")
	if userName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(common.ThrowError(fiber.StatusBadRequest, "CNF009")) // Admin secret not configured
	}

	if contextData.Username != userName {
		return c.Status(fiber.StatusBadRequest).JSON(common.ThrowError(fiber.StatusBadRequest, "CNF008")) // Invalid username
	}

	// Check if the password matches
	if contextData.Password != adminSecret {
		return c.Status(fiber.StatusBadRequest).JSON(common.ThrowError(fiber.StatusBadRequest, "CNF006")) // Invalid password
	}

	// Load JWT secret
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return c.Status(fiber.StatusBadRequest).JSON(common.ThrowError(fiber.StatusBadRequest, "CNF007"))
	}

	responseData, responseError := h.Service.AdminService(contextData, jwtSecret)

	if responseError != nil {
		return responseError
	}
	return c.Status(fiber.StatusOK).JSON(responseData)
}

func (h *ConfigHandler) StoreConfigHandler(c *fiber.Ctx) error {
	environment := c.Params("environment")
	serviceName := c.Params("service")

	var configData map[string]interface{}
	if err := c.BodyParser(&configData); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(common.ThrowError(fiber.StatusBadRequest, "CNF001"))
	}

	responseData, responseError := h.Service.StoreConfigService(environment, serviceName, configData)

	if responseError != nil {
		return responseError
	}
	return c.JSON(responseData)
}

func (h *ConfigHandler) GetfullConfig(c *fiber.Ctx) error {
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

func (h *ConfigHandler) GetByValue(c *fiber.Ctx) error {
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

func (h *ConfigHandler) GetByMetadata(c *fiber.Ctx) error {
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
