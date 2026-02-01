package bom

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"doligo/internal/domain"
	"doligo/internal/domain/bom"
	"doligo/internal/domain/item"
	"doligo/internal/domain/stock"
	"doligo/internal/infrastructure/db"
	infraRepo "doligo/internal/infrastructure/repository" // Alias to avoid conflict with domain.bom.Repository
	"gorm.io/gorm"
)

// BOMUsecase defines the business logic for Bill of Materials operations.
type BOMUsecase struct {
	bomRepo               bom.Repository
	itemRepo              item.Repository
	stockRepo             stock.StockRepository
	stockMovementRepo     stock.StockMovementRepository
	stockLedgerRepo       stock.StockLedgerRepository
	productionRecordRepo  bom.ProductionRecordRepository
	txManager             db.Transactioner
}

// NewBOMUsecase creates a new instance of BOMUsecase.
func NewBOMUsecase(
	br bom.Repository,
	ir item.Repository,
	sr stock.StockRepository,
	smr stock.StockMovementRepository,
	slr stock.StockLedgerRepository,
	prr bom.ProductionRecordRepository,
	tm db.Transactioner,
) *BOMUsecase {
	return &BOMUsecase{
		bomRepo:               br,
		itemRepo:              ir,
		stockRepo:             sr,
		stockMovementRepo:     smr,
		stockLedgerRepo:       slr,
		productionRecordRepo:  prr,
		txManager:             tm,
	}
}

// CreateBOM creates a new BillOfMaterials.
func (uc *BOMUsecase) CreateBOM(ctx context.Context, b *bom.BillOfMaterials) error {
	return uc.bomRepo.Create(ctx, b)
}

// GetBOMByID retrieves a BillOfMaterials by its ID.
func (uc *BOMUsecase) GetBOMByID(ctx context.Context, id uuid.UUID) (*bom.BillOfMaterials, error) {
	return uc.bomRepo.GetByID(ctx, id)
}

// GetBOMByProductID retrieves a BillOfMaterials by the product it produces.
func (uc *BOMUsecase) GetBOMByProductID(ctx context.Context, productID uuid.UUID) (*bom.BillOfMaterials, error) {
	return uc.bomRepo.GetByProductID(ctx, productID)
}

// ListBOMs retrieves all BillOfMaterials.
func (uc *BOMUsecase) ListBOMs(ctx context.Context) ([]*bom.BillOfMaterials, error) {
	return uc.bomRepo.List(ctx)
}

// UpdateBOM updates an existing BillOfMaterials.
func (uc *BOMUsecase) UpdateBOM(ctx context.Context, b *bom.BillOfMaterials) error {
	return uc.bomRepo.Update(ctx, b)
}

// DeleteBOM deletes a BillOfMaterials by its ID.
func (uc *BOMUsecase) DeleteBOM(ctx context.Context, id uuid.UUID) error {
	return uc.bomRepo.Delete(ctx, id)
}


// CalculatePredictiveCost calculates the estimated total production cost for a given BOM.
// It queries the current costs of components (items and services) but does not persist any data.
func (uc *BOMUsecase) CalculatePredictiveCost(ctx context.Context, bomID uuid.UUID) (float64, error) {
	b, err := uc.bomRepo.GetByID(ctx, bomID)
	if err != nil {
		return 0, fmt.Errorf("failed to get BOM by ID: %w", err)
	}
	if b == nil {
		return 0, fmt.Errorf("BOM with ID %s not found", bomID)
	}

	totalCost := 0.0

	for _, component := range b.Components {
		compItem, err := uc.itemRepo.GetByID(ctx, component.ComponentItemID)
		if err != nil {
			return 0, fmt.Errorf("failed to get component item %s: %w", component.ComponentItemID, err)
		}
		if compItem == nil {
			return 0, fmt.Errorf("component item %s not found", component.ComponentItemID)
		}

		// Use CostPrice for inputs, assuming services also have a 'CostPrice'
		// or a similar field that represents their cost to the company.
		componentCost := compItem.CostPrice * component.Quantity
		totalCost += componentCost
	}

	return totalCost, nil
}

