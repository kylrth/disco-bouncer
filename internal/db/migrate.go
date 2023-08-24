package db

import (
	"embed"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed migrations
var migrationsFS embed.FS

// ApplyMigrations applies all migrations to the database at dbURI.
func ApplyMigrations(dbURI string) error {
	d, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return fmt.Errorf("load migrations: %w", err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", d, dbURI)
	if err != nil {
		return fmt.Errorf("connect: %w", err)
	}

	return m.Up()
}
