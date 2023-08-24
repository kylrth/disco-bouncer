package main

import (
	"context"
	"fmt"
	"os"

	"github.com/cobaltspeech/log"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spf13/cobra"

	"github.com/kylrth/disco-bouncer/internal/db"
	"github.com/kylrth/disco-bouncer/internal/server"
)

func withLAndDB(f func(log.Logger, *pgxpool.Pool, []string) error) func(*cobra.Command, []string) {
	return withLogger(func(l log.Logger, args []string) error {
		pgURI := os.Getenv("DATABASE_URL")
		err := db.ApplyMigrations(pgURI)
		if err != nil {
			return fmt.Errorf("apply database migrations: %w", err)
		}

		pool, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))
		if err != nil {
			return fmt.Errorf("connect to database: %w", err)
		}
		defer pool.Close()

		return f(l, pool, args)
	})
}

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Serve the disco-bouncer data manager API",
	Args:  cobra.NoArgs,
	Run: withLAndDB(func(l log.Logger, p *pgxpool.Pool, args []string) error {
		return serve(l, p)
	}),
}

func serve(l log.Logger, pool *pgxpool.Pool) error {
	app := fiber.New()
	app.Use(logger.New(logger.Config{Output: os.Stderr}))

	server.AddAuthHandlers(l, app, pool)
	server.AddCRUDHandlers(l, app, pool)

	return app.Listen(":80")
}