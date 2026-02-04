package middleware

import (
	"github.com/labstack/echo/v4"
)

// SecurityHeadersMiddleware adds security-related headers to the response
func SecurityHeadersMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Protection against Clickjacking
			c.Response().Header().Set("X-Frame-Options", "DENY")
			
			// Protection against MIME sniffing
			c.Response().Header().Set("X-Content-Type-Options", "nosniff")
			
			// Protection against XSS
			c.Response().Header().Set("X-XSS-Protection", "1; mode=block")
			
			// Strict Transport Security (HSTS)
			c.Response().Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
			
			// Content Security Policy (CSP)
			c.Response().Header().Set("Content-Security-Policy", "default-src 'self'")

			return next(c)
		}
	}
}
