package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	stdlog "log" // Alias standard log to avoid conflicts with zerolog.log

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog/log"

	"doligo_001/internal/infrastructure/config"
	"doligo_001/internal/infrastructure/db"
	"doligo_001/internal/infrastructure/logger"
)

func main() {
	// Initialize logger first
	cfg, err := config.LoadConfig()
	if err != nil {
		stdlog.Fatalf("Error loading configuration: %v", err)
	}
	logger.InitLogger(cfg.Log.Level)

	log.Info().Msgf("Application Environment: %s", cfg.AppEnv)
	log.Info().Msgf("Listening on Port: %s", cfg.Port)
	log.Info().Msgf("Database Type: %s", cfg.Database.Type)
	log.Info().Msgf("Log Level: %s", cfg.Log.Level)


	// Initialize database
	gormDB, dsn, err := db.InitDatabase(&cfg.Database)
	if err != nil {
		log.Fatal().Err(err).Msg("Error initializing database")
	}
	defer func() {
		sqlDB, closeErr := gormDB.DB()
		if closeErr == nil {
			closeErr = sqlDB.Close()
		}
		if closeErr != nil {
			log.Error().Err(closeErr).Msg("Failed to close database connection")
		} else {
			log.Info().Msg("Database connection closed")
		}
	}()


	sqlDB, err := gormDB.DB()
	if err != nil {
		log.Fatal().Err(err).Msg("Error getting generic database object from GORM")
	}

	// Run migrations
	if err := db.RunMigrations(sqlDB, cfg.Database.Type, dsn); err != nil {
		log.Fatal().Err(err).Msg("Error running database migrations")
	}

	// Setup Echo
	e := echo.New()
	e.Use(middleware.Recover()) // Echo Recovery Middleware

	// Health check endpoint
	e.GET("/health", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	// Start server in a goroutine
	go func() {
		if err := e.Start(":" + cfg.Port); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("Shutting down the server")
		}
	}()

	// Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info().Msg("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second) // 10 seconds for graceful shutdown
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		log.Fatal().Err(err).Msg("Server forced to shutdown")
	}

	log.Info().Msg("Server gracefully stopped.")
}
