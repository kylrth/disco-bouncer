package server

import (
	"errors"
	"net/http"

	"github.com/cobaltspeech/log"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gofiber/storage/postgres/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kylrth/disco-bouncer/internal/db"
	"golang.org/x/crypto/bcrypt"
)

func AddAuthHandlers(l log.Logger, app *fiber.App, pool *pgxpool.Pool) {
	sessionStore := session.New(session.Config{
		Storage: postgres.New(postgres.Config{
			DB:    pool,
			Table: "sessions",
		}),
	})

	table := db.NewAdminTable(l, pool)

	app.Post("/login", Login(table, sessionStore))
	app.Post("/logout", Logout(l, sessionStore))

	app.Use("/admin", AuthMiddleware(l, sessionStore))
	app.Post("/admin/pass", ChangePassword(table, sessionStore))

	app.Use("/api", AuthMiddleware(l, sessionStore))
}

func Login(table *db.AdminTable, sessionStore *session.Store) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var input struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}

		if err := c.BodyParser(&input); err != nil {
			return c.Status(http.StatusBadRequest).SendString("Failed to parse body")
		}

		success, err := table.CheckPassword(c.Context(), input.Username, input.Password)
		if err != nil {
			return serverError(c, "Failed to check password")
		}
		if !success {
			return c.Status(http.StatusUnauthorized).SendString("Invalid credentials")
		}

		sess, err := sessionStore.Get(c)
		if err != nil {
			return serverError(c, "Failed to initiate session")
		}
		sess.Set("username", input.Username)
		err = sess.Save()
		if err != nil {
			return serverError(c, "Failed to save session")
		}

		return c.SendString("Login successful")
	}
}

func Logout(l log.Logger, sessionStore *session.Store) fiber.Handler {
	return func(c *fiber.Ctx) error {
		sess, err := sessionStore.Get(c)
		if err != nil {
			return serverError(c, "Failed to get session")
		}

		var username string
		switch u := sess.Get("username").(type) {
		case string:
			username = u
		default:
			username = "unknown"
		}

		err = sess.Destroy()
		if err != nil {
			return serverError(c, "Failed to remove session")
		}

		l.Debug("msg", "logged out", "user", username)

		return c.SendString("Logout successful")
	}
}

func ChangePassword(table *db.AdminTable, sessionStore *session.Store) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// get the password
		var input struct {
			Old string `json:"old"`
			New string `json:"new"`
		}
		if err := c.BodyParser(&input); err != nil {
			return c.Status(http.StatusBadRequest).SendString("Failed to parse body")
		}

		// get the username from the logged in session
		sess, err := sessionStore.Get(c)
		if err != nil {
			return c.Status(http.StatusUnauthorized).SendString("Not authenticated")
		}
		var username string
		switch u := sess.Get("username").(type) {
		case string:
			username = u
		default:
			return c.Status(http.StatusUnauthorized).SendString("Invalid session data")
		}

		if len(input.New) < 8 {
			return c.Status(http.StatusBadRequest).SendString("Password too short")
		}

		success, err := table.CheckPassword(c.Context(), username, input.Old)
		if err != nil {
			return serverError(c, "Failed to check password")
		}
		if !success {
			return c.Status(http.StatusUnauthorized).SendString("Invalid credentials")
		}

		err = table.ChangePassword(c.Context(), username, input.New)
		if err != nil {
			if errors.Is(err, bcrypt.ErrPasswordTooLong) {
				return c.Status(http.StatusBadRequest).SendString("Password too long")
			}

			return serverError(c, "Failed to save password")
		}

		return c.SendString("Password updated successfully")
	}
}

func AuthMiddleware(l log.Logger, sessionStore *session.Store) fiber.Handler {
	return func(c *fiber.Ctx) error {
		sess, err := sessionStore.Get(c)
		if err != nil {
			return c.Status(http.StatusUnauthorized).SendString("Not authenticated")
		}

		var username string
		switch s := sess.Get("username").(type) {
		case string:
			username = s
		default:
			return c.Status(http.StatusUnauthorized).SendString("Invalid session data")
		}

		err = c.Next()

		r := c.Route()
		endpoint := r.Method + " " + r.Path

		l.Debug(
			"msg", "authenticated user accessed endpoint", "user", username, "endpoint", endpoint,
		)

		return err
	}
}
