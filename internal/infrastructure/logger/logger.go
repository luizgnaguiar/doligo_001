// Package logger provides standardized logging for the application.
// It initializes and configures a global slog.Logger instance that provides
// structured, level-based logging. The logger is configured to output
// JSON for production environments and a more human-readable format for
// development.
package logger

import (
	"context"
	"log/slog"
	"os"
)

type contextKey string

const (
	loggerKey contextKey = "logger"
)

// InitLogger initializes a new slog.Logger and sets it as the default.
func InitLogger(logLevel string) {
	var level slog.Level
	if err := level.UnmarshalText([]byte(logLevel)); err != nil {
		slog.Error("Invalid log level provided, defaulting to INFO", "provided_level", logLevel, "error", err)
		level = slog.LevelInfo
	}

	// For development, use a more readable handler. For production, use JSON.
	var handler slog.Handler
	// NOTE: This is a simple check. A more robust solution might use an explicit `APP_ENV` variable.
	if level <= slog.LevelDebug {
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	} else {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	}

	logger := slog.New(handler)
	slog.SetDefault(logger)
	slog.Info("Logger initialized", "level", level.String())
}

// FromContext retrieves a logger from the context. If no logger is found,
// it returns the default logger. This is useful for accessing a request-scoped
// logger that may have additional attributes (like a request ID).
func FromContext(ctx context.Context) *slog.Logger {
	if logger, ok := ctx.Value(loggerKey).(*slog.Logger); ok {
		return logger
	}
	return slog.Default()
}

// ToContext embeds a logger into the context. This allows for passing a
// logger with specific attributes (like a request ID) through the call stack.
func ToContext(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}