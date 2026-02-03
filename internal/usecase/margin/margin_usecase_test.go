package margin

import (
	"context"
	"errors"
	"testing"
	"time"

	"doligo_001/internal/domain/margin"
	"github.com/google/uuid"
)

// fakeMarginRepository is a simple fake for the margin.Repository for testing.
type fakeMarginRepository struct {
	reports map[uuid.UUID]*margin.MarginReport
}

// newFakeMarginRepository initializes a new fake repository.
func newFakeMarginRepository() *fakeMarginRepository {
	return &fakeMarginRepository{
		reports: make(map[uuid.UUID]*margin.MarginReport),
	}
}

func (f *fakeMarginRepository) GetMarginReport(ctx context.Context, productID uuid.UUID, startDate, endDate time.Time) (*margin.MarginReport, error) {
	if report, exists := f.reports[productID]; exists {
		// In a real scenario, you'd also check if the report is within the date range.
		// For this fake, we'll just return it if it exists.
		return report, nil
	}
	return nil, errors.New("margin report not found")
}

func (f *fakeMarginRepository) ListMarginReports(ctx context.Context, startDate, endDate time.Time) ([]*margin.MarginReport, error) {
	var reportList []*margin.MarginReport
	// Again, a real implementation would filter by date.
	for _, r := range f.reports {
		reportList = append(reportList, r)
	}
	if len(reportList) == 0 {
		return nil, errors.New("no margin reports found")
	}
	return reportList, nil
}

func TestMarginUsecase_GetProductMarginReport(t *testing.T) {
	repo := newFakeMarginRepository()
	usecase := NewMarginUsecase(repo)

	productID := uuid.New()
	repo.reports[productID] = &margin.MarginReport{
		ProductID:   productID,
		ProductName: "Test Product",
	}

	validStart := time.Now().Add(-24 * time.Hour)
	validEnd := time.Now()

	tests := []struct {
		name      string
		productID uuid.UUID
		startDate time.Time
		endDate   time.Time
		wantErr   bool
		errText   string
	}{
		{
			name:      "happy path - report found",
			productID: productID,
			startDate: validStart,
			endDate:   validEnd,
			wantErr:   false,
		},
		{
			name:      "error - invalid date range (start after end)",
			productID: productID,
			startDate: validEnd,
			endDate:   validStart,
			wantErr:   true,
			errText:   "invalid date range provided",
		},
		{
			name:      "error - zero start date",
			productID: productID,
			startDate: time.Time{},
			endDate:   validEnd,
			wantErr:   true,
			errText:   "invalid date range provided",
		},
		{
			name:      "error - repository error (not found)",
			productID: uuid.New(),
			startDate: validStart,
			endDate:   validEnd,
			wantErr:   true,
			errText:   "margin report not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := usecase.GetProductMarginReport(context.Background(), tt.productID, tt.startDate, tt.endDate)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetProductMarginReport() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.errText {
				t.Errorf("GetProductMarginReport() error text = %q, want %q", err.Error(), tt.errText)
			}
		})
	}
}

func TestMarginUsecase_ListOverallMarginReports(t *testing.T) {
	repo := newFakeMarginRepository()

	repo.reports[uuid.New()] = &margin.MarginReport{ProductName: "Product A"}
	repo.reports[uuid.New()] = &margin.MarginReport{ProductName: "Product B"}

	validStart := time.Now().Add(-24 * time.Hour)
	validEnd := time.Now()

	tests := []struct {
		name      string
		startDate time.Time
		endDate   time.Time
		setupRepo func() margin.Repository
		wantErr   bool
		errText   string
	}{
		{
			name:      "happy path - reports found",
			startDate: validStart,
			endDate:   validEnd,
			setupRepo: func() margin.Repository {
				return repo
			},
			wantErr: false,
		},
		{
			name:      "error - invalid date range",
			startDate: validEnd,
			endDate:   validStart,
			setupRepo: func() margin.Repository {
				return repo
			},
			wantErr: true,
			errText:   "invalid date range provided",
		},
		{
			name:      "error - repository error (no reports)",
			startDate: validStart,
			endDate:   validEnd,
			setupRepo: func() margin.Repository {
				// Return a new empty repo to simulate a repo error/no data
				return newFakeMarginRepository()
			},
			wantErr: true,
			errText:   "no margin reports found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := NewMarginUsecase(tt.setupRepo())
			_, err := uc.ListOverallMarginReports(context.Background(), tt.startDate, tt.endDate)
			if (err != nil) != tt.wantErr {
				t.Errorf("ListOverallMarginReports() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.errText {
				t.Errorf("ListOverallMarginReports() error text = %q, want %q", err.Error(), tt.errText)
			}
		})
	}
}
