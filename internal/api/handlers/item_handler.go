// Package handlers contains the HTTP handlers for the API.
package handlers

import (
	"doligo_001/internal/api/dto"
	"doligo_001/internal/usecase/item"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"net/http"
)

// ItemHandler handles HTTP requests for items.
type ItemHandler struct {
	usecase item.Usecase
}

// NewItemHandler creates a new ItemHandler.
func NewItemHandler(uc item.Usecase) *ItemHandler {
	return &ItemHandler{usecase: uc}
}

// RegisterRoutes registers the item routes to an Echo group.
func (h *ItemHandler) RegisterRoutes(g *echo.Group) {
	g.POST("", h.Create)
	g.GET("/:id", h.GetByID)
	g.PUT("/:id", h.Update)
	g.DELETE("/:id", h.Delete)
	g.GET("", h.List)
}

// Create handles the creation of a new item.
func (h *ItemHandler) Create(c echo.Context) error {
	req := new(dto.CreateItemRequest)
	if err := c.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := c.Validate(req); err != nil {
		return err
	}

	i, err := h.usecase.Create(c.Request().Context(), req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, dto.NewItemResponse(i))
}

// GetByID retrieves an item by its ID.
func (h *ItemHandler) GetByID(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid ID format")
	}

	i, err := h.usecase.GetByID(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, dto.NewItemResponse(i))
}

// Update handles the update of an existing item.
func (h *ItemHandler) Update(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid ID format")
	}

	req := new(dto.UpdateItemRequest)
	if err := c.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := c.Validate(req); err != nil {
		return err
	}

	i, err := h.usecase.Update(c.Request().Context(), id, req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, dto.NewItemResponse(i))
}

// Delete handles the deletion of an item.
func (h *ItemHandler) Delete(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid ID format")
	}

	if err := h.usecase.Delete(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.NoContent(http.StatusNoContent)
}

// List handles listing all items.
func (h *ItemHandler) List(c echo.Context) error {
	items, err := h.usecase.List(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	res := make([]*dto.ItemResponse, len(items))
	for i, item := range items {
		res[i] = dto.NewItemResponse(item)
	}

	return c.JSON(http.StatusOK, res)
}
