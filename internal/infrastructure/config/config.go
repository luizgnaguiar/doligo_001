package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application
type Config struct {
	DBHost         string
	DBPort         string
	DBUser         string
	DBPass         string
	DBName         string
	DBSchema       string
	MaxOpenConns   int
	MaxIdleConns   int
	ConnMaxLifetime time.Duration
}

// Load loads configuration from environment variables
func Load() *Config {
	// In a real scenario, you might not want to ignore this error
	_ = godotenv.Load()

	return &Config{
		DBHost:         getEnv("DB_HOST", "localhost"),
		DBPort:         getEnv("DB_PORT", "5432"),
		DBUser:         getEnv("DB_USER", "postgres"),
		DBPass:         getEnv("DB_PASS", "postgres"),
		DBName:         getEnv("DB_NAME", "doligo"),
		DBSchema:       getEnv("DB_SCHEMA", "public"),
		MaxOpenConns:   getEnvAsInt("DB_MAX_OPEN_CONNS", 10),
		MaxIdleConns:   getEnvAsInt("DB_MAX_IDLE_CONNS", 5),
		ConnMaxLifetime: getEnvAsDuration("DB_CONN_MAX_LIFETIME", 5*time.Minute),
	}
}

// Helper to read an environment variable or return a default value
func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

// Helper to read an environment variable as an integer or return a default value
func getEnvAsInt(key string, fallback int) int {
	strValue := getEnv(key, "")
	if value, err := strconv.Atoi(strValue); err == nil {
		return value
	}
	return fallback
}

// Helper to read an environment variable as a duration or return a default value
func getEnvAsDuration(key string, fallback time.Duration) time.Duration {
	strValue := getEnv(key, "")
	if value, err := time.ParseDuration(strValue); err == nil {
		return value
	}
	return fallback
}
