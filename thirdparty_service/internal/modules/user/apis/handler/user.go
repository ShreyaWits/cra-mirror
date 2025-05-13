package handlers

import (
	"net/http"
	app_dto "thirdparty_service/internal/dtos"
	"thirdparty_service/internal/modules/user/apis/dtos"
	services "thirdparty_service/internal/modules/user/service"

	"github.com/gofiber/fiber/v2"
)

type UserHandler struct {
	UserService *services.UserService
}

func NewUserHandler(userService *services.UserService) *UserHandler {
	return &UserHandler{
		UserService: userService,
	}
}
func (uh *UserHandler) HandleCreateUser(c *fiber.Ctx) error {
	var req dtos.CreateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return app_dto.Error{
			Code: http.StatusBadRequest,
			Err:  "failed to parse request body",
		}
	}

	user, err := uh.UserService.CreateUser(req)
	if err != nil {
		return app_dto.Error{
			Code: http.StatusInternalServerError,
			Err:  "failed to create user",
		}
	}

	return app_dto.Response{
		Code: http.StatusOK,
		Msg:  "User Created",
		Data: dtos.UserResponse{
			ID:        user.ID,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Email:     user.Email,
		},
	}
}
