package main

import (
	"context"
	"database/sql"
	"doligo_001/internal/api/middleware"
	"doligo_001/internal/domain/identity"
	"doligo_001/internal/infrastructure/config"
	"doligo_001/internal/infrastructure/database"
	"doligo_001/internal/infrastructure/repository/postgres"
	auth_usecase "doligo_001/internal/usecase/auth"
	"embed"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/golang-migrate/migrate/v4"
	migrate_postgres "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/labstack/echo/v4"
	echo_middleware "github.com/labstack/echo/v4/middleware"
)

//go:embed ../../migrations/*.sql
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
	e.Use(echo_middleware.Recover())
	e.Use(echo_middleware.RequestLoggerWithConfig(echo_middleware.RequestLoggerConfig{
		LogURI:    true,
		LogStatus: true,
		LogValuesFunc: func(c echo.Context, v echo_middleware.RequestLoggerValues) error {
			slog.Info("request",
				"URI", v.URI,
				"status", v.Status)
			return nil
		},
	}))

	// Repositories
	userRepo := postgres.NewGormUserRepository(db)
	roleRepo := postgres.NewGormRoleRepository(db)

	// Use Cases
	authUseCase := auth_usecase.NewAuthUseCase(userRepo, roleRepo)

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

	// Public routes
	e.POST("/login", func(c echo.Context) error {
		var req struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, "invalid request")
		}

		token, err := authUseCase.Login(c.Request().Context(), req.Username, req.Password)
		if err != nil {
			if errors.Is(err, auth_usecase.ErrInvalidCredentials) {
				return c.JSON(http.StatusUnauthorized, "invalid credentials")
			}
			slog.Error("login failed", "error", err)
			return c.JSON(http.StatusInternalServerError, "login failed")
		}

		return c.JSON(http.StatusOK, map[string]string{"token": token})
	})
	
	// API v1 group
	apiV1 := e.Group("/api/v1")
	apiV1.Use(middleware.AuthMiddleware)

	apiV1.GET("/products", func(c echo.Context) error {
		// Example of accessing user ID from context
		userID := c.Request().Context().Value(middleware.UserIDContextKey).(int64)
		slog.Info("products endpoint accessed", "user_id", userID)
		
		return c.JSON(http.StatusOK, []identity.Product{
			{ID: 1, Name: "Sample Product", Price: 99.99, CreatedBy: 1, UpdatedBy: 1},
		})
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

	driver, err := migrate_postgres.WithInstance(db, &migrate_postgres.Config{})
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



