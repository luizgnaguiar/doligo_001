package db

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/mysql" // Import for MySQL driver
	_ "github.com/golang-migrate/migrate/v4/database/postgres" // Import for PostgreSQL driver
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/rs/zerolog/log"

	"doligo_001/internal/infrastructure"
)

// RunMigrations applies database migrations
func RunMigrations(ctx context.Context, db *sql.DB, dbType, dsn string) error {
	// Although the context is passed, the golang-migrate library's Up() method
	// does not directly accept a context. The context's primary role here is
	// to signal overall application shutdown; however, its direct effect on
	// the migration process itself is limited by the library's API.
	sourceDriver, err := iofs.New(infrastructure.MigrationsFS, "migrations")
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
		log.Debug().Msg("No migrations to apply.")
	} else {
		log.Info().Msg("Database migrations applied successfully.")
	}

	return nil
}