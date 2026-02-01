package margin

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"doligo_001/internal/domain/margin"
)

// Usecase defines the business logic for Margin Dashboard operations.
type Usecase struct {
	marginRepo margin.Repository
}

// NewUsecase creates a new instance of Margin Usecase.
func NewUsecase(mr margin.Repository) *Usecase {
	return &Usecase{
		marginRepo: mr,
	}
}

// GetProductMarginReport retrieves the margin report for a specific product.
func (uc *Usecase) GetProductMarginReport(ctx context.Context, productID uuid.UUID, startDate, endDate time.Time) (*margin.MarginReport, error) {
	if startDate.IsZero() || endDate.IsZero() || startDate.After(endDate) {
		return nil, fmt.Errorf("invalid date range provided")
	}
	return uc.marginRepo.GetMarginReport(ctx, productID, startDate, endDate)
}

// ListOverallMarginReports retrieves margin reports for all products.
func (uc *Usecase) ListOverallMarginReports(ctx context.Context, startDate, endDate time.Time) ([]*margin.MarginReport, error) {
	if startDate.IsZero() || endDate.IsZero() || startDate.After(endDate) {
		return nil, fmt.Errorf("invalid date range provided")
	}
	return uc.marginRepo.ListMarginReports(ctx, startDate, endDate)
}