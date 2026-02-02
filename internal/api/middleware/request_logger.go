// Package middleware provides Echo middleware functions for handling common
// cross-cutting concerns, such as authentication, logging, and error handling.
package middleware

import (
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"doligo_001/internal/infrastructure/logger"
)

// RequestLogger is a middleware that provides structured logging for each HTTP request.
// It logs the start and end of a request, including details like the request ID,
// method, URI, status code, and latency. It also embeds a request-scoped logger
// into the context for downstream use.
func RequestLogger(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Start timer
		start := time.Now()

		// Generate a unique request ID
		reqID := uuid.New().String()
		c.Response().Header().Set(echo.HeaderXRequestID, reqID)

		// Create a logger with the request ID and embed it in the context
		requestLogger := slog.With("request_id", reqID)
		ctx := logger.ToContext(c.Request().Context(), requestLogger)
		c.SetRequest(c.Request().WithContext(ctx))

		requestLogger.Info("Request started",
			"method", c.Request().Method,
			"uri", c.Request().RequestURI,
		)

		// Call the next handler
		err := next(c)

		// Stop timer
		latency := time.Since(start)
		status := c.Response().Status

		logLevel := slog.LevelInfo
		if err != nil {
			// If an HTTP error was returned by a handler, use its status code
			if he, ok := err.(*echo.HTTPError); ok {
				status = he.Code
			}
			logLevel = slog.LevelError
		}

		if status >= 500 {
			logLevel = slog.LevelError
		} else if status >= 400 {
			logLevel = slog.LevelWarn
		}

		requestLogger.Log(ctx, logLevel, "Request completed",
			"method", c.Request().Method,
			"uri", c.Request().RequestURI,
			"status", status,
			"latency", latency.String(),
		)

		return err
	}
}
