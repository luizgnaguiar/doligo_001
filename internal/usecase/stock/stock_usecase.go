// Package stock contains the use cases for stock management.
package stock

import (
	"context"
	"errors"
	"time"

	"doligo_001/internal/domain"
	"doligo_001/internal/domain/item"
	"doligo_001/internal/domain/stock"
	"doligo_001/internal/infrastructure/db"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrInsufficientStock = errors.New("insufficient stock for movement")
	ErrBinRequired       = errors.New("bin_id is required for all stock movements")
)

// UseCase defines the interface for stock management use cases.
type UseCase interface {
	CreateStockMovement(ctx context.Context, itemID, warehouseID, binID uuid.UUID, movementType stock.MovementType, quantity float64, reason string) (*stock.StockMovement, error)
	CreateWarehouse(ctx context.Context, name string) (*stock.Warehouse, error)
	ListWarehouses(ctx context.Context) ([]*stock.Warehouse, error)
	GetWarehouseByID(ctx context.Context, id uuid.UUID) (*stock.Warehouse, error)
	CreateBin(ctx context.Context, name string, warehouseID uuid.UUID) (*stock.Bin, error)
	ListBinsByWarehouse(ctx context.Context, warehouseID uuid.UUID) ([]*stock.Bin, error)
}

// stockUseCase implements the UseCase interface.
type stockUseCase struct {
	txManager    db.Transactioner
	stockRepo    stock.StockRepository
	stockMoveRepo stock.StockMovementRepository
	stockLedgerRepo stock.StockLedgerRepository
	warehouseRepo stock.WarehouseRepository
	binRepo      stock.BinRepository
	itemRepo     item.Repository
}

// NewUseCase creates a new stockUseCase.
func NewUseCase(
	txManager db.Transactioner,
	stockRepo stock.StockRepository,
	stockMoveRepo stock.StockMovementRepository,
	stockLedgerRepo stock.StockLedgerRepository,
	warehouseRepo stock.WarehouseRepository,
	binRepo stock.BinRepository,
	itemRepo item.Repository,
) UseCase {
	return &stockUseCase{
		txManager:    txManager,
		stockRepo:    stockRepo,
		stockMoveRepo: stockMoveRepo,
		stockLedgerRepo: stockLedgerRepo,
		warehouseRepo: warehouseRepo,
		binRepo:      binRepo,
		itemRepo:     itemRepo,
	}
}

// CreateStockMovement handles the logic for creating a stock movement atomically.
func (uc *stockUseCase) CreateStockMovement(ctx context.Context, itemID, warehouseID, binID uuid.UUID, movementType stock.MovementType, quantity float64, reason string) (*stock.StockMovement, error) {
	var createdMovement *stock.StockMovement

	// Explicitly check for zero UUID as a safeguard, though validation should catch it.
	if binID == uuid.Nil {
		return nil, ErrBinRequired
	}

	err := uc.txManager.Transaction(ctx, func(tx *gorm.DB) error {
		// Create transactional repositories
		txStockRepo := uc.stockRepo.WithTx(tx)
		txMovementRepo := uc.stockMoveRepo.WithTx(tx)
		txLedgerRepo := uc.stockLedgerRepo.WithTx(tx)
		txItemRepo := uc.itemRepo.WithTx(tx)
		txWarehouseRepo := uc.warehouseRepo.WithTx(tx)
		txBinRepo := uc.binRepo.WithTx(tx)

		// Validate ItemID
		_, err := txItemRepo.GetByID(ctx, itemID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("item not found")
			}
			return err
		}

		// Validate WarehouseID
		warehouse, err := txWarehouseRepo.GetByID(ctx, warehouseID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("warehouse not found")
			}
			return err
		}
		if !warehouse.IsActive {
			return errors.New("warehouse is inactive")
		}

		// Validate BinID
		bin, err := txBinRepo.GetByID(ctx, binID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("bin not found")
			}
			return err
		}
		if bin.WarehouseID != warehouseID {
			return errors.New("bin does not belong to the specified warehouse")
		}
		if !bin.IsActive {
			return errors.New("bin is inactive")
		}

		// 1. Get current stock with pessimistic lock
		currentStock, err := txStockRepo.GetStockForUpdate(ctx, itemID, warehouseID, &binID)
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
			BinID:       &binID,
			Type:        movementType,
			Quantity:    quantity,
			Reason:      reason,
			HappenedAt:  time.Now(),
		}
		movement.SetCreatedBy(userID)

		if err := txMovementRepo.Create(ctx, movement); err != nil {
			return err
		}

		createdMovement = movement

		// 4. Upsert Stock
		stockToUpdate := &stock.Stock{
			ItemID:      itemID,
			WarehouseID: warehouseID,
			BinID:       &binID,
			Quantity:    quantityAfter,
			UpdatedAt:   time.Now(),
		}
		if err := txStockRepo.UpsertStock(ctx, stockToUpdate); err != nil {
			return err
		}

		// 5. Create StockLedger entry
		ledgerEntry := &stock.StockLedger{
			ID:              uuid.New(),
			StockMovementID: movement.ID,
			ItemID:          itemID,
			WarehouseID:     warehouseID,
			BinID:           &binID,
			MovementType:    movementType,
			QuantityChange:  quantity,
			QuantityBefore:  quantityBefore,
			QuantityAfter:   quantityAfter,
			Reason:          reason,
			HappenedAt:      movement.HappenedAt,
			RecordedAt:      time.Now(),
			RecordedBy:      userID,
		}

		return txLedgerRepo.Create(ctx, ledgerEntry)
	})

	return createdMovement, err
}

func (uc *stockUseCase) CreateWarehouse(ctx context.Context, name string) (*stock.Warehouse, error) {
	var createdWarehouse *stock.Warehouse
	err := uc.txManager.Transaction(ctx, func(tx *gorm.DB) error {
		repo := uc.warehouseRepo.WithTx(tx)
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
	return uc.warehouseRepo.List(ctx)
}

func (uc *stockUseCase) GetWarehouseByID(ctx context.Context, id uuid.UUID) (*stock.Warehouse, error) {
	return uc.warehouseRepo.GetByID(ctx, id)
}

func (uc *stockUseCase) CreateBin(ctx context.Context, name string, warehouseID uuid.UUID) (*stock.Bin, error) {
	var createdBin *stock.Bin
	err := uc.txManager.Transaction(ctx, func(tx *gorm.DB) error {
		repo := uc.binRepo.WithTx(tx)
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
	return uc.binRepo.ListByWarehouse(ctx, warehouseID)
}
