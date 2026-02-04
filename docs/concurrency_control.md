# Concurrency Control Strategy

## Overview
This document details the concurrency control mechanisms implemented in the Doligo ERP/CRM system to ensure data integrity and prevent race conditions in highly concurrent environments, specifically focusing on inventory and production operations.

## Pessimistic Locking Strategy

The system employs **Pessimistic Locking** via `SELECT FOR UPDATE` (or database-equivalent) to manage critical sections where multiple transactions might attempt to modify the same resource simultaneously.

### Stock Management
In the `stockUsecase` and `bomUsecase`, stock updates are protected by a pessimistic lock on the `stock` table record.

- **Mechanism:** `stockRepo.GetStockForUpdate(ctx, itemID, warehouseID, binID)`
- **SQL Implementation:** `SELECT * FROM stock WHERE item_id = ? AND warehouse_id = ? AND bin_id = ? FOR UPDATE`
- **Scope:**
  - `CreateStockMovement`: Locks the source/destination stock record before updating quantity.
  - `ProduceItem`: Locks all component stock records (OUT) and the finished product stock record (IN).

### Production (BOM)
During production, the system ensures that component availability is verified and consumed atomically.

1.  **Start Transaction:** All operations within `ProduceItem` are wrapped in a single database transaction.
2.  **Lock Components:** For each component in the Bill of Materials:
    - Fetch the current stock record using `FOR UPDATE`.
    - Validate that `Quantity >= neededQty`.
    - Deduct the quantity and `UpsertStock`.
3.  **Lock Product:** Fetch or create the finished product's stock record using `FOR UPDATE`.
    - Increment the quantity and `UpsertStock`.
4.  **Audit and Commit:** Record the production, stock movements, and ledger entries before committing the transaction.

## Evidence of Success (CT-02)

The concurrency strategy was validated through a high-concurrency stress test (`TestHighConcurrencyStockAndBOM` in `internal/infrastructure/repository/integration_stress_test.go`).

### Test Parameters:
- **Concurrent Operations:** 100 simultaneous goroutines.
  - 50 `CreateStockMovement` calls (direct stock deduction).
  - 50 `ProduceItem` calls (BOM-based stock consumption and product generation).
- **Environment:** Shared component and product stock records.

### Results:
- **Zero Errors:** All 100 operations completed successfully without deadlocks or constraint violations.
- **Data Integrity:** The final stock quantities for both the component and the produced item matched the expected mathematical results exactly.
- **Ledger Consistency:** Every stock change was correctly reflected in the `stock_ledger` and `stock_movement` tables, maintaining a perfect audit trail.

## Operational Considerations

### Lock Contention
While pessimistic locking ensures integrity, it can lead to lock contention under extremely high load on a single item.

- **Monitoring:** Monitor database wait times and `Locked` states in the process list.
- **Connection Pool:** Ensure the connection pool (`MaxOpenConns`) is sized appropriately to handle pending transactions during lock waits.
- **Timeout:** Transactional operations have a context timeout (default 30s) to prevent indefinite blocking in case of database issues.

### Impact on Performance
`SELECT FOR UPDATE` serializes access to specific rows. For the majority of ERP operations, which are distributed across thousands of items, the impact is negligible. For "hot" items, consider:
- Optimizing transaction length (keeping the lock duration as short as possible).
- Ensuring indexes are present on `(item_id, warehouse_id, bin_id)` to avoid table scans during locking.
