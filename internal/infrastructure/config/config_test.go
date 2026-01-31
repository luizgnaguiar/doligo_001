package config

import (
	"os"
	"testing"
	"time"
)

func TestLoadConfig(t *testing.T) {
	// Clean up environment variables after test
	defer func() {
		os.Unsetenv("APP_ENV")
		os.Unsetenv("PORT")
		os.Unsetenv("DB_TYPE")
		os.Unsetenv("DB_HOST")
		os.Unsetenv("DB_PORT")
		os.Unsetenv("DB_USER")
		os.Unsetenv("DB_PASSWORD")
		os.Unsetenv("DB_NAME")
		os.Unsetenv("DB_SSLMODE")
		os.Unsetenv("DB_MAX_OPEN_CONNS")
		os.Unsetenv("DB_MAX_IDLE_CONNS")
		os.Unsetenv("DB_CONN_MAX_LIFETIME")
		os.Unsetenv("LOG_LEVEL")
		os.Unsetenv("INTERNAL_WORKER_POOL_SIZE")
		os.Unsetenv("INTERNAL_WORKER_SHUTDOWN_TIMEOUT")
		os.Unsetenv("JWT_SECRET")
	}()

	// Test default values
	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if cfg.AppEnv != "development" {
		t.Errorf("Expected APP_ENV to be 'development', got %s", cfg.AppEnv)
	}
	if cfg.Port != "8080" {
		t.Errorf("Expected PORT to be '8080', got %s", cfg.Port)
	}
	if cfg.Database.MaxOpenConns != 10 {
		t.Errorf("Expected DB_MAX_OPEN_CONNS to be 10, got %d", cfg.Database.MaxOpenConns)
	}
	if cfg.Database.ConnMaxLifetime != 5*time.Minute {
		t.Errorf("Expected DB_CONN_MAX_LIFETIME to be 5 minutes, got %s", cfg.Database.ConnMaxLifetime)
	}
	if cfg.Log.Level != "info" {
		t.Errorf("Expected LOG_LEVEL to be 'info', got %s", cfg.Log.Level)
	}

	// Test environment variable override
	os.Setenv("APP_ENV", "production")
	os.Setenv("PORT", "9000")
	os.Setenv("DB_MAX_OPEN_CONNS", "20")
	os.Setenv("DB_CONN_MAX_LIFETIME", "10m")
	os.Setenv("LOG_LEVEL", "debug")
	os.Setenv("INTERNAL_WORKER_POOL_SIZE", "10")
	os.Setenv("INTERNAL_WORKER_SHUTDOWN_TIMEOUT", "30s")
	os.Setenv("JWT_SECRET", "test-secret")


	cfg, err = LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig failed after setting env vars: %v", err)
	}

	if cfg.AppEnv != "production" {
		t.Errorf("Expected APP_ENV to be 'production', got %s", cfg.AppEnv)
	}
	if cfg.Port != "9000" {
		t.Errorf("Expected PORT to be '9000', got %s", cfg.Port)
	}
	if cfg.Database.MaxOpenConns != 20 {
		t.Errorf("Expected DB_MAX_OPEN_CONNS to be 20, got %d", cfg.Database.MaxOpenConns)
	}
	if cfg.Database.ConnMaxLifetime != 10*time.Minute {
		t.Errorf("Expected DB_CONN_MAX_LIFETIME to be 10 minutes, got %s", cfg.Database.ConnMaxLifetime)
	}
	if cfg.Log.Level != "debug" {
		t.Errorf("Expected LOG_LEVEL to be 'debug', got %s", cfg.Log.Level)
	}
	if cfg.InternalWorker.PoolSize != 10 {
		t.Errorf("Expected INTERNAL_WORKER_POOL_SIZE to be 10, got %d", cfg.InternalWorker.PoolSize)
	}
	if cfg.InternalWorker.ShutdownTimeout != 30*time.Second {
		t.Errorf("Expected INTERNAL_WORKER_SHUTDOWN_TIMEOUT to be 30 seconds, got %s", cfg.InternalWorker.ShutdownTimeout)
	}
	if cfg.Auth.JWTSecret != "test-secret" {
		t.Errorf("Expected JWT_SECRET to be 'test-secret', got %s", cfg.Auth.JWTSecret)
	}
}
