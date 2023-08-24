package server

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
)

func serverError(c *fiber.Ctx, err string) error {
	return c.Status(http.StatusInternalServerError).SendString(err)
}
