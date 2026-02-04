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
// This implementation now uses raw SQL to aggregate data from the real `invoice_lines` table.
func (r *GormMarginRepository) GetMarginReport(ctx context.Context, productID uuid.UUID, startDate, endDate time.Time) (*margin.MarginReport, error) {
	// Technical Debt: This query only considers sales data from invoices.
	// It does not yet factor in production costs, service costs, or other operational expenses.
	// TotalTaxes are also not yet implemented in the invoice domain.
	query := `
		SELECT
			il.item_id AS product_id,
			i.name AS product_name,
			SUM(il.total_cost) AS total_input_cost,
			SUM(il.total_amount) AS total_selling_price,
			SUM(il.tax_amount * il.quantity) AS total_taxes
		FROM invoice_lines il
		JOIN items i ON il.item_id = i.id
		JOIN invoices inv ON il.invoice_id = inv.id
		WHERE il.item_id = ?
		AND inv.date >= ? AND inv.date <= ?
		GROUP BY il.item_id, i.name
	`

	var result struct {
		ProductID         uuid.UUID `gorm:"column:product_id"`
		ProductName       string    `gorm:"column:product_name"`
		TotalInputCost    float64   `gorm:"column:total_input_cost"`
		TotalSellingPrice float64   `gorm:"column:total_selling_price"`
		TotalTaxes        float64   `gorm:"column:total_taxes"`
	}

	err := r.db.WithContext(ctx).Raw(query, productID, startDate, endDate).Scan(&result).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // No records found
		}
		return nil, fmt.Errorf("failed to get margin report for product %s: %w", productID, err)
	}

	report := &margin.MarginReport{
		PeriodStart:       startDate,
		PeriodEnd:         endDate,
		ProductID:         result.ProductID,
		ProductName:       result.ProductName,
		TotalInputCost:    result.TotalInputCost,
		TotalServiceCost:  0, // Technical Debt: Service costs are not yet tracked.
		TotalTaxes:        result.TotalTaxes,
		TotalSellingPrice: result.TotalSellingPrice,
	}

	report.GrossMargin = report.TotalSellingPrice - (report.TotalInputCost + report.TotalServiceCost + report.TotalTaxes)
	if report.TotalSellingPrice > 0 {
		report.GrossMarginPercentage = (report.GrossMargin / report.TotalSellingPrice) * 100
	}

	return report, nil
}

// ListMarginReports retrieves margin reports for all products within a given period.
func (r *GormMarginRepository) ListMarginReports(ctx context.Context, startDate, endDate time.Time) ([]*margin.MarginReport, error) {
	query := `
		SELECT
			il.item_id AS product_id,
			i.name AS product_name,
			SUM(il.total_cost) AS total_input_cost,
			SUM(il.total_amount) AS total_selling_price,
			SUM(il.tax_amount * il.quantity) AS total_taxes
		FROM invoice_lines il
		JOIN items i ON il.item_id = i.id
		JOIN invoices inv ON il.invoice_id = inv.id
		WHERE inv.date >= ? AND inv.date <= ?
		GROUP BY il.item_id, i.name
		ORDER BY i.name
	`

	var results []struct {
		ProductID         uuid.UUID `gorm:"column:product_id"`
		ProductName       string    `gorm:"column:product_name"`
		TotalInputCost    float64   `gorm:"column:total_input_cost"`
		TotalSellingPrice float64   `gorm:"column:total_selling_price"`
		TotalTaxes        float64   `gorm:"column:total_taxes"`
	}

	err := r.db.WithContext(ctx).Raw(query, startDate, endDate).Scan(&results).Error
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
			TotalServiceCost:  0, // Technical Debt: Not tracked.
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