// ProduceItem executes a production order for a given BOM and quantity.
// It handles pessimistic locking, stock validation, consumption of inputs,
// generation of the finished product, and recording of all movements
// within a single atomic transaction. It also calculates and persists the actual production cost.
func (uc *BOMUsecase) ProduceItem(
	ctx context.Context,
	bomID uuid.UUID,
	warehouseID uuid.UUID,
	productionQuantity float64,
	userID uuid.UUID,
) (uuid.UUID, float64, error) { // Added ProductionRecordID to return values
	var actualProductionCost float64
	var productionRecordID uuid.UUID

	err := uc.txManager.Transaction(ctx, func(tx *gorm.DB) error {
		// Create transactional repositories
		txItemRepo := infraRepo.NewGormItemRepository(tx)
		txStockRepo := infraRepo.NewGormStockRepository(tx)
		txStockMovementRepo := infraRepo.NewGormStockMovementRepository(tx)
		txStockLedgerRepo := infraRepo.NewGormStockLedgerRepository(tx)
		txBomRepo := infraRepo.NewGormBomRepository(tx)
		txProductionRecordRepo := infraRepo.NewGormProductionRecordRepository(tx)

		b, err := txBomRepo.GetByID(ctx, bomID)
		if err != nil {
			return fmt.Errorf("failed to get BOM by ID: %w", err)
		}
		if b == nil {
			return fmt.Errorf("BOM with ID %s not found", bomID)
		}

		// Ensure product exists
		productItem, err := txItemRepo.GetByID(ctx, b.ProductID)
		if err != nil {
			return fmt.Errorf("failed to get product item %s: %w", b.ProductID, err)
		}
		if productItem == nil {
			return fmt.Errorf("product item %s not found", b.ProductID)
		}

		// 1. Pessimistic Lock & Validate Stock for Inputs
		inputStocks := make(map[uuid.UUID]*stock.Stock) // Map component ID to its locked stock
		for _, component := range b.Components {
			compItem, err := txItemRepo.GetByID(ctx, component.ComponentItemID)
			if err != nil {
				return fmt.Errorf("failed to get component item %s: %w", component.ComponentItemID, err)
			}
			if compItem == nil {
				return fmt.Errorf("component item %s not found", component.ComponentItemID)
			}

			// Only lock and check storable items (raw materials)
			if compItem.Type == item.Storable {
				requiredQuantity := component.Quantity * productionQuantity
				lockedStock, err := txStockRepo.GetStockForUpdate(ctx, component.ComponentItemID, warehouseID, nil) // Assuming no bin for simplicity
				if err != nil {
					return fmt.Errorf("failed to get stock for update for item %s: %w", component.ComponentItemID, err)
				}
				if lockedStock == nil || lockedStock.Quantity < requiredQuantity {
					return fmt.Errorf("insufficient stock for component %s. Available: %f, Required: %f",
						compItem.Name, lockedStock.Quantity, requiredQuantity)
				}
				inputStocks[component.ComponentItemID] = lockedStock
			}
		}

		// 2. Calculate Actual Production Cost
		actualProductionCost = 0.0
		for _, component := range b.Components {
			compItem, err := txItemRepo.GetByID(ctx, component.ComponentItemID)
			if err != nil {
				return fmt.Errorf("failed to get component item %s for cost calculation: %w", component.ComponentItemID, err)
			}
			actualProductionCost += compItem.CostPrice * component.Quantity * productionQuantity
		}

		// 3. Consume Inputs (Storable items)
		for _, component := range b.Components {
			compItem, err := txItemRepo.GetByID(ctx, component.ComponentItemID)
			if err != nil {
				return fmt.Errorf("failed to get component item %s for consumption: %w", component.ComponentItemID, err)
			}

			if compItem.Type == item.Storable {
				requiredQuantity := component.Quantity * productionQuantity
				currentStock := inputStocks[component.ComponentItemID] // Use the locked stock

				// Record StockMovement (OUT)
				movement := &stock.StockMovement{
					ID:          uuid.New(),
					ItemID:      component.ComponentItemID,
					WarehouseID: warehouseID,
					BinID:       nil, // Assuming no bin
					Type:        stock.MovementTypeOut,
					Quantity:    requiredQuantity,
					Reason:      fmt.Sprintf("Consumption for BOM production of %s (BOM ID: %s)", productItem.Name, bomID),
					HappenedAt:  time.Now(),
					CreatedBy:   userID,
				}
				if err := txStockMovementRepo.Create(ctx, movement); err != nil {
					return fmt.Errorf("failed to create stock movement OUT for item %s: %w", component.ComponentItemID, err)
				}

				// Record StockLedger (OUT)
				ledgerEntry := &stock.StockLedger{
					ID:              uuid.New(),
					StockMovementID: movement.ID,
					ItemID:          component.ComponentItemID,
					WarehouseID:     warehouseID,
					BinID:           nil, // Assuming no bin
					MovementType:    stock.MovementTypeOut,
					QuantityChange:  requiredQuantity,
					QuantityBefore:  currentStock.Quantity,
					QuantityAfter:   currentStock.Quantity - requiredQuantity,
					Reason:          movement.Reason,
					HappenedAt:      movement.HappenedAt,
					RecordedAt:      time.Now(),
					RecordedBy:      userID,
				}
				if err := txStockLedgerRepo.Create(ctx, ledgerEntry); err != nil {
					return fmt.Errorf("failed to create stock ledger entry OUT for item %s: %w", component.ComponentItemID, err)
				}

				// Update Stock
				currentStock.Quantity -= requiredQuantity
				currentStock.UpdatedAt = time.Now()
				if err := txStockRepo.UpsertStock(ctx, currentStock); err != nil {
					return fmt.Errorf("failed to upsert stock for item %s: %w", component.ComponentItemID, err)
				}
			}
		}

		// 4. Generate Finished Product
		// Get current stock of the finished product for ledger purposes
		productCurrentStock, err := txStockRepo.GetStock(ctx, b.ProductID, warehouseID, nil)
		var productQuantityBefore float64
		if err != nil && err != gorm.ErrRecordNotFound {
			return fmt.Errorf("failed to get current stock for product %s: %w", b.ProductID, err)
		}
		if productCurrentStock != nil {
			productQuantityBefore = productCurrentStock.Quantity
		}

		// Record StockMovement (IN) for finished product
		productMovement := &stock.StockMovement{
			ID:          uuid.New(),
			ItemID:      b.ProductID,
			WarehouseID: warehouseID,
			BinID:       nil, // Assuming no bin
			Type:        stock.MovementTypeIn,
			Quantity:    productionQuantity,
			Reason:      fmt.Sprintf("Production of %s using BOM (BOM ID: %s)", productItem.Name, bomID),
			HappenedAt:  time.Now(),
			CreatedBy:   userID,
		}
		if err := txStockMovementRepo.Create(ctx, productMovement); err != nil {
			return fmt.Errorf("failed to create stock movement IN for product %s: %w", b.ProductID, err)
		}

		// Record StockLedger (IN) for finished product
		productLedgerEntry := &stock.StockLedger{
			ID:              uuid.New(),
			StockMovementID: productMovement.ID,
			ItemID:          b.ProductID,
			WarehouseID:     warehouseID,
			BinID:           nil, // Assuming no bin
			MovementType:    stock.MovementTypeIn,
			QuantityChange:  productionQuantity,
			QuantityBefore:  productQuantityBefore,
			QuantityAfter:   productQuantityBefore + productionQuantity,
			Reason:          productMovement.Reason,
			HappenedAt:      productMovement.HappenedAt,
			RecordedAt:      time.Now(),
			RecordedBy:      userID,
		}
		if err := txStockLedgerRepo.Create(ctx, productLedgerEntry); err != nil {
			return fmt.Errorf("failed to create stock ledger entry IN for product %s: %w", b.ProductID, err)
		}

		// Upsert Stock for finished product
		newProductStock := &stock.Stock{
			ItemID:      b.ProductID,
			WarehouseID: warehouseID,
			BinID:       nil, // Assuming no bin
			Quantity:    productQuantityBefore + productionQuantity,
			UpdatedAt:   time.Now(),
		}
		if err := txStockRepo.UpsertStock(ctx, newProductStock); err != nil {
			return fmt.Errorf("failed to upsert stock for product %s: %w", b.ProductID, err)
		}

		// 5. Persist Production Record
		productionRecord := &bom.ProductionRecord{
			ID:                   uuid.New(),
			BillOfMaterialsID:    bomID,
			ProducedProductID:    b.ProductID,
			ProductionQuantity:   productionQuantity,
			ActualProductionCost: actualProductionCost,
			WarehouseID:          warehouseID,
			ProducedAt:           time.Now(),
			CreatedBy:            userID,
		}
		productionRecord.SetCreatedBy(userID) // Ensure auditable interface is used
		if err := txProductionRecordRepo.Create(ctx, productionRecord); err != nil {
			return fmt.Errorf("failed to create production record: %w", err)
		}
		productionRecordID = productionRecord.ID // Capture the generated ID

		return nil
	})

	if err != nil {
		return uuid.Nil, 0, err
	}

	return productionRecordID, actualProductionCost, nil
}