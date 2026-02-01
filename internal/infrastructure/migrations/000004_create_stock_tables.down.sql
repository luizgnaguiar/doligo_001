-- 000004_create_stock_tables.down.sql
-- This script reverts the changes from 000004_create_stock_tables.up.sql.

DROP TABLE IF EXISTS stock_ledger;
DROP TABLE IF EXISTS stock_movements;
DROP TABLE IF EXISTS stocks;
DROP TABLE IF EXISTS bins;
DROP TABLE IF EXISTS warehouses;
