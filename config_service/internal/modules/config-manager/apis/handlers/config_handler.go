package handler

import (
	"fmt"
	"nps-config-service/common"
	"nps-config-service/internal/modules/config-manager/apis/dtos"
	"nps-config-service/internal/modules/config-manager/services"
	"os"

	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	Service *services.ConfigService
}

func NewHandler(service *services.ConfigService) *Handler {
	return &Handler{Service: service}
}

func (h *Handler) AdminHandler(c *fiber.Ctx) error {
	contextData, ok := c.Locals("contextData").(*dtos.AdminDto)
	fmt.Print(contextData)

	if !ok{
		return c.Status(fiber.StatusBadRequest).JSON(common.ThrowError(fiber.StatusBadRequest,"CNF004"))
	}
	if contextData.Username != "admin" {
		return c.Status(fiber.StatusBadRequest).JSON(common.ThrowError(fiber.StatusBadRequest, "CNF008")) // Invalid username
	}

	// Load the admin secret from .env
	adminSecret := os.Getenv("ADMIN_SECRET")
	if adminSecret == "" {
		return c.Status(fiber.StatusBadRequest).JSON(common.ThrowError(fiber.StatusBadRequest, "CNF005")) // Admin secret not configured
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

	if responseError != nil{
		return responseError
	}
	return c.Status(fiber.StatusOK).JSON(responseData)
}

func (h *Handler) StoreConfigHandler(c *fiber.Ctx) error {
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
