package margin

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"doligo_001/internal/domain/margin"
)

// MarginUsecase defines the interface for Margin related business logic.
type MarginUsecase interface {
	GetProductMarginReport(ctx context.Context, productID uuid.UUID, startDate, endDate time.Time) (*margin.MarginReport, error)
	ListOverallMarginReports(ctx context.Context, startDate, endDate time.Time) ([]*margin.MarginReport, error)
	// Add other Margin related methods here as they are defined.
}

type marginUsecase struct {
	marginRepo margin.Repository
}

// NewMarginUsecase creates a new instance of Margin Usecase.
func NewMarginUsecase(mr margin.Repository) MarginUsecase {
	return &marginUsecase{
		marginRepo: mr,
	}
}

// GetProductMarginReport retrieves the margin report for a specific product.
func (uc *marginUsecase) GetProductMarginReport(ctx context.Context, productID uuid.UUID, startDate, endDate time.Time) (*margin.MarginReport, error) {
	if startDate.IsZero() || endDate.IsZero() || startDate.After(endDate) {
		return nil, fmt.Errorf("invalid date range provided")
	}
	return uc.marginRepo.GetMarginReport(ctx, productID, startDate, endDate)
}

// ListOverallMarginReports retrieves margin reports for all products.
func (uc *marginUsecase) ListOverallMarginReports(ctx context.Context, startDate, endDate time.Time) ([]*margin.MarginReport, error) {
	if startDate.IsZero() || endDate.IsZero() || startDate.After(endDate) {
		return nil, fmt.Errorf("invalid date range provided")
	}
	return uc.marginRepo.ListMarginReports(ctx, startDate, endDate)
}