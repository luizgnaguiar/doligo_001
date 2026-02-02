package middleware

import (
	"github.com/labstack/echo/v4"
	"time"
	"doligo_001/internal/infrastructure/metrics"
)

// MetricsMiddleware updates metrics for each request.
func MetricsMiddleware(metrics *metrics.Metrics) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			
			// Increment total requests
			metrics.IncTotalRequests()

			err := next(c)

			// Record endpoint timing
			duration := time.Since(start)
			path := c.Path() // Use the route path as the endpoint identifier
			if path == "" {
				path = c.Request().URL.Path
			}
			metrics.AddEndpointTiming(path, duration)


			// Increment error requests if status is 4xx or 5xx
			status := c.Response().Status
			if status >= 400 && status < 600 {
				metrics.IncErrorRequests()
			}

			return err
		}
	}
}
