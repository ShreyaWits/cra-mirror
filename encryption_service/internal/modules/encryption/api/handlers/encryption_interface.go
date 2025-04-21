package handlers

import "github.com/gofiber/fiber/v2"

type EncryptionHandler interface {
	// HandleEncrypt handles the encryption request
	HandleEncrypt(c *fiber.Ctx) error

	// HandleDecrypt handles the decryption request
	HandleDecrypt(c *fiber.Ctx) error

	// HandleGenerateEDEK handles the EDEK generation request
	HandleGenerateEDEK(c *fiber.Ctx) error

	HandleHealth(c *fiber.Ctx) error
}
