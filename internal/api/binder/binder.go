package binder

import (
	"doligo_001/internal/api/sanitizer"
	"github.com/labstack/echo/v4"
)

// CustomBinder wraps the default binder and adds sanitization.
type CustomBinder struct {
	DefaultBinder *echo.DefaultBinder
}

// Bind binds the request body to the provided interface and then sanitizes it.
func (cb *CustomBinder) Bind(i interface{}, c echo.Context) error {
	if err := cb.DefaultBinder.Bind(i, c); err != nil {
		return err
	}
	if s, ok := i.(sanitizer.Sanitizable); ok {
		s.Sanitize()
	}
	return nil
}
