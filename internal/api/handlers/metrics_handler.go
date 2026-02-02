package handlers

import (
	"github.com/labstack/echo/v4"
	"net/http"
	"doligo_001/internal/infrastructure/metrics"
)

type MetricsHandler struct {
	metrics *metrics.Metrics
}

func NewMetricsHandler(m *metrics.Metrics) *MetricsHandler {
	return &MetricsHandler{
		metrics: m,
	}
}

// GetMetrics returns the current application metrics.
// @Summary Get internal application metrics
// @Description Retrieves a set of internal, in-memory metrics for monitoring application health and performance.
// @Tags Metrics
// @Accept  json
// @Produce  json
// @Success 200 {object} map[string]interface{}
// @Router /metrics/internal [get]
func (h *MetricsHandler) GetMetrics(c echo.Context) error {
	return c.JSON(http.StatusOK, h.metrics.GetMetrics())
}
