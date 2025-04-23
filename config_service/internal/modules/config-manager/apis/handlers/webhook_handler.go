package handler

import (
	"nps-config-service/internal/modules/config-manager/apis/dtos"

	"github.com/gofiber/fiber/v2"
)

func (h *Handler) RegisterWebhook(c *fiber.Ctx) error {
	// Implement the logic to register a webhook
	// This is just a placeholder for the actual implementation
	var configData dtos.RegisterWebhookRequest
	if err := c.BodyParser(&configData); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid config data",
		})
	}
	res, err := h.Service.RegisterWebhookService(configData)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.JSON(res)
}
func (h *Handler) GetWebhooks(c *fiber.Ctx) error {
	env := c.Params("environment")
	service := c.Params("service")

	res, err := h.Service.GetWebhooks(env, service)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.JSON(res)
}
func (h *Handler) DeleteWebhook(c *fiber.Ctx) error {
	env := c.Params("environment")
	service := c.Params("service")

	var deleteReq dtos.RegisterWebhookRequest
	if err := c.BodyParser(&deleteReq); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid delete request",
		})
	}

	res, err := h.Service.DeleteWebhook(env, service, deleteReq.URL, deleteReq.Method)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(res)
}
