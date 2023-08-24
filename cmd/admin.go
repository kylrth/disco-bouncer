package main

import (
	"context"
	"errors"

	"github.com/cobaltspeech/log"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kylrth/disco-bouncer/internal/db"
	"github.com/spf13/cobra"
)

var adminCmd = &cobra.Command{
	Use:   "admin",
	Short: "Manage admins",
}

func init() {
	adminCmd.AddCommand(
		adminSetPass,
		adminDeleteCmd,
	)
}

var adminSetPass = &cobra.Command{
	Use:   "setpass",
	Short: "Set the password for a new or existing admin",
	Args:  cobra.ExactArgs(2),
	Run: withLAndDB(func(l log.Logger, pool *pgxpool.Pool, args []string) error {
		table := db.NewAdminTable(l, pool)

		user := args[0]
		pass := args[1]

		err := table.ChangePassword(context.Background(), user, pass)
		if errors.Is(err, db.ErrNoUser) {
			l.Info("msg", "admin does not exist, creating it", "admin", user)
			err = table.AddAdmin(context.Background(), user, pass)
		}
		if err != nil {
			return err
		}

		l.Info("msg", "set password for admin", "admin", user)

		return nil
	}),
}

var adminDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete an admin from the database",
	Args:  cobra.ExactArgs(1),
	Run: withLAndDB(func(l log.Logger, pool *pgxpool.Pool, args []string) error {
		table := db.NewAdminTable(l, pool)

		user := args[0]

		return table.DeleteAdmin(context.Background(), user)
	}),
}
