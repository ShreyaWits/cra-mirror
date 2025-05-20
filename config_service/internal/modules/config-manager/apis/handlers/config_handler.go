package handler

import (
	common "nps-config-service/internal/common/errors"
	"nps-config-service/internal/modules/config-manager/services"
	"nps-config-service/internal/modules/config-manager/utils"
	"nps-config-service/pkg/observability"

	"github.com/gofiber/fiber/v2"
)

type ConfigHandler struct {
	Service            services.IConfigService
	ObservabilityStack *observability.ObservabilityStack
}

func NewConfigHandler(service services.IConfigService, ObservabilityStack *observability.ObservabilityStack) *ConfigHandler {
	return &ConfigHandler{Service: service}
}

func (h *ConfigHandler) StoreConfigHandler(c *fiber.Ctx) error {
	ctx := c.UserContext() // Get the context.

	environment := c.Params("environment")
	serviceName := c.Params("service")

	contextData := c.Locals("contextData")
	configData, ok := contextData.(*map[string]any)
	if !ok {
		err := common.ThrowError(fiber.StatusBadRequest, "CNF001")
		utils.SendError(c, fiber.StatusBadRequest, "Invalid config data", err.Error())
		return nil
	}

	responseData, responseError := h.Service.StoreConfigService(ctx, environment, serviceName, *configData)
	if responseError != nil {
		utils.SendError(c, fiber.StatusInternalServerError, "Failed to store config", responseError.ErrorMessage)
		return nil
	}

	utils.SendSuccess(c, responseData.StatusCode, responseData.Message, responseData.Data)
	return nil
}

func (h *ConfigHandler) GetfullConfig(c *fiber.Ctx) error {
	ctx := c.UserContext()

	environment := c.Params("environment")
	serviceName := c.Params("service")

	config, err := h.Service.GetConfigService(ctx, serviceName, environment)
	if err != nil {
		utils.SendError(c, fiber.StatusInternalServerError, "Failed to fetch config", err.Error())
		return nil
	}

	utils.SendSuccess(c, fiber.StatusOK, "Config fetched successfully", config)
	return nil
}

func (h *ConfigHandler) GetByValue(c *fiber.Ctx) error {
	ctx := c.UserContext() // Get the context.

	environment := c.Params("environment")
	serviceName := c.Params("service")
	key := c.Params("key")

	value, err := h.Service.GetConfigValueService(ctx, serviceName, environment, key)
	if err != nil {
		utils.SendError(c, fiber.StatusInternalServerError, "Failed to get config value", err.Error())
		return nil
	}

	utils.SendSuccess(c, fiber.StatusOK, "Value fetched successfully", fiber.Map{key: value})
	return nil
}
