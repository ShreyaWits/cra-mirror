package middlewares

// import (
// 	"protected_link/internal/modules/authentication/services"

// 	"github.com/gofiber/fiber/v2"
// )

// func OTPMiddleware(authService *services.AuthenticationService) fiber.Handler {
// 	return func(c *fiber.Ctx) error {
// 		userID := c.Query("user_id")
// 		inputOTP := c.Query("otp")

// 		if userID == "" || inputOTP == "" {
// 			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
// 				"error": "user_id and otp are required",
// 			})
// 		}

// 		valid, err := authService.VerifyOTP(userID, inputOTP)
// 		if err != nil {
// 			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
// 				"error": err.Error(),
// 			})
// 		}

// 		if !valid {
// 			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
// 				"error": "Invalid OTP",
// 			})
// 		}

// 		return c.Next()
// 	}
// };;;
