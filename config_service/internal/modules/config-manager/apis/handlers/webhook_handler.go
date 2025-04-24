package handler

import (
	"nps-config-service/internal/modules/config-manager/apis/dtos"
	"nps-config-service/internal/modules/config-manager/services"

	"github.com/gofiber/fiber/v2"
)

type WebhookHandler struct {
	Service services.IWebhookService
}

func NewWebhookHandler(service services.IWebhookService) *WebhookHandler {
	return &WebhookHandler{Service: service}
}
func (h *WebhookHandler) RegisterWebhook(c *fiber.Ctx) error {

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
func (h *WebhookHandler) GetWebhooks(c *fiber.Ctx) error {
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
func (h *WebhookHandler) DeleteWebhook(c *fiber.Ctx) error {
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
