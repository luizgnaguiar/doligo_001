package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/stretchr/testify/assert"
)

func TestSecurityHeadersMiddleware(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := SecurityHeadersMiddleware()(func(c echo.Context) error {
		return c.String(http.StatusOK, "test")
	})

	if assert.NoError(t, h(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "DENY", rec.Header().Get("X-Frame-Options"))
		assert.Equal(t, "nosniff", rec.Header().Get("X-Content-Type-Options"))
		assert.Equal(t, "1; mode=block", rec.Header().Get("X-XSS-Protection"))
		assert.Equal(t, "max-age=31536000; includeSubDomains", rec.Header().Get("Strict-Transport-Security"))
		assert.Equal(t, "default-src 'self'", rec.Header().Get("Content-Security-Policy"))
	}
}

func TestCORSMiddlewareConfig(t *testing.T) {
	e := echo.New()
	
	// Setup CORS middleware
	allowedOrigins := []string{"https://allowed.com", "https://another-allowed.com"}
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: allowedOrigins,
		AllowMethods: []string{http.MethodGet, http.MethodPost},
	}))

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "test")
	})

	t.Run("Allowed Origin", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Origin", "https://allowed.com")
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "https://allowed.com", rec.Header().Get("Access-Control-Allow-Origin"))
	})

	t.Run("Disallowed Origin", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Origin", "https://evil.com")
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		// With Echo CORS, disallowed origin usually results in no ACAO header, but request might still be processed if not restricted strict enough or default behavior. 
		// Actually, Echo's CORS middleware, if origin is not allowed, does NOT set Access-Control-Allow-Origin.
		// The status code might still be 200 depending on implementation, but the browser blocks it.
		// Let's check if header is absent.
		assert.Equal(t, "", rec.Header().Get("Access-Control-Allow-Origin"))
	})
	
	t.Run("Allowed Method", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodOptions, "/", nil)
		req.Header.Set("Origin", "https://allowed.com")
		req.Header.Set("Access-Control-Request-Method", "POST")
		rec := httptest.NewRecorder()
		
		e.ServeHTTP(rec, req)
		
		assert.Equal(t, http.StatusNoContent, rec.Code)
		assert.Equal(t, "GET,POST", rec.Header().Get("Access-Control-Allow-Methods"))
	})
}
