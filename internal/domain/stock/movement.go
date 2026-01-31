package stock

import (
	"context"
	"time"

	"doligo_001/internal/domain/item"
)

// MovementType define a natureza da movimentação de estoque.
type MovementType string

const (
	// Movimentações de Entrada
	MovementTypePurchaseIn   MovementType = "PURCHASE_IN"   // Entrada por compra
	MovementTypeTransferIn   MovementType = "TRANSFER_IN"   // Entrada por transferência entre armazéns
	MovementTypeAdjustmentIn MovementType = "ADJUSTMENT_IN" // Ajuste positivo de inventário

	// Movimentações de Saída
	MovementTypeSaleOut      MovementType = "SALE_OUT"      // Saída por venda
	MovementTypeTransferOut  MovementType = "TRANSFER_OUT"  // Saída por transferência
	MovementTypeAdjustmentOut MovementType = "ADJUSTMENT_OUT"// Ajuste negativo de inventário
	MovementTypeConsumptionOut MovementType = "CONSUMPTION_OUT"// Saída por consumo interno/produção
)

// StockMovement representa a transação de entrada ou saída de um item do estoque.
// Esta é a entidade central para a lógica de locking (SELECT FOR UPDATE).
type StockMovement struct {
	ID          int64
	Item        *item.Item // Referência ao item sendo movimentado
	WarehouseID int64
	BinID       *int64 // Opcional, se o controle for granular
	Quantity    float64    // A quantidade movimentada. Positiva para entradas, negativa para saídas.
	MovementType MovementType
	Reason      string    // Motivo do ajuste ou observação
	RefSource   string    // Documento de referência (ex: "PO-123", "SO-456", "ADJ-789")
	MovedAt     time.Time
	CreatedBy   *int64
}

// StockMovementRepository define a interface de persistência para movimentações de estoque.
// É aqui que a lógica de locking transacional será implementada.
type StockMovementRepository interface {
	// ApplyMovement aplica uma movimentação de estoque de forma transacional e com lock.
	// A implementação deve garantir que as leituras e escrituras de saldo (em outra tabela)
	// sejam feitas sob um lock pessimista (SELECT ... FOR UPDATE).
	ApplyMovement(ctx context.Context, movement *StockMovement) error
}
