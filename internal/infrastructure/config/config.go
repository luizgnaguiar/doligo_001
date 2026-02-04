package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// Config holds the application's configuration
type Config struct {
	AppEnv         string `mapstructure:"APP_ENV"`
	Port           string `mapstructure:"PORT"`
	Database       DatabaseConfig `mapstructure:",squash"`
	Log            LogConfig      `mapstructure:",squash"`
	InternalWorker InternalWorkerConfig `mapstructure:",squash"`
	JWT            AuthConfig     `mapstructure:",squash"`
	RateLimit      RateLimitConfig `mapstructure:",squash"`
	Security       SecurityConfig  `mapstructure:",squash"`
}

// DatabaseConfig holds database-related configuration
type DatabaseConfig struct {
	Type        string `mapstructure:"DB_TYPE"`
	Host        string `mapstructure:"DB_HOST"`
	Port        int    `mapstructure:"DB_PORT"`
	User        string `mapstructure:"DB_USER"`
	Password    string `mapstructure:"DB_PASSWORD"`
	Name        string `mapstructure:"DB_NAME"`
	SSLMode     string `mapstructure:"DB_SSLMODE"`
	MaxOpenConns    int `mapstructure:"DB_MAX_OPEN_CONNS"`
	MaxIdleConns    int `mapstructure:"DB_MAX_IDLE_CONNS"`
	ConnMaxLifetime time.Duration `mapstructure:"DB_CONN_MAX_LIFETIME"`
}

// LogConfig holds logging-related configuration
type LogConfig struct {
	Level string `mapstructure:"LOG_LEVEL"`
}

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	Enabled           bool `mapstructure:"RATE_LIMIT_ENABLED"`
	RequestsPerSecond int  `mapstructure:"RATE_LIMIT_REQUESTS_PER_SECOND"`
	Burst             int  `mapstructure:"RATE_LIMIT_BURST"`
}

// SecurityConfig holds security-related configuration
type SecurityConfig struct {
	CORSAllowedOrigins []string `mapstructure:"CORS_ALLOWED_ORIGINS"`
	CORSAllowCredentials bool    `mapstructure:"CORS_ALLOW_CREDENTIALS"`
	SecurityHeadersEnabled bool  `mapstructure:"SECURITY_HEADERS_ENABLED"`
}

// InternalWorkerConfig holds configuration for the internal task runner
type InternalWorkerConfig struct {
	PoolSize int `mapstructure:"INTERNAL_WORKER_POOL_SIZE"`
	ShutdownTimeout time.Duration `mapstructure:"INTERNAL_WORKER_SHUTDOWN_TIMEOUT"`
}

// AuthConfig holds authentication related configuration
type AuthConfig struct {
	JWTSecret string `mapstructure:"JWT_SECRET"`
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	viper.SetDefault("APP_ENV", "development")
	viper.SetDefault("PORT", "8080")
	viper.SetDefault("DB_TYPE", "postgres")
	viper.SetDefault("DB_HOST", "localhost")
	viper.SetDefault("DB_PORT", 5432)
	viper.SetDefault("DB_USER", "user")
	viper.SetDefault("DB_PASSWORD", "password")
	viper.SetDefault("DB_NAME", "dolibarr")
	viper.SetDefault("DB_SSLMODE", "disable")
	viper.SetDefault("DB_MAX_OPEN_CONNS", 10)
	viper.SetDefault("DB_MAX_IDLE_CONNS", 5)
	viper.SetDefault("DB_CONN_MAX_LIFETIME", 5 * time.Minute)
	viper.SetDefault("LOG_LEVEL", "info")
	viper.SetDefault("INTERNAL_WORKER_POOL_SIZE", 5)
	viper.SetDefault("INTERNAL_WORKER_SHUTDOWN_TIMEOUT", 15 * time.Second)
	viper.SetDefault("JWT_SECRET", "super-secret-jwt-key")
	viper.SetDefault("RATE_LIMIT_ENABLED", false)
	viper.SetDefault("RATE_LIMIT_REQUESTS_PER_SECOND", 10)
	viper.SetDefault("RATE_LIMIT_BURST", 20)
	viper.SetDefault("CORS_ALLOWED_ORIGINS", []string{"*"})
	viper.SetDefault("CORS_ALLOW_CREDENTIALS", true)
	viper.SetDefault("SECURITY_HEADERS_ENABLED", true)


	viper.AutomaticEnv() // Read from environment variables

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}
