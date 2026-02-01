package handlers

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"doligo_001/internal/api/dto"
	"doligo_001/internal/usecase/margin"
)

// MarginUsecase defines the interface for margin-related business logic.
type MarginUsecase interface {
	GetProductMarginReport(ctx echo.Context, productID uuid.UUID, startDate, endDate time.Time) (*margin.MarginReport, error)
	ListOverallMarginReports(ctx echo.Context, startDate, endDate time.Time) ([]*margin.MarginReport, error)
}

// MarginHandler handles HTTP requests related to margin reports.
type MarginHandler struct {
	marginUsecase MarginUsecase
}

// NewMarginHandler creates a new MarginHandler.
func NewMarginHandler(mu MarginUsecase) *MarginHandler {
	return &MarginHandler{marginUsecase: mu}
}

// GetProductMarginReport godoc
// @Summary Get margin report for a single product
// @Description Retrieves a detailed margin report for a specific product within a given date range.
// @Tags Margin
// @Accept json
// @Produce json
// @Param productID path string true "Product ID" Format(uuid)
// @Param startDate query string true "Start date for the report (YYYY-MM-DD)"
// @Param endDate query string true "End date for the report (YYYY-MM-DD)"
// @Success 200 {object} margin.MarginReport
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /margin/products/{productID} [get]
func (h *MarginHandler) GetProductMarginReport(c echo.Context) error {
	productIDStr := c.Param("productID")
	productID, err := uuid.Parse(productIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Message: "Invalid product ID format"})
	}

	startDateStr := c.QueryParam("startDate")
	endDateStr := c.QueryParam("endDate")

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Message: "Invalid start date format. Use YYYY-MM-DD"})
	}
	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Message: "Invalid end date format. Use YYYY-MM-DD"})
	}

	report, err := h.marginUsecase.GetProductMarginReport(c.Request().Context(), productID, startDate, endDate)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Message: "Failed to retrieve product margin report", Details: err.Error()})
	}
	if report == nil {
		return c.JSON(http.StatusNotFound, dto.ErrorResponse{Message: "Margin report not found for the given product and period"})
	}

	return c.JSON(http.StatusOK, report)
}

// ListOverallMarginReports godoc
// @Summary Get overall margin reports
// @Description Retrieves aggregated margin reports for all products within a given date range.
// @Tags Margin
// @Accept json
// @Produce json
// @Param startDate query string true "Start date for the report (YYYY-MM-DD)"
// @Param endDate query string true "End date for the report (YYYY-MM-DD)"
// @Success 200 {array} margin.MarginReport
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /margin [get]
func (h *MarginHandler) ListOverallMarginReports(c echo.Context) error {
	startDateStr := c.QueryParam("startDate")
	endDateStr := c.QueryParam("endDate")

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Message: "Invalid start date format. Use YYYY-MM-DD"})
	}
	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Message: "Invalid end date format. Use YYYY-MM-DD"})
	}

	reports, err := h.marginUsecase.ListOverallMarginReports(c.Request().Context(), startDate, endDate)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Message: "Failed to retrieve overall margin reports", Details: err.Error()})
	}
	if reports == nil {
		return c.JSON(http.StatusOK, []margin.MarginReport{}) // Return empty array instead of null
	}

	return c.JSON(http.StatusOK, reports)
}

// RegisterRoutes registers margin routes to an Echo group.
func (h *MarginHandler) RegisterRoutes(g *echo.Group) {
	g.GET("/products/:productID", h.GetProductMarginReport)
	g.GET("", h.ListOverallMarginReports)
}
