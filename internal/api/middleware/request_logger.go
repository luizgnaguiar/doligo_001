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

const headerCorrelationID = "X-Correlation-ID"

// RequestLogger is a middleware that provides structured logging for each HTTP request.
// It creates or reuses a `correlation_id` for tracing.
// It logs the start and end of a request, including details like the correlation ID,
// method, URI, status code, and latency. It also embeds the correlation ID and
// a request-scoped logger into the context for downstream use.
func RequestLogger(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Start timer
		start := time.Now()

		// Get or generate correlation ID
		correlationID := c.Request().Header.Get(headerCorrelationID)
		if correlationID == "" {
			correlationID = uuid.New().String()
		}

		// Set correlation ID on the response header
		c.Response().Header().Set(headerCorrelationID, correlationID)

		// Create a logger with the correlation ID and embed it in the context
		requestLogger := slog.With("correlation_id", correlationID)
		ctx := logger.ToContext(c.Request().Context(), requestLogger)

		// Add correlation ID to context for propagation
		ctxWithCorrelation := ToContext(ctx, correlationID)
		c.SetRequest(c.Request().WithContext(ctxWithCorrelation))

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

		requestLogger.Log(ctxWithCorrelation, logLevel, "Request completed",
			"method", c.Request().Method,
			"uri", c.Request().RequestURI,
			"status", status,
			"latency", latency.String(),
		)

		return err
	}
}
