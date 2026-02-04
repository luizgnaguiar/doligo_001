package repository_test

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"sync"
	"testing"
	"time"

	"doligo_001/internal/domain/item"
	"doligo_001/internal/domain/stock"
	"doligo_001/internal/infrastructure/config"
	"doligo_001/internal/infrastructure/db"
	"doligo_001/internal/infrastructure/repository"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

var (
	testDB       *gorm.DB
	testSQLDB    *sql.DB
	testTxManager db.Transactioner
)

func TestMain(m *testing.M) {
	// Load test configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Use a test database name
	cfg.Database.Name = fmt.Sprintf("%s_test_%d", cfg.Database.Name, time.Now().UnixNano())

	// Initialize test database
	ctx := context.Background()
	var dsn string
	testDB, dsn, err = db.InitDatabase(ctx, &cfg.Database)
	if err != nil {
		log.Fatalf("Failed to initialize test database: %v", err)
	}

	testSQLDB, err = testDB.DB()
	if err != nil {
		log.Fatalf("Failed to get generic database object: %v", err)
	}
	defer func() {
		if testSQLDB != nil {
			err := testSQLDB.Close()
			if err != nil {
				log.Printf("Error closing test database: %v", err)
			}
		}
	}()

	// Run migrations
	err = db.RunMigrations(ctx, testSQLDB, cfg.Database.Type, dsn)
	if err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	testTxManager = db.NewGormTransactioner(testDB)

	// Run tests
	code := m.Run()

	// Teardown: Drop the test database
	// This part might need to be adjusted based on the database type and permissions
	// For PostgreSQL, connecting to default 'postgres' db to drop the test db
	if cfg.Database.Type == "postgres" {
		cleanupDB, _, err := db.InitDatabase(ctx, &config.DatabaseConfig{
			Type: cfg.Database.Type,
			Host: cfg.Database.Host,
			Port: cfg.Database.Port,
			User: cfg.Database.User,
			Password: cfg.Database.Password,
			Name: "postgres", // Connect to default db to drop test db
			SSLMode: cfg.Database.SSLMode,
			MaxOpenConns: 1,
			MaxIdleConns: 1,
			ConnMaxLifetime: time.Minute,
		})
		if err != nil {
			log.Printf("WARNING: Failed to connect to default database for cleanup: %v", err)
		} else {
			cleanupSQLDB, _ := cleanupDB.DB()
			_, err = cleanupSQLDB.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s WITH (FORCE);", cfg.Database.Name))
			if err != nil {
				log.Printf("WARNING: Failed to drop test database '%s': %v", cfg.Database.Name, err)
			} else {
				log.Printf("Test database '%s' dropped successfully.", cfg.Database.Name)
			}
			cleanupSQLDB.Close()
		}
	} else if cfg.Database.Type == "mysql" {
		// MySQL cleanup needs to be handled with caution, typically requires connecting without a specific DB
		// and then dropping it. For simplicity, we'll just log a reminder.
		log.Printf("WARNING: Manual cleanup might be required for MySQL test database: %s", cfg.Database.Name)
	}


	os.Exit(code)
}

func setupTest(t *testing.T) (context.Context, *gorm.DB, db.Transactioner) {
	// Each test runs in its own transaction to ensure isolation
	// The transaction is rolled back at the end of the test
	tx := testDB.Begin()
	require.NotNil(t, tx, "failed to begin transaction")

	// Use this transaction for all repositories in the test
	testTxManager := db.NewGormTransactioner(tx)

	t.Cleanup(func() {
		// Rollback the transaction at the end of the test
		if tx != nil {
			tx.Rollback()
		}
	})

	return context.Background(), tx, testTxManager
}

