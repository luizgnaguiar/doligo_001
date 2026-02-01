package margin

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// MarginReport represents a single entry in the margin dashboard.
type MarginReport struct {
	PeriodStart          time.Time `json:"period_start"`
	PeriodEnd            time.Time `json:"period_end"`
	ProductID            uuid.UUID `json:"product_id"`
	ProductName          string    `json:"product_name"`
	TotalSellingPrice    float64   `json:"total_selling_price"`
	TotalInputCost       float64   `json:"total_input_cost"`
	TotalServiceCost     float64   `json:"total_service_cost"` // Assuming service cost is part of input cost or separate production overhead
	TotalTaxes           float64   `json:"total_taxes"`
	GrossMargin          float64   `json:"gross_margin"` // TotalSellingPrice - TotalInputCost - TotalServiceCost - TotalTaxes
	GrossMarginPercentage float64   `json:"gross_margin_percentage"`
}

// Repository defines the interface for retrieving margin-related data.
type Repository interface {
	GetMarginReport(ctx context.Context, productID uuid.UUID, startDate, endDate time.Time) (*MarginReport, error)
	ListMarginReports(ctx context.Context, startDate, endDate time.Time) ([]*MarginReport, error)
}
