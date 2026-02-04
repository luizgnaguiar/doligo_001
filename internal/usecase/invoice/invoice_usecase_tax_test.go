package invoice_test

import (
	"context"
	"testing"

	"doligo_001/internal/api/dto"
	domain_invoice "doligo_001/internal/domain/invoice"
	"doligo_001/internal/domain/item"
	uc_invoice "doligo_001/internal/usecase/invoice"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mocks

type MockInvoiceRepo struct {
	mock.Mock
}

func (m *MockInvoiceRepo) Create(ctx context.Context, inv *domain_invoice.Invoice) error {
	args := m.Called(ctx, inv)
	return args.Error(0)
}

func (m *MockInvoiceRepo) Update(ctx context.Context, inv *domain_invoice.Invoice) error {
	args := m.Called(ctx, inv)
	return args.Error(0)
}

func (m *MockInvoiceRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain_invoice.Invoice, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain_invoice.Invoice), args.Error(1)
}

func (m *MockInvoiceRepo) FindByIDWithDetails(ctx context.Context, id uuid.UUID) (*domain_invoice.Invoice, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain_invoice.Invoice), args.Error(1)
}

type MockItemRepo struct {
	mock.Mock
}

func (m *MockItemRepo) GetByID(ctx context.Context, id uuid.UUID) (*item.Item, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*item.Item), args.Error(1)
}

// Add other ItemRepo methods if needed, stubbing them for now
func (m *MockItemRepo) Create(ctx context.Context, i *item.Item) error { return nil }
func (m *MockItemRepo) Update(ctx context.Context, i *item.Item) error { return nil }
func (m *MockItemRepo) List(ctx context.Context, limit, offset int) ([]*item.Item, error) { return nil, nil }
func (m *MockItemRepo) Delete(ctx context.Context, id uuid.UUID) error { return nil }

type MockPDFGen struct {
	mock.Mock
}

func (m *MockPDFGen) Generate(ctx context.Context, inv *domain_invoice.Invoice) ([]byte, error) {
	args := m.Called(ctx, inv)
	return args.Get(0).([]byte), args.Error(1)
}

type MockEmailSender struct {
	mock.Mock
}

func (m *MockEmailSender) Send(ctx context.Context, to, subject, body string) error {
	return nil
}

// Tests

func TestCreateInvoice_TaxCalculation(t *testing.T) {
	// Setup
	mockInvoiceRepo := new(MockInvoiceRepo)
	mockItemRepo := new(MockItemRepo)
	mockPDFGen := new(MockPDFGen)
	mockEmailSender := new(MockEmailSender)

	usecase := uc_invoice.NewUsecase(mockInvoiceRepo, mockItemRepo, mockPDFGen, mockEmailSender, nil, "storage/pdfs")

	ctx := context.Background()
	thirdPartyID := uuid.New().String()
	itemID := uuid.New().String()
	
	req := &dto.CreateInvoiceRequest{
		ThirdPartyID: thirdPartyID,
		Number:       "INV-001",
		Date:         "2023-10-27",
		Lines: []dto.CreateInvoiceLineRequest{
			{
				ItemID:      itemID,
				Description: "Test Item",
				Quantity:    1,
				UnitPrice:   100.0,
				TaxRate:     10.0, // 10%
			},
		},
	}

	// Mock Item Response
	mockItemRepo.On("GetByID", ctx, mock.Anything).Return(&item.Item{
		ID:        uuid.MustParse(itemID),
		CostPrice: 50.0,
	}, nil)

	// Mock Invoice Create
	mockInvoiceRepo.On("Create", ctx, mock.MatchedBy(func(inv *domain_invoice.Invoice) bool {
		// Validation Logic
		if len(inv.Lines) != 1 {
			return false
		}
		line := inv.Lines[0]
		
		// Expected Calculation:
		// TaxAmount = 100 * (10/100) = 10
		// NetPrice = 100 + 10 = 110
		// TotalAmount (Line) = 1 * 110 = 110
		// TotalTax (Invoice) = 10
		// TotalAmount (Invoice) = 110
		
		ok := true
		ok = ok && assert.InDelta(t, 10.0, line.TaxAmount, 0.001)
		ok = ok && assert.InDelta(t, 110.0, line.NetPrice, 0.001)
		ok = ok && assert.InDelta(t, 110.0, line.TotalAmount, 0.001)
		ok = ok && assert.InDelta(t, 110.0, inv.TotalAmount, 0.001)
		ok = ok && assert.InDelta(t, 10.0, inv.TotalTax, 0.001)
		
		return ok
	})).Return(nil)

	// Execute
	createdInvoice, err := usecase.Create(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, createdInvoice)
	mockItemRepo.AssertExpectations(t)
	mockInvoiceRepo.AssertExpectations(t)
}
