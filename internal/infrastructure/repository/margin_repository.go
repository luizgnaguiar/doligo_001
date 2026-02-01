package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"doligo_001/internal/domain/margin"
)

// GormMarginRepository implements the margin.Repository interface using GORM and raw SQL.
type GormMarginRepository struct {
	db *gorm.DB
}

// NewGormMarginRepository creates a new GormMarginRepository.
func NewGormMarginRepository(db *gorm.DB) *GormMarginRepository {
	return &GormMarginRepository{db: db}
}

// GetMarginReport retrieves the margin report for a single product within a given period.
// This implementation uses raw SQL to aggregate data from production records and (assumed) invoices.
func (r *GormMarginRepository) GetMarginReport(ctx context.Context, productID uuid.UUID, startDate, endDate time.Time) (*margin.MarginReport, error) {
	// For simplicity, this example assumes that "service cost" is implicitly part of "actual_production_cost"
	// or requires specific logic to be factored in. For now, we'll combine all costs.
	query := `
		SELECT
			pr.produced_product_id AS product_id,
			i.name AS product_name,
			SUM(pr.actual_production_cost) AS total_input_cost,
			COALESCE(SUM(inv_items.unit_price * inv_items.quantity), 0) AS total_selling_price,
			COALESCE(SUM(inv_items.tax_amount), 0) AS total_taxes
		FROM production_records pr
		LEFT JOIN items i ON pr.produced_product_id = i.id
		LEFT JOIN invoice_line_items inv_items ON pr.produced_product_id = inv_items.product_id
		LEFT JOIN invoices inv ON inv_items.invoice_id = inv.id
		WHERE pr.produced_product_id = ?
		AND pr.produced_at >= ? AND pr.produced_at <= ?
		AND inv.invoice_date >= ? AND inv.invoice_date <= ?
		GROUP BY pr.produced_product_id, i.name
	`

	var result struct {
		ProductID         uuid.UUID `gorm:"column:product_id"`
		ProductName       string    `gorm:"column:product_name"`
		TotalInputCost    float64   `gorm:"column:total_input_cost"`
		TotalSellingPrice float64   `gorm:"column:total_selling_price"`
		TotalTaxes        float64   `gorm:"column:total_taxes"`
	}

	// Execute the raw SQL query
	err := r.db.WithContext(ctx).Raw(query, productID, startDate, endDate, startDate, endDate).Scan(&result).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // No records found for this product in the given period
		}
		return nil, fmt.Errorf("failed to get margin report for product %s: %w", productID, err)
	}

	report := &margin.MarginReport{
		PeriodStart:       startDate,
		PeriodEnd:         endDate,
		ProductID:         result.ProductID,
		ProductName:       result.ProductName,
		TotalInputCost:    result.TotalInputCost,
		TotalServiceCost:  0, // Placeholder, as not explicitly tracked in production_records
		TotalTaxes:        result.TotalTaxes,
		TotalSellingPrice: result.TotalSellingPrice,
	}

	// Calculate Gross Margin
	report.GrossMargin = report.TotalSellingPrice - (report.TotalInputCost + report.TotalServiceCost + report.TotalTaxes)
	if report.TotalSellingPrice > 0 {
		report.GrossMarginPercentage = (report.GrossMargin / report.TotalSellingPrice) * 100
	}

	return report, nil
}

// ListMarginReports retrieves margin reports for all products within a given period.
// This implementation uses raw SQL for aggregation.
func (r *GormMarginRepository) ListMarginReports(ctx context.Context, startDate, endDate time.Time) ([]*margin.MarginReport, error) {
	query := `
		SELECT
			pr.produced_product_id AS product_id,
			i.name AS product_name,
			SUM(pr.actual_production_cost) AS total_input_cost,
			COALESCE(SUM(inv_items.unit_price * inv_items.quantity), 0) AS total_selling_price,
			COALESCE(SUM(inv_items.tax_amount), 0) AS total_taxes
		FROM production_records pr
		LEFT JOIN items i ON pr.produced_product_id = i.id
		LEFT JOIN invoice_line_items inv_items ON pr.produced_product_id = inv_items.product_id
		LEFT JOIN invoices inv ON inv_items.invoice_id = inv.id
		WHERE pr.produced_at >= ? AND pr.produced_at <= ?
		AND inv.invoice_date >= ? AND inv.invoice_date <= ?
		GROUP BY pr.produced_product_id, i.name
		ORDER BY i.name
	`

	var results []struct {
		ProductID         uuid.UUID `gorm:"column:product_id"`
		ProductName       string    `gorm:"column:product_name"`
		TotalInputCost    float64   `gorm:"column:total_input_cost"`
		TotalSellingPrice float64   `gorm:"column:total_selling_price"`
		TotalTaxes        float64   `gorm:"column:total_taxes"`
	}

	err := r.db.WithContext(ctx).Raw(query, startDate, endDate, startDate, endDate).Scan(&results).Error
	if err != nil {
		return nil, fmt.Errorf("failed to list margin reports: %w", err)
	}

	var reports []*margin.MarginReport
	for _, result := range results {
		report := &margin.MarginReport{
			PeriodStart:       startDate,
			PeriodEnd:         endDate,
			ProductID:         result.ProductID,
			ProductName:       result.ProductName,
			TotalInputCost:    result.TotalInputCost,
			TotalServiceCost:  0, // Placeholder
			TotalTaxes:        result.TotalTaxes,
			TotalSellingPrice: result.TotalSellingPrice,
		}
		report.GrossMargin = report.TotalSellingPrice - (report.TotalInputCost + report.TotalServiceCost + report.TotalTaxes)
		if report.TotalSellingPrice > 0 {
			report.GrossMarginPercentage = (report.GrossMargin / report.TotalSellingPrice) * 100
		}
		reports = append(reports, report)
	}

	return reports, nil
}