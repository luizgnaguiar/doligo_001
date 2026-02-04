// Package handlers contains the HTTP handlers for the API.
package handlers

import (
	"doligo_001/internal/api/dto"
	"doligo_001/internal/api/sanitizer"
	"doligo_001/internal/usecase/thirdparty"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"net/http"
)

// ThirdPartyHandler handles HTTP requests for third parties.
type ThirdPartyHandler struct {
	usecase thirdparty.Usecase
}

// NewThirdPartyHandler creates a new ThirdPartyHandler.
func NewThirdPartyHandler(uc thirdparty.Usecase) *ThirdPartyHandler {
	return &ThirdPartyHandler{usecase: uc}
}

// RegisterRoutes registers the third party routes to an Echo group.
func (h *ThirdPartyHandler) RegisterRoutes(g *echo.Group) {
	g.POST("", h.Create)
	g.GET("/:id", h.GetByID)
	g.PUT("/:id", h.Update)
	g.DELETE("/:id", h.Delete)
	g.GET("", h.List)
}

// Create handles the creation of a new third party.
func (h *ThirdPartyHandler) Create(c echo.Context) error {
	req := new(dto.CreateThirdPartyRequest)
	if err := c.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	req.Name = sanitizer.SanitizeString(req.Name)
	req.Email = sanitizer.SanitizeString(req.Email)

	if err := c.Validate(req); err != nil {
		return err
	}

	tp, err := h.usecase.Create(c.Request().Context(), req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, dto.NewThirdPartyResponse(tp))
}

// GetByID retrieves a third party by its ID.
func (h *ThirdPartyHandler) GetByID(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid ID format")
	}

	tp, err := h.usecase.GetByID(c.Request().Context(), id)
	if err != nil {
		// Consider checking for gorm.ErrRecordNotFound to return 404
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, dto.NewThirdPartyResponse(tp))
}

// Update handles the update of an existing third party.
func (h *ThirdPartyHandler) Update(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid ID format")
	}

	req := new(dto.UpdateThirdPartyRequest)
	if err := c.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	req.Name = sanitizer.SanitizeString(req.Name)
	req.Email = sanitizer.SanitizeString(req.Email)

	if err := c.Validate(req); err != nil {
		return err
	}

	tp, err := h.usecase.Update(c.Request().Context(), id, req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, dto.NewThirdPartyResponse(tp))
}

// Delete handles the deletion of a third party.
func (h *ThirdPartyHandler) Delete(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid ID format")
	}

	if err := h.usecase.Delete(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.NoContent(http.StatusNoContent)
}

// List handles listing all third parties.
func (h *ThirdPartyHandler) List(c echo.Context) error {
	tps, err := h.usecase.List(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	res := make([]*dto.ThirdPartyResponse, len(tps))
	for i, tp := range tps {
		res[i] = dto.NewThirdPartyResponse(tp)
	}

	return c.JSON(http.StatusOK, res)
}
