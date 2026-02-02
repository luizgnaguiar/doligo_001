package stock_test

import (
	"context"
	"errors"
	"testing"

	"doligo_001/internal/domain"
	"doligo_001/internal/domain/item"
	"doligo_001/internal/domain/stock"
	usecase "doligo_001/internal/usecase/stock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// --- Mock Implementations for Repositories and Transactioner ---

// MockTransactioner
type MockTransactioner struct {
	mock.Mock
}

func (m *MockTransactioner) Transaction(ctx context.Context, fc func(tx *gorm.DB) error) error {
	args := m.Called(ctx, fc)
	// Execute the provided function with a dummy *gorm.DB for testing purposes
	// In a real scenario, you might pass a mock *gorm.DB or nil if the fc doesn't use it much
	err := fc(nil) // Pass nil or a mock gorm.DB
	if err != nil {
		return err
	}
	return args.Error(0)
}

// MockItemRepository
type MockItemRepository struct {
	mock.Mock
}

func (m *MockItemRepository) WithTx(tx *gorm.DB) item.Repository {
	args := m.Called(tx)
	if args.Get(0) == nil {
		return m // Return self if no specific transactional mock is set
	}
	return args.Get(0).(item.Repository)
}

func (m *MockItemRepository) Create(ctx context.Context, item *item.Item) error {
	args := m.Called(ctx, item)
	return args.Error(0)
}

func (m *MockItemRepository) GetByID(ctx context.Context, id uuid.UUID) (*item.Item, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*item.Item), args.Error(1)
}

func (m *MockItemRepository) Update(ctx context.Context, item *item.Item) error {
	args := m.Called(ctx, item)
	return args.Error(0)
}

func (m *MockItemRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockItemRepository) List(ctx context.Context) ([]*item.Item, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*item.Item), args.Error(1)
}

// MockWarehouseRepository
type MockWarehouseRepository struct {
	mock.Mock
}

