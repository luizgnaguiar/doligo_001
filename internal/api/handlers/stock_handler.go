// Package handlers contains the HTTP handlers for the API.
package handlers

import (
	"doligo_001/internal/api/dto"
	"doligo_001/internal/domain/stock"
	stock_usecase "doligo_001/internal/usecase/stock"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"net/http"
)

// StockHandler handles HTTP requests for stock management.
type StockHandler struct {
	usecase stock_usecase.UseCase
}

// NewStockHandler creates a new StockHandler.
func NewStockHandler(uc stock_usecase.UseCase) *StockHandler {
	return &StockHandler{usecase: uc}
}

// RegisterRoutes registers the stock-related routes to an Echo group.
func (h *StockHandler) RegisterRoutes(g *echo.Group) {
	// Warehouse routes
	g.POST("/warehouses", h.CreateWarehouse)
	g.GET("/warehouses", h.ListWarehouses)
	g.GET("/warehouses/:id", h.GetWarehouseByID)
	
	// Bin routes
	g.POST("/bins", h.CreateBin)
	g.GET("/warehouses/:id/bins", h.ListBinsByWarehouse)

	// Stock movement routes
	g.POST("/stock/movements", h.CreateStockMovement)
}

func (h *StockHandler) CreateWarehouse(c echo.Context) error {
	req := new(dto.CreateWarehouseRequest)
	if err := c.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	
	if err := c.Validate(req); err != nil {
		return err
	}

	w, err := h.usecase.CreateWarehouse(c.Request().Context(), req.Name)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, dto.NewWarehouseResponse(w))
}

func (h *StockHandler) ListWarehouses(c echo.Context) error {
	warehouses, err := h.usecase.ListWarehouses(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	res := make([]*dto.WarehouseResponse, len(warehouses))
	for i, w := range warehouses {
		res[i] = dto.NewWarehouseResponse(w)
	}
	return c.JSON(http.StatusOK, res)
}

func (h *StockHandler) GetWarehouseByID(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid ID format")
	}
	w, err := h.usecase.GetWarehouseByID(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, dto.NewWarehouseResponse(w))
}

func (h *StockHandler) CreateBin(c echo.Context) error {
	req := new(dto.CreateBinRequest)
	if err := c.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	
	if err := c.Validate(req); err != nil {
		return err
	}
	warehouseID, err := uuid.Parse(req.WarehouseID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid Warehouse ID format")
	}

	b, err := h.usecase.CreateBin(c.Request().Context(), req.Name, warehouseID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, dto.NewBinResponse(b))
}

func (h *StockHandler) ListBinsByWarehouse(c echo.Context) error {
	warehouseID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid Warehouse ID format")
	}
	bins, err := h.usecase.ListBinsByWarehouse(c.Request().Context(), warehouseID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	res := make([]*dto.BinResponse, len(bins))
	for i, b := range bins {
		res[i] = dto.NewBinResponse(b)
	}
	return c.JSON(http.StatusOK, res)
}

func (h *StockHandler) CreateStockMovement(c echo.Context) error {
	req := new(dto.CreateStockMovementRequest)
	if err := c.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	
	if err := c.Validate(req); err != nil {
		return err
	}

	itemID, err := uuid.Parse(req.ItemID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid Item ID format")
	}
	warehouseID, err := uuid.Parse(req.WarehouseID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid Warehouse ID format")
	}
	binID, err := uuid.Parse(req.BinID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid Bin ID format")
	}

	movement, err := h.usecase.CreateStockMovement(
		c.Request().Context(),
		itemID,
		warehouseID,
		binID,
		stock.MovementType(req.Type),
		req.Quantity,
		req.Reason,
	)
	if err != nil {
		if err == stock_usecase.ErrInsufficientStock || err == stock_usecase.ErrBinRequired {
			return echo.NewHTTPError(http.StatusConflict, err.Error())
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, dto.NewStockMovementResponse(movement))
}
