package handlers

import (
	"net/http"

	v1 "thirdparty_service/api/rest/v1"

	"github.com/gofiber/fiber/v2"
)

func HandleHealthCheck(c *fiber.Ctx) error {

	return v1.Response{
		Code: http.StatusOK,
		Msg:  "Service OK",
	}
}
