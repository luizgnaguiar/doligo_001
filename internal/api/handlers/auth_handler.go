// Package handlers contains the HTTP handlers for the API. Handlers are responsible
// for parsing incoming requests, calling the appropriate use cases with validated
// data, and formatting the responses. They act as the entry point to the
// application's business logic from the web layer.
package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"golang.org/x/net/context"

	"doligo_001/internal/api/dto"
)

// AuthUsecase defines the contract for authentication-related business logic.
// This interface will be implemented by a use case in the usecase layer.
type AuthUsecase interface {
	Login(ctx context.Context, email, password string) (string, error)
}

// AuthHandler handles HTTP requests related to authentication.
type AuthHandler struct {
	usecase AuthUsecase
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(usecase AuthUsecase) *AuthHandler {
	return &AuthHandler{usecase: usecase}
}

// Login handles the user login request.
// It expects a JSON body with email and password, validates it,
// calls the login use case, and returns a JWT token upon success.
func (h *AuthHandler) Login(c echo.Context) error {
	var req dto.LoginRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request format")
	}

	// In a real application, you would use a validator here.
	// For now, we'll keep it simple.
	if req.Email == "" || req.Password == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Email and password are required")
	}

	token, err := h.usecase.Login(c.Request().Context(), req.Email, req.Password)
	if err != nil {
		// In a real app, you'd check for specific error types,
		// e.g., to return 401 for bad credentials.
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to login")
	}

	return c.JSON(http.StatusOK, dto.LoginResponse{Token: token})
}

// RegisterRoutes registers the authentication-related routes to the Echo router.
func (h *AuthHandler) RegisterRoutes(e *echo.Echo) {
	e.POST("/login", h.Login)
}
