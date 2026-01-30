package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"doligo_001/internal/api/auth"
)

type contextKey string

const (
	UserIDContextKey      contextKey = "userID"
	PermissionsContextKey contextKey = "permissions"
)

func AuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		authHeader := c.Request().Header.Get("Authorization")
		if authHeader == "" {
			return c.JSON(http.StatusUnauthorized, "missing authorization header")
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			return c.JSON(http.StatusUnauthorized, "invalid authorization header")
		}

		claims, err := auth.ValidateToken(parts[1])
		if err != nil {
			return c.JSON(http.StatusUnauthorized, "invalid token")
		}

		ctx := c.Request().Context()
		ctx = context.WithValue(ctx, UserIDContextKey, claims.UserID)
		ctx = context.WithValue(ctx, PermissionsContextKey, claims.Permissions)
		c.SetRequest(c.Request().WithContext(ctx))

		return next(c)
	}
}
