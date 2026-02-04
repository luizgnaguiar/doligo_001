package repository_test

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"doligo_001/internal/domain/bom"
	"doligo_001/internal/domain/identity"
	"doligo_001/internal/domain/item"
	"doligo_001/internal/domain/stock"
	"doligo_001/internal/infrastructure/db"
	"doligo_001/internal/infrastructure/repository"
	"doligo_001/internal/usecase"
	bom_uc "doligo_001/internal/usecase/bom"
	stock_uc "doligo_001/internal/usecase/stock"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHighConcurrencyStockAndBOM(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping stress test in short mode")
	}

	// Reusing testDB and testTxManager from TestMain in stock_repository_ct02_test.go
	if testDB == nil {
		t.Skip("testDB not initialized, skipping integration test")
	}

	ctx := context.Background()
	gormDB := testDB
	txManager := testTxManager

	// Repositories
	itemRepo := repository.NewGormItemRepository(gormDB)
	bomRepo := repository.NewGormBomRepository(gormDB, txManager)
	productionRepo := repository.NewGormProductionRecordRepository(gormDB)
	stockRepo := repository.NewGormStockRepository(gormDB)
	stockMoveRepo := repository.NewGormStockMovementRepository(gormDB)
	stockLedgerRepo := repository.NewGormStockLedgerRepository(gormDB)
	warehouseRepo := repository.NewGormWarehouseRepository(gormDB)
	binRepo := repository.NewGormBinRepository(gormDB)
	userRepo := repository.NewGormUserRepository(gormDB)
	auditRepo := db.NewGormAuditRepository(gormDB)

	// Services
	auditService := usecase.NewAuditService(auditRepo)
	stockUsecase := stock_uc.NewUseCase(txManager, stockRepo, stockMoveRepo, stockLedgerRepo, warehouseRepo, binRepo, itemRepo, auditService)
	bomUsecase := bom_uc.NewBOMUsecase(txManager, bomRepo, productionRepo, stockRepo, stockMoveRepo, stockLedgerRepo, itemRepo, auditService)

	// 0. Setup Test Data
	testUser := &identity.User{
		ID:        uuid.New(),
		FirstName: "Stress",
		LastName:  "Tester",
		Email:     fmt.Sprintf("stress_%d@example.com", time.Now().UnixNano()),
		Password:  "hash",
		IsActive:  true,
	}
	require.NoError(t, userRepo.Create(ctx, testUser))

	// Items: Component and Product
	compItem := &item.Item{ID: uuid.New(), Name: "Component Item", Type: item.Storable}
	compItem.SetCreatedBy(testUser.ID)
	require.NoError(t, itemRepo.Create(ctx, compItem))

	productItem := &item.Item{ID: uuid.New(), Name: "Product Item", Type: item.Storable}
	productItem.SetCreatedBy(testUser.ID)
	require.NoError(t, itemRepo.Create(ctx, productItem))

	warehouse := &stock.Warehouse{ID: uuid.New(), Name: "Main Warehouse", IsActive: true}
	warehouse.SetCreatedBy(testUser.ID)
	require.NoError(t, warehouseRepo.Create(ctx, warehouse))

	bin := &stock.Bin{ID: uuid.New(), Name: "Bin 1", WarehouseID: warehouse.ID, IsActive: true}
	bin.SetCreatedBy(testUser.ID)
	require.NoError(t, binRepo.Create(ctx, bin))

	// BOM: 1 Component = 1 Product
	testBOM := &bom.BillOfMaterials{
		ID:        uuid.New(),
		ProductID: productItem.ID,
		Name:      "Simple BOM",
		IsActive:  true,
		Components: []bom.BillOfMaterialsComponent{
			{
				ID:              uuid.New(),
				ComponentItemID: compItem.ID,
				Quantity:        1.0,
				IsActive:        true,
			},
		},
	}
	testBOM.SetCreatedBy(testUser.ID)
	require.NoError(t, bomRepo.Create(ctx, testBOM))

	// Initial Stock for Component
	initialCompQty := 1000.0
	require.NoError(t, stockRepo.UpsertStock(ctx, &stock.Stock{
		ItemID:      compItem.ID,
		WarehouseID: warehouse.ID,
		BinID:       nil, // ProduceItem assumes nil bin
		Quantity:    initialCompQty,
	}))

	// 1. Stress Test Execution
	numStockMovements := 50
	numProductions := 50
	totalOperations := numStockMovements + numProductions

	var wg sync.WaitGroup
	wg.Add(totalOperations)

	startSignal := make(chan struct{})
	
	errorsChan := make(chan error, totalOperations)

	// User Context
	userCtx := context.WithValue(ctx, "user_id", testUser.ID)

	// Goroutines for Stock Movements (OUT)
	for i := 0; i < numStockMovements; i++ {
		go func(id int) {
			defer wg.Done()
			<-startSignal
			_, err := stockUsecase.CreateStockMovement(userCtx, compItem.ID, warehouse.ID, bin.ID, stock.MovementTypeOut, 1.0, "Stress Test Move")
			if err != nil {
				errorsChan <- fmt.Errorf("StockMovement %d failed: %w", id, err)
			}
		}(i)
	}

	// Goroutines for BOM Production
	for i := 0; i < numProductions; i++ {
		go func(id int) {
			defer wg.Done()
			<-startSignal
			_, _, err := bomUsecase.ProduceItem(userCtx, testBOM.ID, warehouse.ID, testUser.ID, 1.0)
			if err != nil {
				errorsChan <- fmt.Errorf("ProduceItem %d failed: %w", id, err)
			}
		}(i)
	}

	// Start all at once
	close(startSignal)
	wg.Wait()
	close(errorsChan)

	// 2. Validate Results
	var fatalErrors []error
	for err := range errorsChan {
		fatalErrors = append(fatalErrors, err)
	}

	for _, err := range fatalErrors {
		t.Errorf("Error during stress test: %v", err)
	}

	assert.Equal(t, 0, len(fatalErrors), "Should have zero errors")

	// Verify Final Stock
	// Component stock should be: initial - numMovements - numProductions * compQtyPerProduct
	// 1000 - 50 - 50 * 1 = 900
	
	// Let's check both bin and nil bin.
	compStockWithBin, _ := stockRepo.GetStock(ctx, compItem.ID, warehouse.ID, &bin.ID)
	compStockNilBin, _ := stockRepo.GetStock(ctx, compItem.ID, warehouse.ID, nil)
	
	totalCompQty := 0.0
	if compStockWithBin != nil {
		totalCompQty += compStockWithBin.Quantity
	}
	if compStockNilBin != nil {
		totalCompQty += compStockNilBin.Quantity
	}
	
	expectedFinalCompQty := initialCompQty - float64(numStockMovements) - float64(numProductions)
	assert.Equal(t, expectedFinalCompQty, totalCompQty, "Final component stock quantity mismatch")

	// Product stock should be: 0 + numProductions
	prodStockNilBin, err := stockRepo.GetStock(ctx, productItem.ID, warehouse.ID, nil)
	require.NoError(t, err)
	assert.Equal(t, float64(numProductions), prodStockNilBin.Quantity, "Final product stock quantity mismatch")

	// 3. Verify Ledger Consistency
	var totalMoves int64
	gormDB.Model(&stock.StockMovement{}).Where("item_id IN ?", []uuid.UUID{compItem.ID, productItem.ID}).Count(&totalMoves)
	// Moves: numStockMovements (comp OUT) + numProductions (comp OUT) + numProductions (product IN)
	// 50 + 50 + 50 = 150
	assert.Equal(t, int64(numStockMovements + 2*numProductions), totalMoves)

	var totalLedger int64
	gormDB.Model(&stock.StockLedger{}).Where("item_id IN ?", []uuid.UUID{compItem.ID, productItem.ID}).Count(&totalLedger)
	assert.Equal(t, int64(numStockMovements + 2*numProductions), totalLedger)
}
