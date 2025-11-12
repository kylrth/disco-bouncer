package server

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/cobaltspeech/log"
	"github.com/gofiber/fiber/v2"
	"github.com/kylrth/disco-bouncer/internal/db"
)

func AddCRUDHandlers(l log.Logger, app *fiber.App, table *db.UserTable) {
	app.Get("/api/users", GetAllUsers(l, table))
	app.Get("/api/users/:id", GetUser(l, table))
	app.Post("/api/users", CreateUser(l, table))
	app.Put("/api/users/:id", UpdateUser(l, table))
	app.Delete("/api/users/:id", DeleteUser(l, table))
}

// GetAllUsers sends the entire users table, possibly filtered by provided query parameters.
func GetAllUsers(l log.Logger, table *db.UserTable) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// get filter values
		keyHash := c.Query("keyHash", "")

		users, err := table.GetUsers(c.Context(), db.WithKeyHash(keyHash))
		if err != nil {
			return serverError(l, c, "Database error", err)
		}

		return c.JSON(users)
	}
}

// GetUser sends the information of the user with the specified ID.
func GetUser(l log.Logger, table *db.UserTable) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, err := strconv.Atoi(c.Params("id"))
		if err != nil {
			l.Debug("msg", "invalid ID", "error", err, "id", c.Params("id"))

			return c.Status(http.StatusBadRequest).SendString(fmt.Sprintf("Invalid ID: %v", err))
		}

		u, err := table.GetUser(c.Context(), id)
		if err != nil {
			if errors.Is(err, db.ErrNoUser) {
				return c.Status(http.StatusNotFound).SendString("User not found")
			}

			return serverError(l, c, "Database error", err)
		}

		return c.JSON(u)
	}
}

// CreateUser creates a new user and returns the ID.
func CreateUser(l log.Logger, table *db.UserTable) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var user db.User
		err := c.BodyParser(&user)
		if err != nil {
			return c.Status(http.StatusBadRequest).SendString(err.Error())
		}

		user.ID, err = table.CreateUser(c.Context(), &user)
		if err != nil {
			return serverError(l, c, "Database error", err)
		}

		return c.JSON(&user)
	}
}

// UpdateUser updates the information for a user.
func UpdateUser(l log.Logger, table *db.UserTable) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, err := strconv.Atoi(c.Params("id"))
		if err != nil {
			l.Debug("msg", "invalid ID", "error", err, "id", c.Params("id"))

			return c.Status(http.StatusBadRequest).SendString(fmt.Sprintf("Invalid ID: %v", err))
		}

		var user db.User
		err = c.BodyParser(&user)
		if err != nil {
			return c.Status(http.StatusBadRequest).SendString(err.Error())
		}

		user.ID = id

		err = table.UpdateUser(c.Context(), &user)
		if err != nil {
			if errors.Is(err, db.ErrNoUser) {
				return c.Status(http.StatusNotFound).SendString("User not found")
			}

			return serverError(l, c, "Database error", err)
		}

		return c.JSON(user)
	}
}

// DeleteUser removes a user by ID.
func DeleteUser(l log.Logger, table *db.UserTable) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, err := strconv.Atoi(c.Params("id"))
		if err != nil {
			l.Debug("msg", "invalid ID", "error", err, "id", c.Params("id"))

			return c.Status(http.StatusBadRequest).SendString(fmt.Sprintf("Invalid ID: %v", err))
		}

		err = table.DeleteUser(c.Context(), id)
		if err != nil {
			if errors.Is(err, db.ErrNoUser) {
				return c.Status(http.StatusNotFound).SendString("User not found")
			}

			return serverError(l, c, "Database error", err)
		}

		return c.SendString("User deleted successfully")
	}
}
