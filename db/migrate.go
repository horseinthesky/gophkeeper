package db

import (
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

const (
	migrationsURL = "file://db/migrations"
)

func RunDBMigration(to string) error {
	migration, err := migrate.New(migrationsURL, to)
	if err != nil {
		return fmt.Errorf("cannot create new migrate instance: %w", err)
	}

	if err = migration.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrate up: %w", err)
	}

	return nil
}
