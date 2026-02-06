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
	if args.Error(0) != nil {
		return args.Error(0)
	}
	return fc(nil) 
}

// MockItemRepository
type MockItemRepository struct {
	mock.Mock
}

func (m *MockItemRepository) WithTx(tx *gorm.DB) item.Repository {
	m.Called(tx)
	return m
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
	m.Called(tx)
	return m
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
	m.Called(tx)
	return m
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
	m.Called(tx)
	return m
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

func (m *MockStockRepository) GetTotalQuantity(ctx context.Context, itemID uuid.UUID) (float64, error) {
	args := m.Called(ctx, itemID)
	return args.Get(0).(float64), args.Error(1)
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
	m.Called(tx)
	return m
}
func (m *MockStockMovementRepository) Create(ctx context.Context, movement *stock.StockMovement) error {
	args := m.Called(ctx, movement)
	return args.Error(0)
}

func (m *MockStockMovementRepository) GetByID(ctx context.Context, id uuid.UUID) (*stock.StockMovement, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*stock.StockMovement), args.Error(1)
}

// MockStockLedgerRepository
type MockStockLedgerRepository struct {
	mock.Mock
}

func (m *MockStockLedgerRepository) WithTx(tx *gorm.DB) stock.StockLedgerRepository {
	m.Called(tx)
	return m
}
func (m *MockStockLedgerRepository) Create(ctx context.Context, entry *stock.StockLedger) error {
	args := m.Called(ctx, entry)
	return args.Error(0)
}

// MockAuditService
type MockAuditService struct {
	mock.Mock
}

func (m *MockAuditService) Log(ctx context.Context, userID uuid.UUID, resourceName, resourceID, action string, oldValues, newValues interface{}, correlationID string) {
	m.Called(ctx, userID, resourceName, resourceID, action, oldValues, newValues, correlationID)
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
	auditService    *MockAuditService
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
		auditService:    new(MockAuditService),
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
		s.auditService,
	)

	s.ctx = context.WithValue(context.Background(), domain.UserIDKey, s.userID)

	// Default Mock setup for WithTx (used in almost all tests)
	s.itemRepo.On("WithTx", mock.Anything).Return(s.itemRepo).Maybe()
	s.warehouseRepo.On("WithTx", mock.Anything).Return(s.warehouseRepo).Maybe()
	s.binRepo.On("WithTx", mock.Anything).Return(s.binRepo).Maybe()
	s.stockRepo.On("WithTx", mock.Anything).Return(s.stockRepo).Maybe()
	s.stockMoveRepo.On("WithTx", mock.Anything).Return(s.stockMoveRepo).Maybe()
	s.stockLedgerRepo.On("WithTx", mock.Anything).Return(s.stockLedgerRepo).Maybe()

	// Default mock for Bin validation
	mockBin := &stock.Bin{ID: s.binID, WarehouseID: s.warehouseID, IsActive: true}
	s.binRepo.On("GetByID", mock.Anything, s.binID).Return(mockBin, nil).Maybe()
	
	// Default mock for Audit
	s.auditService.On("Log", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe()

	return s
}

func TestCreateStockMovement_In_HappyPath_NoExistingStock(t *testing.T) {
	s := setupTestSuite()
	mockItem := &item.Item{ID: s.itemID, Name: "Test Item", AverageCost: 100.0}
	mockWarehouse := &stock.Warehouse{ID: s.warehouseID, Name: "Main Warehouse", IsActive: true}

	s.txManager.On("Transaction", mock.Anything, mock.Anything).Return(nil).Once()
	s.itemRepo.On("GetByID", mock.Anything, s.itemID).Return(mockItem, nil).Once()
	s.stockRepo.On("GetTotalQuantity", mock.Anything, s.itemID).Return(0.0, nil).Once()
	s.itemRepo.On("Update", mock.Anything, mock.AnythingOfType("*item.Item")).Return(nil).Once()
	s.warehouseRepo.On("GetByID", mock.Anything, s.warehouseID).Return(mockWarehouse, nil).Once()
	s.stockRepo.On("GetStockForUpdate", mock.Anything, s.itemID, s.warehouseID, &s.binID).Return(nil, gorm.ErrRecordNotFound).Once()
	s.stockMoveRepo.On("Create", mock.Anything, mock.AnythingOfType("*stock.StockMovement")).Return(nil).Once()
	s.stockRepo.On("UpsertStock", mock.Anything, mock.AnythingOfType("*stock.Stock")).Return(nil).Once()
	s.stockLedgerRepo.On("Create", mock.Anything, mock.AnythingOfType("*stock.StockLedger")).Return(nil).Once()

	movement, err := s.useCase.CreateStockMovement(s.ctx, s.itemID, s.warehouseID, s.binID, stock.MovementTypeIn, 10.0, 120.0, "Initial Stock")

	assert.NoError(t, err)
	assert.NotNil(t, movement)
	assert.Equal(t, s.itemID, movement.ItemID)
	assert.Equal(t, s.binID, *movement.BinID)
}

