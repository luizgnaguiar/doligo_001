package bom

import (
	"context"
	"errors"
	"fmt"
	"time"

	"doligo_001/internal/api/middleware"
	domainBom "doligo_001/internal/domain/bom"
	"doligo_001/internal/domain/item"
	"doligo_001/internal/domain/stock"
	"doligo_001/internal/infrastructure/db"
	"doligo_001/internal/usecase"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// BOMUsecase defines the interface for BOM related business logic.
type BOMUsecase interface {
	CreateBOM(ctx context.Context, bom *domainBom.BillOfMaterials) error
	GetBOMByID(ctx context.Context, id uuid.UUID) (*domainBom.BillOfMaterials, error)
	GetBOMByProductID(ctx context.Context, productID uuid.UUID) (*domainBom.BillOfMaterials, error)
	ListBOMs(ctx context.Context) ([]*domainBom.BillOfMaterials, error)
	UpdateBOM(ctx context.Context, bom *domainBom.BillOfMaterials) error
	DeleteBOM(ctx context.Context, id uuid.UUID) error
	CalculatePredictiveCost(ctx context.Context, bomID uuid.UUID) (float64, error)
	ProduceItem(ctx context.Context, bomID, warehouseID, userID uuid.UUID, productionQuantity float64) (uuid.UUID, float64, error)
}

type bomUsecase struct {
	txManager       db.Transactioner
	bomRepo         domainBom.Repository
	productionRepo  domainBom.ProductionRecordRepository
	stockRepo       stock.StockRepository
	stockMoveRepo   stock.StockMovementRepository
	stockLedgerRepo stock.StockLedgerRepository
	itemRepo        item.Repository
	auditService    usecase.AuditService
}

func NewBOMUsecase(
	txManager db.Transactioner,
	bomRepo domainBom.Repository,
	productionRepo domainBom.ProductionRecordRepository,
	stockRepo stock.StockRepository,
	stockMoveRepo stock.StockMovementRepository,
	stockLedgerRepo stock.StockLedgerRepository,
	itemRepo item.Repository,
	auditService usecase.AuditService,
) BOMUsecase {
	return &bomUsecase{
		txManager:       txManager,
		bomRepo:         bomRepo,
		productionRepo:  productionRepo,
		stockRepo:       stockRepo,
		stockMoveRepo:   stockMoveRepo,
		stockLedgerRepo: stockLedgerRepo,
		itemRepo:        itemRepo,
		auditService:    auditService,
	}
}

func (u *bomUsecase) CreateBOM(ctx context.Context, bom *domainBom.BillOfMaterials) error {
	return u.bomRepo.Create(ctx, bom)
}

func (u *bomUsecase) GetBOMByID(ctx context.Context, id uuid.UUID) (*domainBom.BillOfMaterials, error) {
	return u.bomRepo.GetByID(ctx, id)
}

func (u *bomUsecase) GetBOMByProductID(ctx context.Context, productID uuid.UUID) (*domainBom.BillOfMaterials, error) {
	return u.bomRepo.GetByProductID(ctx, productID)
}

func (u *bomUsecase) ListBOMs(ctx context.Context) ([]*domainBom.BillOfMaterials, error) {
	return u.bomRepo.List(ctx)
}

func (u *bomUsecase) UpdateBOM(ctx context.Context, bom *domainBom.BillOfMaterials) error {
	if bom.ID == uuid.Nil {
		return fmt.Errorf("BOM ID is required for update")
	}
	return u.bomRepo.Update(ctx, bom)
}

func (u *bomUsecase) DeleteBOM(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return fmt.Errorf("BOM ID is required for deletion")
	}
	return u.bomRepo.Delete(ctx, id)
}

func (u *bomUsecase) CalculatePredictiveCost(ctx context.Context, bomID uuid.UUID) (float64, error) {
	return 0, fmt.Errorf("CalculatePredictiveCost not implemented")
}

