package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/bwmarrin/discordgo"
	"github.com/cobaltspeech/log"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spf13/cobra"

	"github.com/kylrth/disco-bouncer/internal/db"
	"github.com/kylrth/disco-bouncer/internal/server"
	"github.com/kylrth/disco-bouncer/pkg/bouncerbot"
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
	aTable := db.NewAdminTable(l, pool)
	uTable := db.NewUserTable(l, pool)

	app := fiber.New()
	app.Use(logger.New(logger.Config{Output: os.Stderr}))
	server.AddAuthHandlers(l, app, pool, aTable)
	server.AddCRUDHandlers(l, app, uTable)

	token := os.Getenv("DISCORD_TOKEN")
	if token == "disable" {
		l.Info("msg", "running without Discord bot")

		return app.Listen(":80")
	}

	bot, err := bouncerbot.New(l, token, uTable)
	if err != nil {
		return fmt.Errorf("set up Discord bot: %w", err)
	}
	err = addGuildInfo(l, bot)
	if err != nil {
		return fmt.Errorf("add guild info: %w", err)
	}

	err = bot.Open()
	if err != nil {
		return fmt.Errorf("start Discord bot: %w", err)
	}
	defer func() {
		err = bot.Close()
		if err != nil {
			l.Error("msg", "error closing Discord connection", "error", err)
		}
	}()
	l.Info("msg", "started bot; press Ctrl+C to exit")

	return app.Listen(":80")
}

const guildInfoFile = "/data/guildinfo.json"

func addGuildInfo(l log.Logger, bot *bouncerbot.Bot) error {
	b, err := os.ReadFile(guildInfoFile)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			l.Info("msg", "guild info file does not exist")

			// We'll set the bot to listen for guild messages this time, so we can retrieve the
			// guild info as soon as possible.
			bot.Identify.Intents |= discordgo.IntentGuildMessages
			// We'll give the bot a callback that saves the guild info to disk once it's received.
			bot.AddGuildInfoCallback(saveGuildInfo(l))

			return nil
		}

		return err
	}

	var info bouncerbot.GuildInfo

	err = json.Unmarshal(b, &info)
	if err != nil {
		return err
	}

	bot.Guild = &info

	return nil
}

func saveGuildInfo(l log.Logger) func(*bouncerbot.GuildInfo) {
	return func(info *bouncerbot.GuildInfo) {
		b, err := json.Marshal(info)
		if err != nil {
			l.Error("msg", "failed to marshal guild info to save it", "error", err)

			return
		}

		err = os.WriteFile(guildInfoFile, b, 0o600)
		if err != nil {
			l.Error("msg", "failed to write guild info file", "error", err)
		}
	}
}
