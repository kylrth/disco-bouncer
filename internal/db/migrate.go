package db

import (
	"embed"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed migrations
var migrationsFS embed.FS

// ApplyMigrations ensures all migrations are applied to the database at dbURI.
func ApplyMigrations(dbURI string) error {
	d, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return fmt.Errorf("load migrations: %w", err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", d, dbURI)
	if err != nil {
		return fmt.Errorf("connect: %w", err)
	}

	err = m.Up()
	if errors.Is(err, migrate.ErrNoChange) {
		return nil
	}

	return err
}
