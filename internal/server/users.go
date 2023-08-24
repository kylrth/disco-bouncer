package server

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/cobaltspeech/log"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kylrth/disco-bouncer/internal/db"
)

func AddCRUDHandlers(l log.Logger, app *fiber.App, pool *pgxpool.Pool) {
	table := db.NewUserTable(l, pool)

	app.Get("/api/users", GetAllUsers(table))
	app.Get("/api/users/:id", GetUser(table))
	app.Post("/api/users", CreateUser(table))
	app.Put("/api/users/:id", UpdateUser(table))
	app.Delete("/api/users/:id", DeleteUser(table))
}

// GetAllUsers sends the entire users table.
func GetAllUsers(table *db.UserTable) fiber.Handler {
	return func(c *fiber.Ctx) error {
		users, err := table.GetUsers(c.Context())
		if err != nil {
			return serverError(c, fmt.Sprintf("Database error: %v", err))
		}

		return c.JSON(users)
	}
}

// GetUser sends the information of the user with the specified ID.
func GetUser(table *db.UserTable) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, err := strconv.Atoi(c.Params("id"))
		if err != nil {
			return c.Status(http.StatusBadRequest).SendString("Invalid ID")
		}

		u, err := table.GetUser(c.Context(), id)
		if err != nil {
			if errors.Is(err, db.ErrNoUser) {
				return c.Status(http.StatusNotFound).SendString("User not found")
			}
			return serverError(c, fmt.Sprintf("Database error: %v", err))
		}

		return c.JSON(u)
	}
}

// CreateUser creates a new user and returns the ID.
func CreateUser(table *db.UserTable) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var user db.User
		if err := c.BodyParser(&user); err != nil {
			return c.Status(http.StatusBadRequest).SendString(err.Error())
		}

		var err error
		user.ID, err = table.CreateUser(c.Context(), &user)
		if err != nil {
			return serverError(c, fmt.Sprintf("Database error: %v", err))
		}

		return c.JSON(&user)
	}
}

// UpdateUser updates the information for a user.
func UpdateUser(table *db.UserTable) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, err := strconv.Atoi(c.Params("id"))
		if err != nil {
			return c.Status(http.StatusBadRequest).SendString("Invalid ID")
		}

		var user db.User
		if err := c.BodyParser(&user); err != nil {
			return c.Status(http.StatusBadRequest).SendString(err.Error())
		}

		user.ID = id

		err = table.UpdateUser(c.Context(), &user)
		if err != nil {
			if errors.Is(err, db.ErrNoUser) {
				return c.Status(http.StatusNotFound).SendString("User not found")
			}
			return serverError(c, fmt.Sprintf("Database error: %v", err))
		}

		return c.JSON(user)
	}
}

// DeleteUser removes a user by ID.
func DeleteUser(table *db.UserTable) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, err := strconv.Atoi(c.Params("id"))
		if err != nil {
			return c.Status(http.StatusBadRequest).SendString("Invalid ID")
		}

		err = table.DeleteUser(c.Context(), id)
		if err != nil {
			if errors.Is(err, db.ErrNoUser) {
				return c.Status(http.StatusNotFound).SendString("User not found")
			}
			return serverError(c, fmt.Sprintf("Database error: %v", err))
		}

		return c.SendString("User deleted successfully")
	}
}
