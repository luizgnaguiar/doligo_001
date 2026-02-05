// Package middleware provides Echo middleware functions for handling common
// cross-cutting concerns, such as authentication, logging, and error handling.
// Middlewares process requests before they reach the handlers, allowing for
// centralized and reusable logic.
package middleware

import (
	"doligo_001/internal/domain"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// CustomContextKey is a custom type to avoid key collisions in context.
type CustomContextKey string

const (
	// PermissionsKey is the key for storing user permissions in the context.
	PermissionsKey CustomContextKey = "permissions"
)

// Claims represents the JWT claims.
type Claims struct {
	UserID      uuid.UUID `json:"user_id"`
	Permissions []string  `json:"permissions"`
	jwt.RegisteredClaims
}

// JWTConfig holds the configuration for the JWT middleware.
type JWTConfig struct {
	Secret []byte
}

// JWT middleware validates the JWT token and extracts user information.
// It injects the userID and permissions into the request context for downstream
// handlers to use.
func (config *JWTConfig) JWT(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		authHeader := c.Request().Header.Get("Authorization")
		if authHeader == "" {
			return echo.NewHTTPError(http.StatusUnauthorized, "Missing Authorization header")
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return echo.NewHTTPError(http.StatusUnauthorized, "Invalid Authorization header format")
		}

		tokenString := parts[1]
		claims := &Claims{}

		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return config.Secret, nil
		})

		if err != nil || !token.Valid {
			return echo.NewHTTPError(http.StatusUnauthorized, "Invalid or expired token")
		}

		// Inject user info into context using the domain's function
		ctx := domain.ContextWithUserID(c.Request().Context(), claims.UserID)
		ctx = domain.ContextWithPermissions(ctx, claims.Permissions)
		c.SetRequest(c.Request().WithContext(ctx))

		return next(c)
	}
}
