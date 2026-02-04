package db

import (
	"context"
	"fmt"
	"log/slog"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"doligo_001/internal/infrastructure/config"
)

// InitDatabase initializes and returns a GORM database connection
func InitDatabase(ctx context.Context, cfg *config.DatabaseConfig) (*gorm.DB, string, error) {
	var dialector gorm.Dialector
	dsn := ""

	switch cfg.Type {
	case "postgres":
		dsn = fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
			cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Name, cfg.SSLMode)
		dialector = postgres.Open(dsn)
	case "mysql":
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Name)
		dialector = mysql.Open(dsn)
	default:
		return nil, "", fmt.Errorf("unsupported database type: %s", cfg.Type)
	}

	db, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		return nil, "", fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, "", fmt.Errorf("failed to get generic database object: %w", err)
	}

	// Set connection pool parameters
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	// Ping the database to verify connection
	if err = sqlDB.PingContext(ctx); err != nil {
		return nil, "", fmt.Errorf("failed to ping database: %w", err)
	}

	slog.Debug("Database connection established", "type", cfg.Type)
	return db, dsn, nil
}

// Ping checks the database connection.
func Ping(ctx context.Context, db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get generic database object: %w", err)
	}
	return sqlDB.PingContext(ctx)
}
