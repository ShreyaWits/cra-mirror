package handler

import (
	"nps-config-service/internal/modules/config-manager/apis/dtos"
	"nps-config-service/internal/modules/config-manager/utils"

	customErr "nps-config-service/internal/common/errors"

	"github.com/gofiber/fiber/v2"
)

func (h *Handler) RegisterWebhook(c *fiber.Ctx) error {
	contextData := c.Locals("contextData")
	bodyMap, ok := contextData.(*map[string]any)
	if !ok {
		utils.SendError(c, fiber.StatusBadRequest, "Invalid config data", "Context data missing or invalid")
		return nil
	}

	var configData dtos.RegisterWebhookRequest
	if err := utils.MapToStruct(*bodyMap, &configData); err != nil {
		utils.SendError(c, fiber.StatusBadRequest, "Invalid config structure", err.Error())
		return nil
	}

	res, err := h.Service.RegisterWebhookService(configData)
	if err != nil {
		if conflictErr, ok := err.(*customErr.ConflictError); ok {
			utils.SendError(c, fiber.StatusConflict, "Webhook already exists", conflictErr.Error())
		} else {
			utils.SendError(c, fiber.StatusInternalServerError, "Failed to register webhook", err.Error())
		}
		return nil
	}

	utils.SendSuccess(c, fiber.StatusOK, "Webhook registered successfully", res)
	return nil
}

func (h *Handler) GetWebhooks(c *fiber.Ctx) error {
	env := c.Params("environment")
	service := c.Params("service")

	res, err := h.Service.GetWebhooks(env, service)
	if err != nil {
		utils.SendError(c, fiber.StatusNotFound, "Failed to retrieve webhooks", err.Error())
		return nil
	}

	utils.SendSuccess(c, fiber.StatusOK, "Webhooks retrieved successfully", res)
	return nil
}

func (h *Handler) DeleteWebhook(c *fiber.Ctx) error {
	env := c.Params("environment")
	service := c.Params("service")

	contextData := c.Locals("contextData")
	bodyMap, ok := contextData.(*map[string]any)
	if !ok {
		utils.SendError(c, fiber.StatusBadRequest, "Invalid delete request", "Context data missing or invalid")
		return nil
	}

	var deleteReq dtos.RegisterWebhookRequest
	if err := utils.MapToStruct(*bodyMap, &deleteReq); err != nil {
		utils.SendError(c, fiber.StatusBadRequest, "Invalid delete structure", err.Error())
		return nil
	}

	res, err := h.Service.DeleteWebhook(env, service, deleteReq.URL, deleteReq.Method)
	if err != nil {
		utils.SendError(c, fiber.StatusInternalServerError, "Failed to delete webhook", err.Error())
		return nil
	}

	utils.SendSuccess(c, fiber.StatusOK, "Webhook deleted successfully", res)
	return nil
}
