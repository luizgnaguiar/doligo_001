package db

import (
	"database/sql"
	"embed"
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/mysql" // Import for MySQL driver
	_ "github.com/golang-migrate/migrate/v4/database/postgres" // Import for PostgreSQL driver
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed ../migrations/*.sql
var fs embed.FS

// RunMigrations applies database migrations
func RunMigrations(db *sql.DB, dbType, dsn string) error {
	sourceDriver, err := iofs.New(fs, "../migrations")
	if err != nil {
		return fmt.Errorf("failed to create migration source driver: %w", err)
	}

	// golang-migrate needs the DSN to connect to the database.
	// The dbType prefix is important for `migrate` to know which driver to use.
	databaseURL := fmt.Sprintf("%s://%s", dbType, dsn)

	m, err := migrate.NewWithSourceInstance("iofs", sourceDriver, databaseURL)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	if err == migrate.ErrNoChange {
		log.Println("No migrations to apply.")
	} else {
		log.Println("Database migrations applied successfully.")
	}

	return nil
}