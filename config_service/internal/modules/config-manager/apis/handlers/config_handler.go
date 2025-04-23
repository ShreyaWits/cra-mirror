package handler

import (
	common "nps-config-service/internal/common/errors"
	"nps-config-service/internal/modules/config-manager/services"
	"nps-config-service/internal/modules/config-manager/utils"

	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	Service *services.ConfigService
}

func NewHandler(service *services.ConfigService) *Handler {
	return &Handler{Service: service}
}

// StoreConfigHandler handles the storing of configuration data
func (h *Handler) StoreConfigHandler(c *fiber.Ctx) error {
	environment := c.Params("environment")
	serviceName := c.Params("service")

	contextData := c.Locals("contextData")
	configData, ok := contextData.(*map[string]any)
	if !ok {
		utils.SendError(c, fiber.StatusBadRequest, "Invalid config data", common.ThrowError("CNF001"))
		return nil
	}

	responseData, responseError := h.Service.StoreConfigService(environment, serviceName, *configData)
	if responseError != nil {
		utils.SendError(c, fiber.StatusInternalServerError, "Failed to store config", responseError.Error())
		return nil
	}

	utils.SendSuccess(c, fiber.StatusOK, "Config stored successfully", responseData)
	return nil
}

func (h *Handler) GetfullConfig(c *fiber.Ctx) error {
	environment := c.Params("environment")
	serviceName := c.Params("service")

	config, err := h.Service.GetConfigService(serviceName, environment)
	if err != nil {
		utils.SendError(c, fiber.StatusInternalServerError, "Failed to fetch config", err.Error())
		return nil
	}

	utils.SendSuccess(c, fiber.StatusOK, "Config fetched successfully", config)
	return nil
}

func (h *Handler) GetByValue(c *fiber.Ctx) error {
	environment := c.Params("environment")
	serviceName := c.Params("service")
	key := c.Params("key")

	value, err := h.Service.GetConfigValueService(serviceName, environment, key)
	if err != nil {
		utils.SendError(c, fiber.StatusInternalServerError, "Failed to get config value", err.Error())
		return nil
	}

	utils.SendSuccess(c, fiber.StatusOK, "Value fetched successfully", fiber.Map{key: value})
	return nil
}

func (h *Handler) GetByMetadata(c *fiber.Ctx) error {
	environment := c.Params("environment")
	serviceName := c.Params("service")

	metadata, err := h.Service.GetConfigMetadataService(serviceName, environment)
	if err != nil {
		utils.SendError(c, fiber.StatusInternalServerError, "Failed to fetch metadata", err.Error())
		return nil
	}

	utils.SendSuccess(c, fiber.StatusOK, "Metadata fetched successfully", metadata)
	return nil
}
