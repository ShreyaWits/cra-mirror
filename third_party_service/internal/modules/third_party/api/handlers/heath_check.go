package handlers

import (
	"net/http"
	"third_party_service/internal/modules/third_party/api"

	"github.com/gofiber/fiber/v2"
)

func HandleHealthCheck(c *fiber.Ctx) error {

	return api.Response{
		Code: http.StatusOK,
		Msg:  "Service OK",
		Data: nil,
	}

}
