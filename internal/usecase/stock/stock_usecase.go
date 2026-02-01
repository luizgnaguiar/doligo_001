// Package stock contains the use cases for stock management.
package stock

import (
	"context"
	"errors"
	"time"

	"doligo_001/internal/domain"
	"doligo_001/internal/domain/stock"
	"doligo_001/internal/infrastructure/db"
	"doligo_001/internal/infrastructure/repository"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrInsufficientStock = errors.New("insufficient stock for movement")
)

// UseCase defines the interface for stock management use cases.
type UseCase interface {
	CreateStockMovement(ctx context.Context, itemID, warehouseID uuid.UUID, binID *uuid.UUID, movementType stock.MovementType, quantity float64, reason string) (*stock.StockMovement, error)
	CreateWarehouse(ctx context.Context, name string) (*stock.Warehouse, error)
	ListWarehouses(ctx context.Context) ([]*stock.Warehouse, error)
	GetWarehouseByID(ctx context.Context, id uuid.UUID) (*stock.Warehouse, error)
	CreateBin(ctx context.Context, name string, warehouseID uuid.UUID) (*stock.Bin, error)
	ListBinsByWarehouse(ctx context.Context, warehouseID uuid.UUID) ([]*stock.Bin, error)
}

// stockUseCase implements the UseCase interface.
type stockUseCase struct {
	db        *gorm.DB
	txManager db.Transactioner
}

// NewUseCase creates a new stockUseCase.
func NewUseCase(db *gorm.DB, txManager db.Transactioner) UseCase {
	return &stockUseCase{db: db, txManager: txManager}
}

// CreateStockMovement handles the logic for creating a stock movement atomically.
func (uc *stockUseCase) CreateStockMovement(ctx context.Context, itemID, warehouseID uuid.UUID, binID *uuid.UUID, movementType stock.MovementType, quantity float64, reason string) (*stock.StockMovement, error) {
	var createdMovement *stock.StockMovement

	err := uc.txManager.Transaction(ctx, func(tx *gorm.DB) error {
		stockRepo := repository.NewGormStockRepository(tx)
		movementRepo := repository.NewGormStockMovementRepository(tx)
		ledgerRepo := repository.NewGormStockLedgerRepository(tx)

		// 1. Get current stock with pessimistic lock
		currentStock, err := stockRepo.GetStockForUpdate(ctx, itemID, warehouseID, binID)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		quantityBefore := 0.0
		if currentStock != nil {
			quantityBefore = currentStock.Quantity
		}

		// 2. Validate movement
		var quantityAfter float64
		if movementType == stock.MovementTypeOut {
			if quantityBefore < quantity {
				return ErrInsufficientStock
			}
			quantityAfter = quantityBefore - quantity
		} else {
			quantityAfter = quantityBefore + quantity
		}

		// 3. Create StockMovement
		userID, _ := domain.UserIDFromContext(ctx)
		movement := &stock.StockMovement{
			ID:          uuid.New(),
			ItemID:      itemID,
			WarehouseID: warehouseID,
			BinID:       binID,
			Type:        movementType,
			Quantity:    quantity,
			Reason:      reason,
			HappenedAt:  time.Now(),
			CreatedBy:   userID,
		}

		if err := movementRepo.Create(ctx, movement); err != nil {
			return err
		}

		createdMovement = movement

		// 4. Upsert Stock
		stockToUpdate := &stock.Stock{
			ItemID:      itemID,
			WarehouseID: warehouseID,
			BinID:       binID,
			Quantity:    quantityAfter,
			UpdatedAt:   time.Now(),
		}
		if err := stockRepo.UpsertStock(ctx, stockToUpdate); err != nil {
			return err
		}

		// 5. Create StockLedger entry
		ledgerEntry := &stock.StockLedger{
			ID:              uuid.New(),
			StockMovementID: movement.ID,
			ItemID:          itemID,
			WarehouseID:     warehouseID,
			BinID:           binID,
			MovementType:    movementType,
			QuantityChange:  quantity,
			QuantityBefore:  quantityBefore,
			QuantityAfter:   quantityAfter,
			Reason:          reason,
			HappenedAt:      movement.HappenedAt,
			RecordedAt:      time.Now(),
			RecordedBy:      userID,
		}

		return ledgerRepo.Create(ctx, ledgerEntry)
	})

	return createdMovement, err
}

func (uc *stockUseCase) CreateWarehouse(ctx context.Context, name string) (*stock.Warehouse, error) {
	var createdWarehouse *stock.Warehouse
	err := uc.txManager.Transaction(ctx, func(tx *gorm.DB) error {
		repo := repository.NewGormWarehouseRepository(tx)
		userID, _ := domain.UserIDFromContext(ctx)

		warehouse := &stock.Warehouse{
			ID:       uuid.New(),
			Name:     name,
			IsActive: true,
		}
		warehouse.SetCreatedBy(userID)
		warehouse.SetUpdatedBy(userID)

		if err := repo.Create(ctx, warehouse); err != nil {
			return err
		}
		createdWarehouse = warehouse
		return nil
	})
	return createdWarehouse, err
}

func (uc *stockUseCase) ListWarehouses(ctx context.Context) ([]*stock.Warehouse, error) {
	repo := repository.NewGormWarehouseRepository(uc.db)
	return repo.List(ctx)
}

func (uc *stockUseCase) GetWarehouseByID(ctx context.Context, id uuid.UUID) (*stock.Warehouse, error) {
	repo := repository.NewGormWarehouseRepository(uc.db)
	return repo.GetByID(ctx, id)
}

func (uc *stockUseCase) CreateBin(ctx context.Context, name string, warehouseID uuid.UUID) (*stock.Bin, error) {
	var createdBin *stock.Bin
	err := uc.txManager.Transaction(ctx, func(tx *gorm.DB) error {
		repo := repository.NewGormBinRepository(tx)
		userID, _ := domain.UserIDFromContext(ctx)

		bin := &stock.Bin{
			ID:          uuid.New(),
			Name:        name,
			WarehouseID: warehouseID,
			IsActive:    true,
		}
		bin.SetCreatedBy(userID)
		bin.SetUpdatedBy(userID)

		if err := repo.Create(ctx, bin); err != nil {
			return err
		}
		createdBin = bin
		return nil
	})
	return createdBin, err
}

func (uc *stockUseCase) ListBinsByWarehouse(ctx context.Context, warehouseID uuid.UUID) ([]*stock.Bin, error) {
	repo := repository.NewGormBinRepository(uc.db)
	return repo.ListByWarehouse(ctx, warehouseID)
}
