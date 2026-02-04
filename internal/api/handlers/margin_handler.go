package handlers

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"doligo_001/internal/api/dto"
	domainMargin "doligo_001/internal/domain/margin" // Alias for domain margin
	marginUseCase "doligo_001/internal/usecase/margin" // Alias for usecase margin
)

// MarginHandler handles HTTP requests related to margin reports.
type MarginHandler struct {
	marginUsecase marginUseCase.MarginUsecase
}

// NewMarginHandler creates a new MarginHandler.
func NewMarginHandler(mu marginUseCase.MarginUsecase) *MarginHandler {
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
// @Success 200 {object} domainMargin.MarginReport
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /margin/products/{productID} [get]
func (h *MarginHandler) GetProductMarginReport(c echo.Context) error {
	var req dto.ProductMarginReportRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Message: "Invalid request parameters", Details: err.Error()})
	}

	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Message: "Validation failed", Details: err.Error()})
	}

	productID, _ := uuid.Parse(req.ProductID) // Already validated
	startDate, _ := time.Parse("2006-01-02", req.StartDate)
	endDate, _ := time.Parse("2006-01-02", req.EndDate)

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
// @Success 200 {array} domainMargin.MarginReport
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /margin [get]
func (h *MarginHandler) ListOverallMarginReports(c echo.Context) error {
	var req dto.OverallMarginReportRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Message: "Invalid request parameters", Details: err.Error()})
	}

	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Message: "Validation failed", Details: err.Error()})
	}

	startDate, _ := time.Parse("2006-01-02", req.StartDate)
	endDate, _ := time.Parse("2006-01-02", req.EndDate)

	reports, err := h.marginUsecase.ListOverallMarginReports(c.Request().Context(), startDate, endDate)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Message: "Failed to retrieve overall margin reports", Details: err.Error()})
	}
	if reports == nil {
		return c.JSON(http.StatusOK, []domainMargin.MarginReport{}) // Return empty array instead of null
	}

	return c.JSON(http.StatusOK, reports)
}

// RegisterRoutes registers margin routes to an Echo group.
func (h *MarginHandler) RegisterRoutes(g *echo.Group) {
	g.GET("/products/:productID", h.GetProductMarginReport)
	g.GET("", h.ListOverallMarginReports)
}