func (u *bomUsecase) ProduceItem(ctx context.Context, bomID, warehouseID, userID uuid.UUID, productionQuantity float64) (uuid.UUID, float64, error) {
	var productionRecordID uuid.UUID
	var actualProductionCost float64

	err := u.txManager.Transaction(ctx, func(tx *gorm.DB) error {
		// 1. Initialize transactional repositories
		txBomRepo := u.bomRepo.WithTx(tx)
		txStockRepo := u.stockRepo.WithTx(tx)
		txStockMoveRepo := u.stockMoveRepo.WithTx(tx)
		txStockLedgerRepo := u.stockLedgerRepo.WithTx(tx)
		txProductionRepo := u.productionRepo.WithTx(tx)

		// 2. Fetch BOM
		bom, err := txBomRepo.GetByID(ctx, bomID)
		if err != nil {
			return err
		}

		now := time.Now()

		// 3. Process Components (Stock OUT)
		for _, comp := range bom.Components {
			neededQty := comp.Quantity * productionQuantity

			// Pessimistic Lock on component stock
			// Assuming BinID is nil for simplification in ProduceItem as noted in SESSION_LOG
			s, err := txStockRepo.GetStockForUpdate(ctx, comp.ComponentItemID, warehouseID, nil)
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return fmt.Errorf("insufficient stock for component %s: not found", comp.ComponentItemID)
				}
				return err
			}

			if s.Quantity < neededQty {
				return fmt.Errorf("insufficient stock for component %s: have %f, need %f", comp.ComponentItemID, s.Quantity, neededQty)
			}

			oldQty := s.Quantity
			s.Quantity -= neededQty
			if err := txStockRepo.UpsertStock(ctx, s); err != nil {
				return err
			}

			// Create Stock Movement
			move := &stock.StockMovement{
				ID:          uuid.New(),
				ItemID:      comp.ComponentItemID,
				WarehouseID: warehouseID,
				BinID:       nil,
				Type:        stock.MovementTypeOut,
				Quantity:    neededQty,
				Reason:      fmt.Sprintf("Production of BOM %s", bomID),
				HappenedAt:  now,
				CreatedBy:   userID,
			}
			if err := txStockMoveRepo.Create(ctx, move); err != nil {
				return err
			}

			// Create Ledger Entry
			ledger := &stock.StockLedger{
				ID:              uuid.New(),
				StockMovementID: move.ID,
				ItemID:          comp.ComponentItemID,
				WarehouseID:     warehouseID,
				BinID:           nil,
				MovementType:    stock.MovementTypeOut,
				QuantityChange:  neededQty,
				QuantityBefore:  oldQty,
				QuantityAfter:   s.Quantity,
				Reason:          move.Reason,
				HappenedAt:      now,
				RecordedAt:      now,
				RecordedBy:      userID,
			}
			if err := txStockLedgerRepo.Create(ctx, ledger); err != nil {
				return err
			}
		}

		// 4. Process Product (Stock IN)
		// Pessimistic Lock on product stock
		prodStock, err := txStockRepo.GetStockForUpdate(ctx, bom.ProductID, warehouseID, nil)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		var oldProdQty float64 = 0
		if prodStock != nil {
			oldProdQty = prodStock.Quantity
		} else {
			prodStock = &stock.Stock{
				ItemID:      bom.ProductID,
				WarehouseID: warehouseID,
				BinID:       nil,
			}
		}

		prodStock.Quantity += productionQuantity
		prodStock.UpdatedAt = now
		if err := txStockRepo.UpsertStock(ctx, prodStock); err != nil {
			return err
		}

		// Create Stock Movement for Product
		prodMove := &stock.StockMovement{
			ID:          uuid.New(),
			ItemID:      bom.ProductID,
			WarehouseID: warehouseID,
			BinID:       nil,
			Type:        stock.MovementTypeIn,
			Quantity:    productionQuantity,
			Reason:      fmt.Sprintf("Finished production of BOM %s", bomID),
			HappenedAt:  now,
			CreatedBy:   userID,
		}
		if err := txStockMoveRepo.Create(ctx, prodMove); err != nil {
			return err
		}

		// Create Ledger Entry for Product
		prodLedger := &stock.StockLedger{
			ID:              uuid.New(),
			StockMovementID: prodMove.ID,
			ItemID:          bom.ProductID,
			WarehouseID:     warehouseID,
			BinID:           nil,
			MovementType:    stock.MovementTypeIn,
			QuantityChange:  productionQuantity,
			QuantityBefore:  oldProdQty,
			QuantityAfter:   prodStock.Quantity,
			Reason:          prodMove.Reason,
			HappenedAt:      now,
			RecordedAt:      now,
			RecordedBy:      userID,
		}
		if err := txStockLedgerRepo.Create(ctx, prodLedger); err != nil {
			return err
		}

		// 5. Create Production Record
		record := &domainBom.ProductionRecord{
			ID:                   uuid.New(),
			BillOfMaterialsID:    bomID,
			ProducedProductID:    bom.ProductID,
			ProductionQuantity:   productionQuantity,
			ActualProductionCost: 0, // Simplified for now
			WarehouseID:          warehouseID,
			ProducedAt:           now,
			CreatedBy:            userID,
		}
		if err := txProductionRepo.Create(ctx, record); err != nil {
			return err
		}

		productionRecordID = record.ID
		actualProductionCost = record.ActualProductionCost
		return nil
	})

	if err == nil {
		corrID, _ := middleware.FromContext(ctx)
		u.auditService.Log(ctx, userID, "production", productionRecordID.String(), "CREATE",
			nil, map[string]interface{}{"bom_id": bomID, "quantity": productionQuantity},
			corrID)
	}

	return productionRecordID, actualProductionCost, err
}