func (m *MockWarehouseRepository) WithTx(tx *gorm.DB) stock.WarehouseRepository {
	args := m.Called(tx)
	if args.Get(0) == nil {
		return m
	}
	return args.Get(0).(stock.WarehouseRepository)
}
func (m *MockWarehouseRepository) Create(ctx context.Context, warehouse *stock.Warehouse) error {
	args := m.Called(ctx, warehouse)
	return args.Error(0)
}
func (m *MockWarehouseRepository) GetByID(ctx context.Context, id uuid.UUID) (*stock.Warehouse, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*stock.Warehouse), args.Error(1)
}
func (m *MockWarehouseRepository) Update(ctx context.Context, warehouse *stock.Warehouse) error {
	args := m.Called(ctx, warehouse)
	return args.Error(0)
}
func (m *MockWarehouseRepository) List(ctx context.Context) ([]*stock.Warehouse, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*stock.Warehouse), args.Error(1)
}
func (m *MockWarehouseRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// MockBinRepository
type MockBinRepository struct {
	mock.Mock
}

func (m *MockBinRepository) WithTx(tx *gorm.DB) stock.BinRepository {
	args := m.Called(tx)
	if args.Get(0) == nil {
		return m
	}
	return args.Get(0).(stock.BinRepository)
}
func (m *MockBinRepository) Create(ctx context.Context, bin *stock.Bin) error {
	args := m.Called(ctx, bin)
	return args.Error(0)
}
func (m *MockBinRepository) GetByID(ctx context.Context, id uuid.UUID) (*stock.Bin, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*stock.Bin), args.Error(1)
}
func (m *MockBinRepository) Update(ctx context.Context, bin *stock.Bin) error {
	args := m.Called(ctx, bin)
	return args.Error(0)
}
func (m *MockBinRepository) ListByWarehouse(ctx context.Context, warehouseID uuid.UUID) ([]*stock.Bin, error) {
	args := m.Called(ctx, warehouseID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*stock.Bin), args.Error(1)
}
func (m *MockBinRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// MockStockRepository
type MockStockRepository struct {
	mock.Mock
}

func (m *MockStockRepository) WithTx(tx *gorm.DB) stock.StockRepository {
	args := m.Called(tx)
	if args.Get(0) == nil {
		return m
	}
	return args.Get(0).(stock.StockRepository)
}
func (m *MockStockRepository) GetStock(ctx context.Context, itemID, warehouseID uuid.UUID, binID *uuid.UUID) (*stock.Stock, error) {
	args := m.Called(ctx, itemID, warehouseID, binID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*stock.Stock), args.Error(1)
}
func (m *MockStockRepository) GetStockForUpdate(ctx context.Context, itemID, warehouseID uuid.UUID, binID *uuid.UUID) (*stock.Stock, error) {
	args := m.Called(ctx, itemID, warehouseID, binID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*stock.Stock), args.Error(1)
}
func (m *MockStockRepository) UpsertStock(ctx context.Context, stock *stock.Stock) error {
	args := m.Called(ctx, stock)
	return args.Error(0)
}

// MockStockMovementRepository
type MockStockMovementRepository struct {
	mock.Mock
}

func (m *MockStockMovementRepository) WithTx(tx *gorm.DB) stock.StockMovementRepository {
	args := m.Called(tx)
	if args.Get(0) == nil {
		return m
	}
	return args.Get(0).(stock.StockMovementRepository)
}
func (m *MockStockMovementRepository) Create(ctx context.Context, movement *stock.StockMovement) error {
	args := m.Called(ctx, movement)
	return args.Error(0)
}

// MockStockLedgerRepository
type MockStockLedgerRepository struct {
	mock.Mock
}

func (m *MockStockLedgerRepository) WithTx(tx *gorm.DB) stock.StockLedgerRepository {
	args := m.Called(tx)
	if args.Get(0) == nil {
		return m
	}
	return args.Get(0).(stock.StockLedgerRepository)
}
func (m *MockStockLedgerRepository) Create(ctx context.Context, entry *stock.StockLedger) error {
	args := m.Called(ctx, entry)
	return args.Error(0)
}

// --- Test Suite Setup ---

type stockUseCaseTestSuite struct {
	txManager       *MockTransactioner
	itemRepo        *MockItemRepository
	warehouseRepo   *MockWarehouseRepository
	binRepo         *MockBinRepository
	stockRepo       *MockStockRepository
	stockMoveRepo   *MockStockMovementRepository
	stockLedgerRepo *MockStockLedgerRepository
	useCase         usecase.UseCase
	ctx             context.Context
	userID          uuid.UUID
	itemID          uuid.UUID
	warehouseID     uuid.UUID
	binID           uuid.UUID
}

func setupTestSuite() *stockUseCaseTestSuite {
	s := &stockUseCaseTestSuite{
		txManager:       new(MockTransactioner),
		itemRepo:        new(MockItemRepository),
		warehouseRepo:   new(MockWarehouseRepository),
		binRepo:         new(MockBinRepository),
		stockRepo:       new(MockStockRepository),
		stockMoveRepo:   new(MockStockMovementRepository),
		stockLedgerRepo: new(MockStockLedgerRepository),
		userID:          uuid.New(),
		itemID:          uuid.New(),
		warehouseID:     uuid.New(),
		binID:           uuid.New(),
	}

	s.useCase = usecase.NewUseCase(
		s.txManager,
		s.stockRepo,
		s.stockMoveRepo,
		s.stockLedgerRepo,
		s.warehouseRepo,
		s.binRepo,
		s.itemRepo,
	)

	// Context with UserID
	s.ctx = context.WithValue(context.Background(), domain.UserIDKey, s.userID)

	return s
}

// --- Tests for CreateStockMovement ---

func TestCreateStockMovement_In_HappyPath_NoExistingStock(t *testing.T) {
	s := setupTestSuite()
	defer s.txManager.AssertExpectations(t)
	defer s.itemRepo.AssertExpectations(t)
	defer s.warehouseRepo.AssertExpectations(t)
	defer s.binRepo.AssertExpectations(t)
	defer s.stockRepo.AssertExpectations(t)
	defer s.stockMoveRepo.AssertExpectations(t)
	defer s.stockLedgerRepo.AssertExpectations(t)

	mockItem := &item.Item{ID: s.itemID, Name: "Test Item"}
	mockWarehouse := &stock.Warehouse{ID: s.warehouseID, Name: "Main Warehouse", IsActive: true}

	s.txManager.On("Transaction", mock.Anything, mock.AnythingOfType("func(*gorm.DB) error")).Return(nil).Once()
	s.itemRepo.On("WithTx", mock.Anything).Return(s.itemRepo).Maybe()
	s.itemRepo.On("GetByID", s.ctx, s.itemID).Return(mockItem, nil).Once()
	s.warehouseRepo.On("WithTx", mock.Anything).Return(s.warehouseRepo).Maybe()
	s.warehouseRepo.On("GetByID", s.ctx, s.warehouseID).Return(mockWarehouse, nil).Once()
	s.binRepo.On("WithTx", mock.Anything).Return(s.binRepo).Maybe()
	s.stockRepo.On("WithTx", mock.Anything).Return(s.stockRepo).Maybe()
	s.stockRepo.On("GetStockForUpdate", s.ctx, s.itemID, s.warehouseID, mock.Anything).Return(nil, gorm.ErrRecordNotFound).Once()
	s.stockMoveRepo.On("WithTx", mock.Anything).Return(s.stockMoveRepo).Maybe()
	s.stockMoveRepo.On("Create", s.ctx, mock.AnythingOfType("*stock.StockMovement")).Return(nil).Once()
	s.stockRepo.On("UpsertStock", s.ctx, mock.AnythingOfType("*stock.Stock")).Return(nil).Once()
	s.stockLedgerRepo.On("WithTx", mock.Anything).Return(s.stockLedgerRepo).Maybe()
	s.stockLedgerRepo.On("Create", s.ctx, mock.AnythingOfType("*stock.StockLedger")).Return(nil).Once()

	movement, err := s.useCase.CreateStockMovement(s.ctx, s.itemID, s.warehouseID, nil, stock.MovementTypeIn, 10.0, "Initial Stock")

	assert.NoError(t, err)
	assert.NotNil(t, movement)
	assert.Equal(t, s.itemID, movement.ItemID)
	assert.Equal(t, s.warehouseID, movement.WarehouseID)
	assert.Nil(t, movement.BinID)
	assert.Equal(t, stock.MovementTypeIn, movement.Type)
	assert.Equal(t, 10.0, movement.Quantity)
	assert.Equal(t, s.userID, movement.CreatedBy)

	// Verify UpsertStock was called with the correct final quantity
	s.stockRepo.AssertCalled(t, "UpsertStock", s.ctx, mock.MatchedBy(func(st *stock.Stock) bool {
		return st.ItemID == s.itemID && st.WarehouseID == s.warehouseID && st.BinID == nil && st.Quantity == 10.0
	}))

	// Verify StockLedger was called with the correct quantities
	s.stockLedgerRepo.AssertCalled(t, "Create", s.ctx, mock.MatchedBy(func(sl *stock.StockLedger) bool {
		return sl.ItemID == s.itemID && sl.QuantityBefore == 0.0 && sl.QuantityAfter == 10.0 && sl.MovementType == stock.MovementTypeIn
	}))
}

func TestCreateStockMovement_In_HappyPath_ExistingStock(t *testing.T) {
	s := setupTestSuite()
	defer s.txManager.AssertExpectations(t)
	defer s.itemRepo.AssertExpectations(t)
	defer s.warehouseRepo.AssertExpectations(t)
	defer s.binRepo.AssertExpectations(t)
	defer s.stockRepo.AssertExpectations(t)
	defer s.stockMoveRepo.AssertExpectations(t)
	defer s.stockLedgerRepo.AssertExpectations(t)

	mockItem := &item.Item{ID: s.itemID, Name: "Test Item"}
	mockWarehouse := &stock.Warehouse{ID: s.warehouseID, Name: "Main Warehouse", IsActive: true}
	existingStock := &stock.Stock{ItemID: s.itemID, WarehouseID: s.warehouseID, Quantity: 5.0}

	s.txManager.On("Transaction", mock.Anything, mock.AnythingOfType("func(*gorm.DB) error")).Return(nil).Once()
	s.itemRepo.On("WithTx", mock.Anything).Return(s.itemRepo).Maybe()
	s.itemRepo.On("GetByID", s.ctx, s.itemID).Return(mockItem, nil).Once()
	s.warehouseRepo.On("WithTx", mock.Anything).Return(s.warehouseRepo).Maybe()
	s.warehouseRepo.On("GetByID", s.ctx, s.warehouseID).Return(mockWarehouse, nil).Once()
	s.binRepo.On("WithTx", mock.Anything).Return(s.binRepo).Maybe()
	s.stockRepo.On("WithTx", mock.Anything).Return(s.stockRepo).Maybe()
	s.stockRepo.On("GetStockForUpdate", s.ctx, s.itemID, s.warehouseID, mock.Anything).Return(existingStock, nil).Once()
	s.stockMoveRepo.On("WithTx", mock.Anything).Return(s.stockMoveRepo).Maybe()
	s.stockMoveRepo.On("Create", s.ctx, mock.AnythingOfType("*stock.StockMovement")).Return(nil).Once()
	s.stockRepo.On("UpsertStock", s.ctx, mock.AnythingOfType("*stock.Stock")).Return(nil).Once()
	s.stockLedgerRepo.On("WithTx", mock.Anything).Return(s.stockLedgerRepo).Maybe()
	s.stockLedgerRepo.On("Create", s.ctx, mock.AnythingOfType("*stock.StockLedger")).Return(nil).Once()

	movement, err := s.useCase.CreateStockMovement(s.ctx, s.itemID, s.warehouseID, nil, stock.MovementTypeIn, 10.0, "Adding Stock")

	assert.NoError(t, err)
	assert.NotNil(t, movement)
	assert.Equal(t, 10.0, movement.Quantity)

	// Verify UpsertStock was called with the correct final quantity (5.0 + 10.0 = 15.0)
	s.stockRepo.AssertCalled(t, "UpsertStock", s.ctx, mock.MatchedBy(func(st *stock.Stock) bool {
		return st.ItemID == s.itemID && st.WarehouseID == s.warehouseID && st.BinID == nil && st.Quantity == 15.0
	}))

	// Verify StockLedger was called with the correct quantities
	s.stockLedgerRepo.AssertCalled(t, "Create", s.ctx, mock.MatchedBy(func(sl *stock.StockLedger) bool {
		return sl.ItemID == s.itemID && sl.QuantityBefore == 5.0 && sl.QuantityAfter == 15.0 && sl.MovementType == stock.MovementTypeIn
	}))
}

func TestCreateStockMovement_Out_HappyPath_SufficientStock(t *testing.T) {
	s := setupTestSuite()
	defer s.txManager.AssertExpectations(t)
	defer s.itemRepo.AssertExpectations(t)
	defer s.warehouseRepo.AssertExpectations(t)
	defer s.binRepo.AssertExpectations(t)
	defer s.stockRepo.AssertExpectations(t)
	defer s.stockMoveRepo.AssertExpectations(t)
	defer s.stockLedgerRepo.AssertExpectations(t)

	mockItem := &item.Item{ID: s.itemID, Name: "Test Item"}
	mockWarehouse := &stock.Warehouse{ID: s.warehouseID, Name: "Main Warehouse", IsActive: true}
	existingStock := &stock.Stock{ItemID: s.itemID, WarehouseID: s.warehouseID, Quantity: 20.0}

	s.txManager.On("Transaction", mock.Anything, mock.AnythingOfType("func(*gorm.DB) error")).Return(nil).Once()
	s.itemRepo.On("WithTx", mock.Anything).Return(s.itemRepo).Maybe()
	s.itemRepo.On("GetByID", s.ctx, s.itemID).Return(mockItem, nil).Once()
	s.warehouseRepo.On("WithTx", mock.Anything).Return(s.warehouseRepo).Maybe()
	s.warehouseRepo.On("GetByID", s.ctx, s.warehouseID).Return(mockWarehouse, nil).Once()
	s.binRepo.On("WithTx", mock.Anything).Return(s.binRepo).Maybe()
	s.stockRepo.On("WithTx", mock.Anything).Return(s.stockRepo).Maybe()
	s.stockRepo.On("GetStockForUpdate", s.ctx, s.itemID, s.warehouseID, mock.Anything).Return(existingStock, nil).Once()
	s.stockMoveRepo.On("WithTx", mock.Anything).Return(s.stockMoveRepo).Maybe()
	s.stockMoveRepo.On("Create", s.ctx, mock.AnythingOfType("*stock.StockMovement")).Return(nil).Once()
	s.stockRepo.On("UpsertStock", s.ctx, mock.AnythingOfType("*stock.Stock")).Return(nil).Once()
	s.stockLedgerRepo.On("WithTx", mock.Anything).Return(s.stockLedgerRepo).Maybe()
	s.stockLedgerRepo.On("Create", s.ctx, mock.AnythingOfType("*stock.StockLedger")).Return(nil).Once()

	movement, err := s.useCase.CreateStockMovement(s.ctx, s.itemID, s.warehouseID, nil, stock.MovementTypeOut, 10.0, "Selling Item")

	assert.NoError(t, err)
	assert.NotNil(t, movement)
	assert.Equal(t, 10.0, movement.Quantity)

	// Verify UpsertStock was called with the correct final quantity (20.0 - 10.0 = 10.0)
	s.stockRepo.AssertCalled(t, "UpsertStock", s.ctx, mock.MatchedBy(func(st *stock.Stock) bool {
		return st.ItemID == s.itemID && st.WarehouseID == s.warehouseID && st.BinID == nil && st.Quantity == 10.0
	}))

	// Verify StockLedger was called with the correct quantities
	s.stockLedgerRepo.AssertCalled(t, "Create", s.ctx, mock.MatchedBy(func(sl *stock.StockLedger) bool {
		return sl.ItemID == s.itemID && sl.QuantityBefore == 20.0 && sl.QuantityAfter == 10.0 && sl.MovementType == stock.MovementTypeOut
	}))
}

func TestCreateStockMovement_Out_InsufficientStock(t *testing.T) {
	s := setupTestSuite()
	defer s.txManager.AssertExpectations(t)
	defer s.itemRepo.AssertExpectations(t)
	defer s.warehouseRepo.AssertExpectations(t)
	defer s.binRepo.AssertExpectations(t)
	defer s.stockRepo.AssertExpectations(t)
	defer s.stockMoveRepo.AssertExpectations(t)
	defer s.stockLedgerRepo.AssertExpectations(t)

	mockItem := &item.Item{ID: s.itemID, Name: "Test Item"}
	mockWarehouse := &stock.Warehouse{ID: s.warehouseID, Name: "Main Warehouse", IsActive: true}
	existingStock := &stock.Stock{ItemID: s.itemID, WarehouseID: s.warehouseID, Quantity: 5.0}

	s.txManager.On("Transaction", mock.Anything, mock.AnythingOfType("func(*gorm.DB) error")).Return(errors.New("transaction failed due to insufficient stock")).Once() // Simulate transaction rollback
	s.itemRepo.On("WithTx", mock.Anything).Return(s.itemRepo).Maybe()
	s.itemRepo.On("GetByID", s.ctx, s.itemID).Return(mockItem, nil).Once()
	s.warehouseRepo.On("WithTx", mock.Anything).Return(s.warehouseRepo).Maybe()
	s.warehouseRepo.On("GetByID", s.ctx, s.warehouseID).Return(mockWarehouse, nil).Once()
	s.binRepo.On("WithTx", mock.Anything).Return(s.binRepo).Maybe()
	s.stockRepo.On("WithTx", mock.Anything).Return(s.stockRepo).Maybe()
	s.stockRepo.On("GetStockForUpdate", s.ctx, s.itemID, s.warehouseID, mock.Anything).Return(existingStock, nil).Once()

	movement, err := s.useCase.CreateStockMovement(s.ctx, s.itemID, s.warehouseID, nil, stock.MovementTypeOut, 10.0, "Selling Item")

	assert.ErrorIs(t, err, usecase.ErrInsufficientStock)
	assert.Nil(t, movement)

	// Ensure no further repository calls are made after insufficient stock check
	s.stockMoveRepo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
	s.stockRepo.AssertNotCalled(t, "UpsertStock", mock.Anything, mock.Anything)
	s.stockLedgerRepo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

func TestCreateStockMovement_ItemNotFound(t *testing.T) {
	s := setupTestSuite()
	defer s.txManager.AssertExpectations(t)
	defer s.itemRepo.AssertExpectations(t)
	defer s.warehouseRepo.AssertExpectations(t)

	s.txManager.On("Transaction", mock.Anything, mock.AnythingOfType("func(*gorm.DB) error")).Return(errors.New("transaction failed: item not found")).Once()
	s.itemRepo.On("WithTx", mock.Anything).Return(s.itemRepo).Maybe()
	s.itemRepo.On("GetByID", s.ctx, s.itemID).Return(nil, gorm.ErrRecordNotFound).Once()

	movement, err := s.useCase.CreateStockMovement(s.ctx, s.itemID, s.warehouseID, nil, stock.MovementTypeIn, 1.0, "Reason")

	assert.ErrorContains(t, err, "item not found")
	assert.Nil(t, movement)
}

func TestCreateStockMovement_WarehouseNotFound(t *testing.T) {
	s := setupTestSuite()
	defer s.txManager.AssertExpectations(t)
	defer s.itemRepo.AssertExpectations(t)
	defer s.warehouseRepo.AssertExpectations(t)

	mockItem := &item.Item{ID: s.itemID, Name: "Test Item"}

	s.txManager.On("Transaction", mock.Anything, mock.AnythingOfType("func(*gorm.DB) error")).Return(errors.New("transaction failed: warehouse not found")).Once()
	s.itemRepo.On("WithTx", mock.Anything).Return(s.itemRepo).Maybe()
	s.itemRepo.On("GetByID", s.ctx, s.itemID).Return(mockItem, nil).Once()
	s.warehouseRepo.On("WithTx", mock.Anything).Return(s.warehouseRepo).Maybe()
	s.warehouseRepo.On("GetByID", s.ctx, s.warehouseID).Return(nil, gorm.ErrRecordNotFound).Once()

	movement, err := s.useCase.CreateStockMovement(s.ctx, s.itemID, s.warehouseID, nil, stock.MovementTypeIn, 1.0, "Reason")

	assert.ErrorContains(t, err, "warehouse not found")
	assert.Nil(t, movement)
}

func TestCreateStockMovement_WarehouseInactive(t *testing.T) {
	s := setupTestSuite()
	defer s.txManager.AssertExpectations(t)
	defer s.itemRepo.AssertExpectations(t)
	defer s.warehouseRepo.AssertExpectations(t)

	mockItem := &item.Item{ID: s.itemID, Name: "Test Item"}
	mockWarehouse := &stock.Warehouse{ID: s.warehouseID, Name: "Main Warehouse", IsActive: false}

	s.txManager.On("Transaction", mock.Anything, mock.AnythingOfType("func(*gorm.DB) error")).Return(errors.New("transaction failed: warehouse is inactive")).Once()
	s.itemRepo.On("WithTx", mock.Anything).Return(s.itemRepo).Maybe()
	s.itemRepo.On("GetByID", s.ctx, s.itemID).Return(mockItem, nil).Once()
	s.warehouseRepo.On("WithTx", mock.Anything).Return(s.warehouseRepo).Maybe()
	s.warehouseRepo.On("GetByID", s.ctx, s.warehouseID).Return(mockWarehouse, nil).Once()

	movement, err := s.useCase.CreateStockMovement(s.ctx, s.itemID, s.warehouseID, nil, stock.MovementTypeIn, 1.0, "Reason")

	assert.ErrorContains(t, err, "warehouse is inactive")
	assert.Nil(t, movement)
}

func TestCreateStockMovement_BinNotFound(t *testing.T) {
	s := setupTestSuite()
	defer s.txManager.AssertExpectations(t)
	defer s.itemRepo.AssertExpectations(t)
	defer s.warehouseRepo.AssertExpectations(t)
	defer s.binRepo.AssertExpectations(t)

	mockItem := &item.Item{ID: s.itemID, Name: "Test Item"}
	mockWarehouse := &stock.Warehouse{ID: s.warehouseID, Name: "Main Warehouse", IsActive: true}
	binID := &s.binID

	s.txManager.On("Transaction", mock.Anything, mock.AnythingOfType("func(*gorm.DB) error")).Return(errors.New("transaction failed: bin not found")).Once()
	s.itemRepo.On("WithTx", mock.Anything).Return(s.itemRepo).Maybe()
	s.itemRepo.On("GetByID", s.ctx, s.itemID).Return(mockItem, nil).Once()
	s.warehouseRepo.On("WithTx", mock.Anything).Return(s.warehouseRepo).Maybe()
	s.warehouseRepo.On("GetByID", s.ctx, s.warehouseID).Return(mockWarehouse, nil).Once()
	s.binRepo.On("WithTx", mock.Anything).Return(s.binRepo).Maybe()
	s.binRepo.On("GetByID", s.ctx, *binID).Return(nil, gorm.ErrRecordNotFound).Once()

	movement, err := s.useCase.CreateStockMovement(s.ctx, s.itemID, s.warehouseID, binID, stock.MovementTypeIn, 1.0, "Reason")

	assert.ErrorContains(t, err, "bin not found")
	assert.Nil(t, movement)
}

func TestCreateStockMovement_BinDoesNotBelongToWarehouse(t *testing.T) {
	s := setupTestSuite()
	defer s.txManager.AssertExpectations(t)
	defer s.itemRepo.AssertExpectations(t)
	defer s.warehouseRepo.AssertExpectations(t)
	defer s.binRepo.AssertExpectations(t)

	mockItem := &item.Item{ID: s.itemID, Name: "Test Item"}
	mockWarehouse := &stock.Warehouse{ID: s.warehouseID, Name: "Main Warehouse", IsActive: true}
	otherWarehouseID := uuid.New()
	mockBin := &stock.Bin{ID: s.binID, WarehouseID: otherWarehouseID, Name: "Bin 1", IsActive: true}
	binID := &s.binID

	s.txManager.On("Transaction", mock.Anything, mock.AnythingOfType("func(*gorm.DB) error")).Return(errors.New("transaction failed: bin does not belong to the specified warehouse")).Once()
	s.itemRepo.On("WithTx", mock.Anything).Return(s.itemRepo).Maybe()
	s.itemRepo.On("GetByID", s.ctx, s.itemID).Return(mockItem, nil).Once()
	s.warehouseRepo.On("WithTx", mock.Anything).Return(s.warehouseRepo).Maybe()
	s.warehouseRepo.On("GetByID", s.ctx, s.warehouseID).Return(mockWarehouse, nil).Once()
	s.binRepo.On("WithTx", mock.Anything).Return(s.binRepo).Maybe()
	s.binRepo.On("GetByID", s.ctx, *binID).Return(mockBin, nil).Once()

	movement, err := s.useCase.CreateStockMovement(s.ctx, s.itemID, s.warehouseID, binID, stock.MovementTypeIn, 1.0, "Reason")

	assert.ErrorContains(t, err, "bin does not belong to the specified warehouse")
	assert.Nil(t, movement)
}

func TestCreateStockMovement_BinInactive(t *testing.T) {
	s := setupTestSuite()
	defer s.txManager.AssertExpectations(t)
	defer s.itemRepo.AssertExpectations(t)
	defer s.warehouseRepo.AssertExpectations(t)
	defer s.binRepo.AssertExpectations(t)

	mockItem := &item.Item{ID: s.itemID, Name: "Test Item"}
	mockWarehouse := &stock.Warehouse{ID: s.warehouseID, Name: "Main Warehouse", IsActive: true}
	mockBin := &stock.Bin{ID: s.binID, WarehouseID: s.warehouseID, Name: "Bin 1", IsActive: false}
	binID := &s.binID

	s.txManager.On("Transaction", mock.Anything, mock.AnythingOfType("func(*gorm.DB) error")).Return(errors.New("transaction failed: bin is inactive")).Once()
	s.itemRepo.On("WithTx", mock.Anything).Return(s.itemRepo).Maybe()
	s.itemRepo.On("GetByID", s.ctx, s.itemID).Return(mockItem, nil).Once()
	s.warehouseRepo.On("WithTx", mock.Anything).Return(s.warehouseRepo).Maybe()
	s.warehouseRepo.On("GetByID", s.ctx, s.warehouseID).Return(mockWarehouse, nil).Once()
	s.binRepo.On("WithTx", mock.Anything).Return(s.binRepo).Maybe()
	s.binRepo.On("GetByID", s.ctx, *binID).Return(mockBin, nil).Once()

	movement, err := s.useCase.CreateStockMovement(s.ctx, s.itemID, s.warehouseID, binID, stock.MovementTypeIn, 1.0, "Reason")

	assert.ErrorContains(t, err, "bin is inactive")
	assert.Nil(t, movement)
}

func TestCreateStockMovement_StockMovementCreateError(t *testing.T) {
	s := setupTestSuite()
	defer s.txManager.AssertExpectations(t)
	defer s.itemRepo.AssertExpectations(t)
	defer s.warehouseRepo.AssertExpectations(t)
	defer s.stockRepo.AssertExpectations(t)
	defer s.stockMoveRepo.AssertExpectations(t)

	mockItem := &item.Item{ID: s.itemID, Name: "Test Item"}
	mockWarehouse := &stock.Warehouse{ID: s.warehouseID, Name: "Main Warehouse", IsActive: true}

	s.txManager.On("Transaction", mock.Anything, mock.AnythingOfType("func(*gorm.DB) error")).Return(errors.New("transaction failed: create movement error")).Once()
	s.itemRepo.On("WithTx", mock.Anything).Return(s.itemRepo).Maybe()
	s.itemRepo.On("GetByID", s.ctx, s.itemID).Return(mockItem, nil).Once()
	s.warehouseRepo.On("WithTx", mock.Anything).Return(s.warehouseRepo).Maybe()
	s.warehouseRepo.On("GetByID", s.ctx, s.warehouseID).Return(mockWarehouse, nil).Once()
	s.binRepo.On("WithTx", mock.Anything).Return(s.binRepo).Maybe() // Called because binID is nil
	s.stockRepo.On("WithTx", mock.Anything).Return(s.stockRepo).Maybe()
	s.stockRepo.On("GetStockForUpdate", s.ctx, s.itemID, s.warehouseID, mock.Anything).Return(nil, gorm.ErrRecordNotFound).Once()
	s.stockMoveRepo.On("WithTx", mock.Anything).Return(s.stockMoveRepo).Maybe()
	s.stockMoveRepo.On("Create", s.ctx, mock.AnythingOfType("*stock.StockMovement")).Return(errors.New("create movement error")).Once()

	movement, err := s.useCase.CreateStockMovement(s.ctx, s.itemID, s.warehouseID, nil, stock.MovementTypeIn, 10.0, "Initial Stock")

	assert.ErrorContains(t, err, "create movement error")
	assert.Nil(t, movement)
	s.stockRepo.AssertNotCalled(t, "UpsertStock", mock.Anything, mock.Anything)
	s.stockLedgerRepo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

func TestCreateStockMovement_UpsertStockError(t *testing.T) {
	s := setupTestSuite()
	defer s.txManager.AssertExpectations(t)
	defer s.itemRepo.AssertExpectations(t)
	defer s.warehouseRepo.AssertExpectations(t)
	defer s.stockRepo.AssertExpectations(t)
	defer s.stockMoveRepo.AssertExpectations(t)
	defer s.stockLedgerRepo.AssertExpectations(t)

	mockItem := &item.Item{ID: s.itemID, Name: "Test Item"}
	mockWarehouse := &stock.Warehouse{ID: s.warehouseID, Name: "Main Warehouse", IsActive: true}

	s.txManager.On("Transaction", mock.Anything, mock.AnythingOfType("func(*gorm.DB) error")).Return(errors.New("transaction failed: upsert stock error")).Once()
	s.itemRepo.On("WithTx", mock.Anything).Return(s.itemRepo).Maybe()
	s.itemRepo.On("GetByID", s.ctx, s.itemID).Return(mockItem, nil).Once()
	s.warehouseRepo.On("WithTx", mock.Anything).Return(s.warehouseRepo).Maybe()
	s.warehouseRepo.On("GetByID", s.ctx, s.warehouseID).Return(mockWarehouse, nil).Once()
	s.binRepo.On("WithTx", mock.Anything).Return(s.binRepo).Maybe()
	s.stockRepo.On("WithTx", mock.Anything).Return(s.stockRepo).Maybe()
	s.stockRepo.On("GetStockForUpdate", s.ctx, s.itemID, s.warehouseID, mock.Anything).Return(nil, gorm.ErrRecordNotFound).Once()
	s.stockMoveRepo.On("WithTx", mock.Anything).Return(s.stockMoveRepo).Maybe()
	s.stockMoveRepo.On("Create", s.ctx, mock.AnythingOfType("*stock.StockMovement")).Return(nil).Once()
	s.stockRepo.On("UpsertStock", s.ctx, mock.AnythingOfType("*stock.Stock")).Return(errors.New("upsert stock error")).Once()

	movement, err := s.useCase.CreateStockMovement(s.ctx, s.itemID, s.warehouseID, nil, stock.MovementTypeIn, 10.0, "Initial Stock")

	assert.ErrorContains(t, err, "upsert stock error")
	assert.Nil(t, movement)
	s.stockLedgerRepo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

func TestCreateStockMovement_CreateStockLedgerError(t *testing.T) {
	s := setupTestSuite()
	defer s.txManager.AssertExpectations(t)
	defer s.itemRepo.AssertExpectations(t)
	defer s.warehouseRepo.AssertExpectations(t)
	defer s.stockRepo.AssertExpectations(t)
	defer s.stockMoveRepo.AssertExpectations(t)
	defer s.stockLedgerRepo.AssertExpectations(t)

	mockItem := &item.Item{ID: s.itemID, Name: "Test Item"}
	mockWarehouse := &stock.Warehouse{ID: s.warehouseID, Name: "Main Warehouse", IsActive: true}

	s.txManager.On("Transaction", mock.Anything, mock.AnythingOfType("func(*gorm.DB) error")).Return(errors.New("transaction failed: create ledger error")).Once()
	s.itemRepo.On("WithTx", mock.Anything).Return(s.itemRepo).Maybe()
	s.itemRepo.On("GetByID", s.ctx, s.itemID).Return(mockItem, nil).Once()
	s.warehouseRepo.On("WithTx", mock.Anything).Return(s.warehouseRepo).Maybe()
	s.warehouseRepo.On("GetByID", s.ctx, s.warehouseID).Return(mockWarehouse, nil).Once()
	s.binRepo.On("WithTx", mock.Anything).Return(s.binRepo).Maybe()
	s.stockRepo.On("WithTx", mock.Anything).Return(s.stockRepo).Maybe()
	s.stockRepo.On("GetStockForUpdate", s.ctx, s.itemID, s.warehouseID, mock.Anything).Return(nil, gorm.ErrRecordNotFound).Once()
	s.stockMoveRepo.On("WithTx", mock.Anything).Return(s.stockMoveRepo).Maybe()
	s.stockMoveRepo.On("Create", s.ctx, mock.AnythingOfType("*stock.StockMovement")).Return(nil).Once()
	s.stockRepo.On("UpsertStock", s.ctx, mock.AnythingOfType("*stock.Stock")).Return(nil).Once()
	s.stockLedgerRepo.On("WithTx", mock.Anything).Return(s.stockLedgerRepo).Maybe()
	s.stockLedgerRepo.On("Create", s.ctx, mock.AnythingOfType("*stock.StockLedger")).Return(errors.New("create ledger error")).Once()

	movement, err := s.useCase.CreateStockMovement(s.ctx, s.itemID, s.warehouseID, nil, stock.MovementTypeIn, 10.0, "Initial Stock")

	assert.ErrorContains(t, err, "create ledger error")
	assert.Nil(t, movement)
}

// --- Tests for CreateWarehouse ---

func TestCreateWarehouse_HappyPath(t *testing.T) {
	s := setupTestSuite()
	defer s.txManager.AssertExpectations(t)
	defer s.warehouseRepo.AssertExpectations(t)

	warehouseName := "New Warehouse"

	s.txManager.On("Transaction", mock.Anything, mock.AnythingOfType("func(*gorm.DB) error")).Return(nil).Once()
	s.warehouseRepo.On("WithTx", mock.Anything).Return(s.warehouseRepo).Maybe()
	s.warehouseRepo.On("Create", s.ctx, mock.AnythingOfType("*stock.Warehouse")).Return(nil).Once()

	warehouse, err := s.useCase.CreateWarehouse(s.ctx, warehouseName)

	assert.NoError(t, err)
	assert.NotNil(t, warehouse)
	assert.Equal(t, warehouseName, warehouse.Name)
	assert.True(t, warehouse.IsActive)
	assert.Equal(t, s.userID, warehouse.CreatedBy)
	assert.Equal(t, s.userID, warehouse.UpdatedBy)
}

func TestCreateWarehouse_RepositoryError(t *testing.T) {
	s := setupTestSuite()
	defer s.txManager.AssertExpectations(t)
	defer s.warehouseRepo.AssertExpectations(t)

	warehouseName := "New Warehouse"
	repoErr := errors.New("repository error")

	s.txManager.On("Transaction", mock.Anything, mock.AnythingOfType("func(*gorm.DB) error")).Return(repoErr).Once()
	s.warehouseRepo.On("WithTx", mock.Anything).Return(s.warehouseRepo).Maybe()
	s.warehouseRepo.On("Create", s.ctx, mock.AnythingOfType("*stock.Warehouse")).Return(repoErr).Once()

	warehouse, err := s.useCase.CreateWarehouse(s.ctx, warehouseName)

	assert.ErrorIs(t, err, repoErr)
	assert.Nil(t, warehouse)
}

// --- Tests for ListWarehouses ---

func TestListWarehouses_HappyPath(t *testing.T) {
	s := setupTestSuite()
	defer s.warehouseRepo.AssertExpectations(t)

	mockWarehouses := []*stock.Warehouse{
		{ID: uuid.New(), Name: "W1"},
		{ID: uuid.New(), Name: "W2"},
	}

	s.warehouseRepo.On("List", s.ctx).Return(mockWarehouses, nil).Once()

	warehouses, err := s.useCase.ListWarehouses(s.ctx)

	assert.NoError(t, err)
	assert.Equal(t, mockWarehouses, warehouses)
}

func TestListWarehouses_RepositoryError(t *testing.T) {
	s := setupTestSuite()
	defer s.warehouseRepo.AssertExpectations(t)

	repoErr := errors.New("list error")

	s.warehouseRepo.On("List", s.ctx).Return(nil, repoErr).Once()

	warehouses, err := s.useCase.ListWarehouses(s.ctx)

	assert.ErrorIs(t, err, repoErr)
	assert.Nil(t, warehouses)
}

// --- Tests for GetWarehouseByID ---

func TestGetWarehouseByID_HappyPath(t *testing.T) {
	s := setupTestSuite()
	defer s.warehouseRepo.AssertExpectations(t)

	mockWarehouse := &stock.Warehouse{ID: s.warehouseID, Name: "Test Warehouse"}

	s.warehouseRepo.On("GetByID", s.ctx, s.warehouseID).Return(mockWarehouse, nil).Once()

	warehouse, err := s.useCase.GetWarehouseByID(s.ctx, s.warehouseID)

	assert.NoError(t, err)
	assert.Equal(t, mockWarehouse, warehouse)
}

func TestGetWarehouseByID_NotFound(t *testing.T) {
	s := setupTestSuite()
	defer s.warehouseRepo.AssertExpectations(t)

	s.warehouseRepo.On("GetByID", s.ctx, s.warehouseID).Return(nil, gorm.ErrRecordNotFound).Once()

	warehouse, err := s.useCase.GetWarehouseByID(s.ctx, s.warehouseID)

	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
	assert.Nil(t, warehouse)
}

func TestGetWarehouseByID_RepositoryError(t *testing.T) {
	s := setupTestSuite()
	defer s.warehouseRepo.AssertExpectations(t)

	repoErr := errors.New("db error")

	s.warehouseRepo.On("GetByID", s.ctx, s.warehouseID).Return(nil, repoErr).Once()

	warehouse, err := s.useCase.GetWarehouseByID(s.ctx, s.warehouseID)

	assert.ErrorIs(t, err, repoErr)
	assert.Nil(t, warehouse)
}

// --- Tests for CreateBin ---

func TestCreateBin_HappyPath(t *testing.T) {
	s := setupTestSuite()
	defer s.txManager.AssertExpectations(t)
	defer s.binRepo.AssertExpectations(t)

	binName := "New Bin"

	s.txManager.On("Transaction", mock.Anything, mock.AnythingOfType("func(*gorm.DB) error")).Return(nil).Once()
	s.binRepo.On("WithTx", mock.Anything).Return(s.binRepo).Maybe()
	s.binRepo.On("Create", s.ctx, mock.AnythingOfType("*stock.Bin")).Return(nil).Once()

	bin, err := s.useCase.CreateBin(s.ctx, binName, s.warehouseID)

	assert.NoError(t, err)
	assert.NotNil(t, bin)
	assert.Equal(t, binName, bin.Name)
	assert.Equal(t, s.warehouseID, bin.WarehouseID)
	assert.True(t, bin.IsActive)
	assert.Equal(t, s.userID, bin.CreatedBy)
	assert.Equal(t, s.userID, bin.UpdatedBy)
}

func TestCreateBin_RepositoryError(t *testing.T) {
	s := setupTestSuite()
	defer s.txManager.AssertExpectations(t)
	defer s.binRepo.AssertExpectations(t)

	binName := "New Bin"
	repoErr := errors.New("repository error")

	s.txManager.On("Transaction", mock.Anything, mock.AnythingOfType("func(*gorm.DB) error")).Return(repoErr).Once()
	s.binRepo.On("WithTx", mock.Anything).Return(s.binRepo).Maybe()
	s.binRepo.On("Create", s.ctx, mock.AnythingOfType("*stock.Bin")).Return(repoErr).Once()

	bin, err := s.useCase.CreateBin(s.ctx, binName, s.warehouseID)

	assert.ErrorIs(t, err, repoErr)
	assert.Nil(t, bin)
}

// --- Tests for ListBinsByWarehouse ---

func TestListBinsByWarehouse_HappyPath(t *testing.T) {
	s := setupTestSuite()
	defer s.binRepo.AssertExpectations(t)

	mockBins := []*stock.Bin{
		{ID: uuid.New(), Name: "B1", WarehouseID: s.warehouseID},
		{ID: uuid.New(), Name: "B2", WarehouseID: s.warehouseID},
	}

	s.binRepo.On("ListByWarehouse", s.ctx, s.warehouseID).Return(mockBins, nil).Once()

	bins, err := s.useCase.ListBinsByWarehouse(s.ctx, s.warehouseID)

	assert.NoError(t, err)
	assert.Equal(t, mockBins, bins)
}

func TestListBinsByWarehouse_RepositoryError(t *testing.T) {
	s := setupTestSuite()
	defer s.binRepo.AssertExpectations(t)

	repoErr := errors.New("list error")

	s.binRepo.On("ListByWarehouse", s.ctx, s.warehouseID).Return(nil, repoErr).Once()

	bins, err := s.useCase.ListBinsByWarehouse(s.ctx, s.warehouseID)

	assert.ErrorIs(t, err, repoErr)
	assert.Nil(t, bins)
}