func TestPessimisticLockingCT02(t *testing.T) {
	ctx := context.Background()
	// Do NOT use setupTest here because we need real concurrent transactions
	// that can see each other's committed data and block each other.
	// We will manually cleanup.
	
	stockRepo := repository.NewGormStockRepository(testDB)
	itemRepo := repository.NewGormItemRepository(testDB)
	warehouseRepo := repository.NewGormWarehouseRepository(testDB)
	binRepo := repository.NewGormBinRepository(testDB)

	// 1. Create necessary entities
	testItem := &item.Item{
		ID:   uuid.New(),
		Name: "Test Item CT02",
		Type: item.Storable,
	}
	require.NoError(t, itemRepo.Create(ctx, testItem))

	testWarehouse := &stock.Warehouse{
		ID:   uuid.New(),
		Name: "Test Warehouse CT02",
	}
	require.NoError(t, warehouseRepo.Create(ctx, testWarehouse))

	testBin := &stock.Bin{
		ID:          uuid.New(),
		Name:        "Test Bin CT02",
		WarehouseID: testWarehouse.ID,
	}
	require.NoError(t, binRepo.Create(ctx, testBin))

	// Cleanup at the end
	defer func() {
		testDB.Exec("DELETE FROM stocks WHERE item_id = ?", testItem.ID)
		testDB.Exec("DELETE FROM bins WHERE id = ?", testBin.ID)
		testDB.Exec("DELETE FROM warehouses WHERE id = ?", testWarehouse.ID)
		testDB.Exec("DELETE FROM items WHERE id = ?", testItem.ID)
	}()

	initialQuantity := 100.0
	debitQuantity := 10.0
	sleepDuration := 500 * time.Millisecond // Simulate work inside first transaction

	// Create initial stock
	initialStock := &stock.Stock{
		ItemID:      testItem.ID,
		WarehouseID: testWarehouse.ID,
		BinID:       &testBin.ID,
		Quantity:    initialQuantity,
		UpdatedAt:   time.Now(),
	}
	require.NoError(t, stockRepo.UpsertStock(ctx, initialStock))

	var wg sync.WaitGroup
	wg.Add(2)

	startSignal := make(chan struct{})

	// Transaction 1 Goroutine
	go func() {
		defer wg.Done()
		<-startSignal
		err := testTxManager.Transaction(context.Background(), func(tx *gorm.DB) error {
			stockRepo1 := repository.NewGormStockRepository(tx)
			
			// Get stock with pessimistic lock
			s, err := stockRepo1.GetStockForUpdate(context.Background(), testItem.ID, testWarehouse.ID, &testBin.ID)
			if err != nil {
				return err
			}
			assert.Equal(t, initialQuantity, s.Quantity, "Transaction 1: Initial quantity mismatch")

			// Simulate work
			time.Sleep(sleepDuration)

			// Debit stock
			s.Quantity -= debitQuantity
			err = stockRepo1.UpsertStock(context.Background(), s)
			return err
		})
		assert.NoError(t, err, "Transaction 1 failed")
	}()

	// Transaction 2 Goroutine
	var secondTxStartTime time.Time
	var secondTxEndTime time.Time
	go func() {
		defer wg.Done()
		
		// Ensure T1 starts first
		time.Sleep(100 * time.Millisecond)
		<-startSignal

		err := testTxManager.Transaction(context.Background(), func(tx *gorm.DB) error {
			stockRepo2 := repository.NewGormStockRepository(tx)

			secondTxStartTime = time.Now()
			// Attempt to get stock with pessimistic lock - should block
			s, err := stockRepo2.GetStockForUpdate(context.Background(), testItem.ID, testWarehouse.ID, &testBin.ID)
			secondTxEndTime = time.Now()
			
			if err != nil {
				return err
			}
			assert.Equal(t, initialQuantity-debitQuantity, s.Quantity, "Transaction 2: Quantity after first debit mismatch")

			// Debit stock
			s.Quantity -= debitQuantity
			err = stockRepo2.UpsertStock(context.Background(), s)
			return err
		})
		assert.NoError(t, err, "Transaction 2 failed")
	}()

	close(startSignal)
	wg.Wait()

	// Verify the second transaction waited
	duration := secondTxEndTime.Sub(secondTxStartTime)
	t.Logf("Second transaction waited for %v", duration)
	// T1 sleeps for sleepDuration. T2 starts 100ms after startSignal.
	// So T2 should wait at least sleepDuration - 100ms.
	expectedWait := sleepDuration - 150*time.Millisecond 
	assert.GreaterOrEqual(t, duration, expectedWait, "Second transaction did not wait enough for the first transaction's lock")

	// Verify final stock quantity
	finalStock, err := stockRepo.GetStock(ctx, testItem.ID, testWarehouse.ID, &testBin.ID)
	assert.NoError(t, err)
	assert.NotNil(t, finalStock)
	assert.Equal(t, initialQuantity-(2*debitQuantity), finalStock.Quantity, "Final stock quantity mismatch")
}
