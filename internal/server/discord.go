package server

import (
	"errors"
	"net/http"

	"github.com/cobaltspeech/log"
	"github.com/gofiber/fiber/v2"
	"github.com/kylrth/disco-bouncer/pkg/bouncerbot"
)

func AddDiscordHandlers(l log.Logger, app *fiber.App, dg *bouncerbot.Bot) {
	app.Post("/api/discord/migrate", MigrateUser(l, dg))
}

// Migration defines a user that needs to be assigned a cohort role. The name should match the
// display name of a current Discord user on the server.
type Migration struct {
	Name string `json:"name"`
	Year string `json:"role"`
}

// Migrator is something that can migrate a user to the new cohort by name.
type Migrator interface {
	Migrate(name, year string) error
}

func MigrateUser(l log.Logger, dg *bouncerbot.Bot) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var migration Migration
		err := c.BodyParser(&migration)
		if err != nil {
			return c.Status(http.StatusBadRequest).SendString(err.Error())
		}

		err = dg.Migrate(migration.Name, migration.Year)
		if err != nil {
			if errors.Is(err, bouncerbot.ErrNoUser) {
				return c.Status(http.StatusNotFound).SendString("User not found")
			}
			if errors.Is(err, bouncerbot.ErrUnknownYear) {
				return c.Status(http.StatusBadRequest).SendString("Cohort year not found")
			}

			return serverError(l, c, "Discord error", err)
		}

		return nil
	}
}
