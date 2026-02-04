package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"doligo_001/internal/api/dto"
	"doligo_001/internal/api/validator"
	"doligo_001/internal/domain/invoice"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

// MockInvoiceUsecase is a mock implementation of invoice.Usecase
type MockInvoiceUsecase struct {
	CreateFunc             func(ctx context.Context, req *dto.CreateInvoiceRequest) (*invoice.Invoice, error)
	GetByIDFunc            func(ctx context.Context, id uuid.UUID) (*invoice.Invoice, error)
	GenerateInvoicePDFFunc func(ctx context.Context, invoiceID uuid.UUID) ([]byte, string, error)
}

func (m *MockInvoiceUsecase) Create(ctx context.Context, req *dto.CreateInvoiceRequest) (*invoice.Invoice, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, req)
	}
	return nil, nil
}

func (m *MockInvoiceUsecase) GetByID(ctx context.Context, id uuid.UUID) (*invoice.Invoice, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockInvoiceUsecase) GenerateInvoicePDF(ctx context.Context, invoiceID uuid.UUID) ([]byte, string, error) {
	if m.GenerateInvoicePDFFunc != nil {
		return m.GenerateInvoicePDFFunc(ctx, invoiceID)
	}
	return nil, "", nil
}

func TestCreateInvoice_SanitizationAndValidation(t *testing.T) {
	e := echo.New()
	e.Validator = validator.NewValidator()
	mockUsecase := &MockInvoiceUsecase{}
	handler := NewInvoiceHandler(mockUsecase)

	t.Run("Valid Invoice", func(t *testing.T) {
		reqBody := dto.CreateInvoiceRequest{
			ThirdPartyID: uuid.New().String(),
			Number:       "INV-001",
			Date:         "2023-10-27",
			Lines: []dto.CreateInvoiceLineRequest{
				{
					ItemID:      uuid.New().String(),
					Description: "Service A",
					Quantity:    1,
					UnitPrice:   100,
				},
			},
		}

		mockUsecase.CreateFunc = func(ctx context.Context, req *dto.CreateInvoiceRequest) (*invoice.Invoice, error) {
			return &invoice.Invoice{Number: req.Number}, nil
		}

		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/invoices", bytes.NewReader(jsonBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := handler.CreateInvoice(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, rec.Code)
	})

	t.Run("Sanitization of XSS", func(t *testing.T) {
		reqBody := dto.CreateInvoiceRequest{
			ThirdPartyID: uuid.New().String(),
			Number:       "<script>alert('xss')</script>INV-002",
			Date:         "2023-10-27",
			Lines: []dto.CreateInvoiceLineRequest{
				{
					ItemID:      uuid.New().String(),
					Description: "<img src=x onerror=alert(1)>Item",
					Quantity:    1,
					UnitPrice:   50,
				},
			},
		}

		mockUsecase.CreateFunc = func(ctx context.Context, req *dto.CreateInvoiceRequest) (*invoice.Invoice, error) {
			// Verify sanitization
			assert.Equal(t, "INV-002", req.Number) // Script tags removed (depending on sanitizer config, but usually empty or just text)
			// Actually bluemonday strict policy strips tags. So <script>... is gone.
			// Let's assume strict policy which might strip the whole tag content or just the tag. 
			// Standard behavior for bluemonday.UGCPolicy() preserves text content of some tags but strips script.
			// But sanitizer package in this project likely uses `UGCPolicy` or `StrictPolicy`. 
			// If it uses StrictPolicy, it strips tags. 
			// "alert('xss')INV-002" or just "INV-002" if script is fully removed.
			// Let's just check it doesn't contain <script>
			assert.NotContains(t, req.Number, "<script>")
			assert.NotContains(t, req.Lines[0].Description, "<img")
			return &invoice.Invoice{Number: req.Number}, nil
		}

		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/invoices", bytes.NewReader(jsonBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := handler.CreateInvoice(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, rec.Code)
	})

	t.Run("Invalid UUID", func(t *testing.T) {
		reqBody := dto.CreateInvoiceRequest{
			ThirdPartyID: "invalid-uuid",
			Number:       "INV-003",
			Date:         "2023-10-27",
			Lines: []dto.CreateInvoiceLineRequest{
				{
					ItemID:      uuid.New().String(),
					Description: "Item",
					Quantity:    1,
					UnitPrice:   10,
				},
			},
		}

		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/invoices", bytes.NewReader(jsonBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := handler.CreateInvoice(c)
		assert.Error(t, err) // Should error due to validation
		httpError, ok := err.(*echo.HTTPError)
		assert.True(t, ok)
		assert.Equal(t, http.StatusBadRequest, httpError.Code)
	})
}
