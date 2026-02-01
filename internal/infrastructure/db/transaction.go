// Package db provides database-related utilities, including transaction management.
package db

import (
	"context"
	"gorm.io/gorm"
)

// Transactioner defines an interface for executing database transactions.
// This abstraction allows use cases to perform operations atomically without
// being tightly coupled to the specific database implementation.
type Transactioner interface {
	Transaction(ctx context.Context, fc func(tx *gorm.DB) error) error
}

// gormTransactioner is a GORM-based implementation of the Transactioner interface.
type gormTransactioner struct {
	db *gorm.DB
}

// NewGormTransactioner creates a new gormTransactioner.
func NewGormTransactioner(db *gorm.DB) Transactioner {
	return &gormTransactioner{db: db}
}

// Transaction executes the given function within a database transaction.
// It automatically handles committing the transaction if the function returns nil
// or rolling it back if an error is returned.
func (t *gormTransactioner) Transaction(ctx context.Context, fc func(tx *gorm.DB) error) error {
	return t.db.WithContext(ctx).Transaction(fc)
}
