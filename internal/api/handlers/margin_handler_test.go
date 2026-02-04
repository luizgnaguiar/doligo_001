package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"doligo_001/internal/api/validator"
	"doligo_001/internal/domain/margin"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

// MockMarginUsecase is a mock implementation of margin.Usecase
type MockMarginUsecase struct {
	GetProductMarginReportFunc   func(ctx context.Context, productID uuid.UUID, startDate, endDate time.Time) (*margin.MarginReport, error)
	ListOverallMarginReportsFunc func(ctx context.Context, startDate, endDate time.Time) ([]*margin.MarginReport, error)
}

func (m *MockMarginUsecase) GetProductMarginReport(ctx context.Context, productID uuid.UUID, startDate, endDate time.Time) (*margin.MarginReport, error) {
	if m.GetProductMarginReportFunc != nil {
		return m.GetProductMarginReportFunc(ctx, productID, startDate, endDate)
	}
	return nil, nil
}

func (m *MockMarginUsecase) ListOverallMarginReports(ctx context.Context, startDate, endDate time.Time) ([]*margin.MarginReport, error) {
	if m.ListOverallMarginReportsFunc != nil {
		return m.ListOverallMarginReportsFunc(ctx, startDate, endDate)
	}
	return nil, nil
}

func TestGetProductMarginReport_Validation(t *testing.T) {
	e := echo.New()
	e.Validator = validator.NewValidator()
	mockUsecase := &MockMarginUsecase{}
	handler := NewMarginHandler(mockUsecase)

	t.Run("Valid Request", func(t *testing.T) {
		productID := uuid.New().String()
		req := httptest.NewRequest(http.MethodGet, "/margin/products/"+productID+"?startDate=2023-01-01&endDate=2023-01-31", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/margin/products/:productID")
		c.SetParamNames("productID")
		c.SetParamValues(productID)

		mockUsecase.GetProductMarginReportFunc = func(ctx context.Context, pid uuid.UUID, s, e time.Time) (*margin.MarginReport, error) {
			assert.Equal(t, productID, pid.String())
			assert.Equal(t, "2023-01-01", s.Format("2006-01-02"))
			return &margin.MarginReport{}, nil
		}

		err := handler.GetProductMarginReport(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("Invalid UUID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/margin/products/invalid-uuid?startDate=2023-01-01&endDate=2023-01-31", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/margin/products/:productID")
		c.SetParamNames("productID")
		c.SetParamValues("invalid-uuid")

		err := handler.GetProductMarginReport(c)
		assert.NoError(t, err) // Handlers return error via c.JSON usually, or return error object. 
		// My handler implementation returns c.JSON(400) which returns nil error to Echo?
		// Wait, let's check the handler code again. 
		// "return c.JSON(http.StatusBadRequest, ...)" returns nil error.
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("Invalid Date", func(t *testing.T) {
		productID := uuid.New().String()
		req := httptest.NewRequest(http.MethodGet, "/margin/products/"+productID+"?startDate=invalid-date&endDate=2023-01-31", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/margin/products/:productID")
		c.SetParamNames("productID")
		c.SetParamValues(productID)

		err := handler.GetProductMarginReport(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}

func TestListOverallMarginReports_Validation(t *testing.T) {
	e := echo.New()
	e.Validator = validator.NewValidator()
	mockUsecase := &MockMarginUsecase{}
	handler := NewMarginHandler(mockUsecase)

	t.Run("Valid Request", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/margin?startDate=2023-01-01&endDate=2023-01-31", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		mockUsecase.ListOverallMarginReportsFunc = func(ctx context.Context, s, e time.Time) ([]*margin.MarginReport, error) {
			return []*margin.MarginReport{}, nil
		}

		err := handler.ListOverallMarginReports(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("Invalid Date", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/margin?startDate=2023/01/01&endDate=2023-01-31", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := handler.ListOverallMarginReports(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}
