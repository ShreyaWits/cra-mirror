package handlers

import (
	"net/http"

	v1 "thirdparty_service/api/rest/v1"
	"thirdparty_service/api/rest/v1/schemas"
	"thirdparty_service/internal/services"

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
	var req schemas.CreateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return v1.Error{
			Code: http.StatusBadRequest,
			Err:  "failed to parse request body",
		}
	}

	user, err := uh.UserService.CreateUser(req)
	if err != nil {
		return v1.Error{
			Code: http.StatusInternalServerError,
			Err:  "failed to create user",
		}
	}

	return v1.Response{
		Code: http.StatusOK,
		Msg:  "User Created",
		Data: schemas.UserResponse{
			ID:        user.ID,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Email:     user.Email,
		},
	}
}
