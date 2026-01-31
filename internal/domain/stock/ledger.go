package stock

import (
	"context"
	"time"

	"doligo_001/internal/domain/item"
)

// StockLedger registra o histórico imutável de todas as movimentações de estoque.
// Serve como um livro-razão para fins de auditoria e rastreabilidade.
// Cada StockMovement bem-sucedido DEVE gerar uma entrada no StockLedger.
type StockLedger struct {
	ID               int64
	MovementID       int64      // Referência ao StockMovement que originou esta entrada
	Item             *item.Item // Referência ao item
	WarehouseID      int64
	QuantityChange   float64   // A quantidade que mudou (+/-)
	ResultingQuantity float64   // O saldo do item no armazém *após* a movimentação
	MovedAt          time.Time // Timestamp exato da movimentação
	CreatedBy        *int64
}

// StockLedgerRepository define a interface de persistência para o livro-razão.
// A única operação permitida é a criação, garantindo a imutabilidade do registro.
type StockLedgerRepository interface {
	// LogMovement registra uma entrada no ledger.
	// Esta operação deve ocorrer na mesma transação que o ApplyMovement.
	LogMovement(ctx context.Context, entry *StockLedger) error
}
