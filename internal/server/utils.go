package server

import (
	"net/http"

	"github.com/cobaltspeech/log"
	"github.com/gofiber/fiber/v2"
)

// HiddenError signifies that an error should be logged but not reported to the client.
type HiddenErr struct {
	error
}

func serverError(l log.Logger, c *fiber.Ctx, msg string, err error) error {
	l.Error("msg", "internal server error", "message", msg, "error", err)

	if _, ok := err.(HiddenErr); !ok { //nolint:errorlint // just checking top error type
		msg += ": " + err.Error()
	}

	return c.Status(http.StatusInternalServerError).SendString(msg)
}