func TestCreateStockMovement_Out_InsufficientStock(t *testing.T) {
	s := setupTestSuite()
	mockItem := &item.Item{ID: s.itemID, Name: "Test Item"}
	mockWarehouse := &stock.Warehouse{ID: s.warehouseID, Name: "Main Warehouse", IsActive: true}
	existingStock := &stock.Stock{ItemID: s.itemID, WarehouseID: s.warehouseID, Quantity: 5.0}

	s.txManager.On("Transaction", mock.Anything, mock.Anything).Return(errors.New("insufficient stock")).Once()
	s.itemRepo.On("GetByID", mock.Anything, s.itemID).Return(mockItem, nil).Once()
	s.warehouseRepo.On("GetByID", mock.Anything, s.warehouseID).Return(mockWarehouse, nil).Once()
	s.stockRepo.On("GetStockForUpdate", mock.Anything, s.itemID, s.warehouseID, &s.binID).Return(existingStock, nil).Once()

	movement, err := s.useCase.CreateStockMovement(s.ctx, s.itemID, s.warehouseID, s.binID, stock.MovementTypeOut, 10.0, 0.0, "Selling Item")

	assert.Error(t, err)
	assert.Nil(t, movement)
}

func TestCreateWarehouse_HappyPath(t *testing.T) {
	s := setupTestSuite()
	s.txManager.On("Transaction", mock.Anything, mock.Anything).Return(nil).Once()
	s.warehouseRepo.On("Create", mock.Anything, mock.AnythingOfType("*stock.Warehouse")).Return(nil).Once()

	warehouse, err := s.useCase.CreateWarehouse(s.ctx, "New Warehouse")

	assert.NoError(t, err)
	assert.NotNil(t, warehouse)
}

func TestCreateBin_HappyPath(t *testing.T) {
	s := setupTestSuite()
	s.txManager.On("Transaction", mock.Anything, mock.Anything).Return(nil).Once()
	s.binRepo.On("Create", mock.Anything, mock.AnythingOfType("*stock.Bin")).Return(nil).Once()

	bin, err := s.useCase.CreateBin(s.ctx, "New Bin", s.warehouseID)

	assert.NoError(t, err)
	assert.NotNil(t, bin)
}

func TestReverseStockMovement_ReversingOut_UpdatesCMP(t *testing.T) {
	s := setupTestSuite()
	origMoveID := uuid.New()
	
	// Original Movement was OUT (Sale)
	origMove := &stock.StockMovement{
		ID:          origMoveID,
		ItemID:      s.itemID,
		WarehouseID: s.warehouseID,
		BinID:       &s.binID,
		Type:        stock.MovementTypeOut,
		Quantity:    5.0,
	}

	mockItem := &item.Item{ID: s.itemID, Name: "Test Item", AverageCost: 15.0}
	currentStock := &stock.Stock{ItemID: s.itemID, WarehouseID: s.warehouseID, BinID: &s.binID, Quantity: 10.0}

	s.txManager.On("Transaction", mock.Anything, mock.Anything).Return(nil).Once()
	
	// 1. Get Original Movement
	s.stockMoveRepo.On("GetByID", mock.Anything, origMoveID).Return(origMove, nil).Once()
	
	// 2. CMP Logic (Reversing OUT -> IN)
	s.itemRepo.On("GetByID", mock.Anything, s.itemID).Return(mockItem, nil).Once()
	s.stockRepo.On("GetTotalQuantity", mock.Anything, s.itemID).Return(10.0, nil).Once()
	s.itemRepo.On("Update", mock.Anything, mock.MatchedBy(func(i *item.Item) bool {
		return i.AverageCost == 15.0 // (10*15 + 5*15)/15 = 15
	})).Return(nil).Once()

	// 3. Get Stock For Update
	s.stockRepo.On("GetStockForUpdate", mock.Anything, s.itemID, s.warehouseID, &s.binID).Return(currentStock, nil).Once()

	// 4. Create Reversal Movement
	s.stockMoveRepo.On("Create", mock.Anything, mock.AnythingOfType("*stock.StockMovement")).Return(nil).Once()

	// 5. Upsert Stock (10 + 5 = 15)
	s.stockRepo.On("UpsertStock", mock.Anything, mock.MatchedBy(func(st *stock.Stock) bool {
		return st.Quantity == 15.0
	})).Return(nil).Once()

	// 6. Ledger
	s.stockLedgerRepo.On("Create", mock.Anything, mock.AnythingOfType("*stock.StockLedger")).Return(nil).Once()

	reversedMovement, err := s.useCase.ReverseStockMovement(s.ctx, origMoveID, "Customer Return")

	assert.NoError(t, err)
	assert.NotNil(t, reversedMovement)
	assert.Equal(t, stock.MovementTypeIn, reversedMovement.Type)
	assert.Equal(t, 5.0, reversedMovement.Quantity)
}
