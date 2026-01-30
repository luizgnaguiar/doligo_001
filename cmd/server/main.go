package main

import (
	"context"
	"database/sql"
	"doligo_001/internal/infrastructure/config"
	"doligo_001/internal/infrastructure/database"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"gorm.io/gorm"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

func main() {
	// Setup structured logging
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	slog.Info("Starting Doligo ERP/CRM Service...")

	// Load Configuration
	cfg := config.Load()
	slog.Info("Configuration loaded")

	// Connect to Database
	db, err := database.New(cfg)
	if err != nil {
		slog.Error("Database connection failed", "error", err)
		os.Exit(1)
	}
	slog.Info("Database connection successful")

	sqlDB, err := db.DB()
	if err != nil {
		slog.Error("Failed to get underlying *sql.DB", "error", err)
		os.Exit(1)
	}

	// Run Migrations
	if err := runMigrations(sqlDB); err != nil {
		slog.Error("Migrations failed", "error", err)
		os.Exit(1)
	}

	// Setup Echo Webserver
	e := echo.New()

	// Middleware
	e.Use(middleware.Recover())
	// A simple logger middleware for now. A custom one with request_id will be added later.
	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogURI:    true,
		LogStatus: true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			slog.Info("request",
				"URI", v.URI,
				"status", v.Status)
			return nil
		},
	}))

	// Health Check Endpoints
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	e.GET("/ready", func(c echo.Context) error {
		ctx, cancel := context.WithTimeout(c.Request().Context(), 5*time.Second)
		defer cancel()

		if err := sqlDB.PingContext(ctx); err != nil {
			slog.Error("Readiness check failed: database ping failed", "error", err)
			return c.JSON(http.StatusServiceUnavailable, map[string]string{"status": "database unreachable"})
		}
		return c.JSON(http.StatusOK, map[string]string{"status": "ready"})
	})

	// Start Server
	slog.Info("Starting server on port 8080")
	if err := e.Start(":8080"); err != nil {
		slog.Error("Server failed to start", "error", err)
		os.Exit(1)
	}
}

func runMigrations(db *sql.DB) error {
	slog.Info("Applying database migrations...")

	sourceInstance, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return fmt.Errorf("failed to create migration source instance: %w", err)
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create migration database instance: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", sourceInstance, "postgres", driver)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	slog.Info("Migrations applied successfully")
	return nil
}

// Placeholder for a function to get a *gorm.DB instance for a specific schema
func DBForSchema(db *gorm.DB, schema string) *gorm.DB {
	return db.Session(&gorm.Session{
		Context: context.WithValue(context.Background(), "gorm:schema", schema),
	})
}
