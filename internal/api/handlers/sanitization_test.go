package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"doligo_001/internal/api/dto"
	"doligo_001/internal/api/handlers"
	"doligo_001/internal/api/validator"
	domainItem "doligo_001/internal/domain/item"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockItemUsecase for testing
type MockItemUsecase struct {
	mock.Mock
}

func (m *MockItemUsecase) Create(ctx context.Context, req *dto.CreateItemRequest) (*domainItem.Item, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domainItem.Item), args.Error(1)
}

func (m *MockItemUsecase) GetByID(ctx context.Context, id uuid.UUID) (*domainItem.Item, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*domainItem.Item), args.Error(1)
}

func (m *MockItemUsecase) Update(ctx context.Context, id uuid.UUID, req *dto.UpdateItemRequest) (*domainItem.Item, error) {
	args := m.Called(ctx, id, req)
	return args.Get(0).(*domainItem.Item), args.Error(1)
}

func (m *MockItemUsecase) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockItemUsecase) List(ctx context.Context) ([]*domainItem.Item, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*domainItem.Item), args.Error(1)
}

func TestItemSanitization(t *testing.T) {
	e := echo.New()
	e.Validator = validator.NewValidator()
	
	mockUsecase := new(MockItemUsecase)
	h := handlers.NewItemHandler(mockUsecase)

	t.Run("Sanitize Name and Description", func(t *testing.T) {
		reqBody := dto.CreateItemRequest{
			Name:        "Dirty <script>alert('xss')</script> Name",
			Description: "Dirty <b>Description</b>",
			Type:        "STORABLE",
			CostPrice:   10,
			SalePrice:   20,
		}
		
		// For the mock, we expect the sanitized version.
		// bluemonday.StrictPolicy() removes <script> tags AND their content usually.
		mockUsecase.On("Create", mock.Anything, mock.MatchedBy(func(req *dto.CreateItemRequest) bool {
			// "Dirty <script>alert('xss')</script> Name" -> "Dirty  Name" (script and content removed)
			// "Dirty <b>Description</b>" -> "Dirty Description" (tags removed, content kept)
			return req.Name == "Dirty  Name" && req.Description == "Dirty Description"
		})).Return(&domainItem.Item{ID: uuid.New(), Name: "Dirty  Name"}, nil)

		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/items", bytes.NewReader(jsonBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := h.Create(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, rec.Code)
	})
}

func TestUUIDValidation_BOM(t *testing.T) {
	e := echo.New()
	e.Validator = validator.NewValidator()
	
	// We don't need a full handler for this, just testing the validator on the struct would suffice,
	// but let's test the validator logic via a dummy handler validation call or similar.
	// Actually, we can just test the struct validation directly.

	v := validator.NewValidator()

	t.Run("Invalid UUID in BOM Request", func(t *testing.T) {
		req := dto.CreateBOMRequest{
			ProductID: "invalid-uuid",
			Name: "Valid Name",
			Components: []dto.BOMComponentRequest{
				{ComponentItemID: uuid.New().String(), Quantity: 1, UnitOfMeasure: "pcs"},
			},
		}
		
		err := v.Validate(req)
		assert.Error(t, err)
		// We expect an error related to ProductID uuid tag
	})

	t.Run("Valid UUID in BOM Request", func(t *testing.T) {
		req := dto.CreateBOMRequest{
			ProductID: uuid.New().String(),
			Name: "Valid Name",	
			Components: []dto.BOMComponentRequest{
				{ComponentItemID: uuid.New().String(), Quantity: 1, UnitOfMeasure: "pcs"},
			},
		}
		
		err := v.Validate(req)
		assert.NoError(t, err)
	})
}
