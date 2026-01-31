package stock

import (
	"context"
	"errors"
	"time"

	"doligo_001/internal/domain/item"
	"doligo_001/internal/domain/stock"
)

var (
	ErrNegativeStock = errors.New("operation would result in negative stock")
	ErrItemNotFound  = errors.New("item not found")
	ErrWarehouseNotFound = errors.New("warehouse not found")
)

// Usecase defines the business logic for stock operations.
type Usecase struct {
	stockRepo stock.StockMovementRepository
	ledgerRepo stock.StockLedgerRepository
	itemRepo  item.ItemRepository
	warehouseRepo stock.WarehouseRepository
	// config      *Config // Config can be added for rules like "allowNegativeStock"
}

// NewUsecase creates a new stock usecase instance.
func NewUsecase(
	stockRepo stock.StockMovementRepository,
	ledgerRepo stock.StockLedgerRepository,
	itemRepo item.ItemRepository,
	warehouseRepo stock.WarehouseRepository,
) *Usecase {
	return &Usecase{
		stockRepo:     stockRepo,
		ledgerRepo:    ledgerRepo,
		itemRepo:      itemRepo,
		warehouseRepo: warehouseRepo,
	}
}

// MoveStockInput represents the data required to execute a stock movement.
type MoveStockInput struct {
	ItemID       int64
	WarehouseID  int64
	Quantity     float64 // Positive for IN, Negative for OUT
	MovementType stock.MovementType
	Reason       string
	RefSource    string
}

// MoveStock orchestrates a stock movement, including validation and auditing.
func (uc *Usecase) MoveStock(ctx context.Context, input MoveStockInput) error {
	// 1. Validar inputs
	itemData, err := uc.itemRepo.FindByID(ctx, input.ItemID)
	if err != nil {
		return ErrItemNotFound
	}
	if _, err := uc.warehouseRepo.FindByID(ctx, input.WarehouseID); err != nil {
		return ErrWarehouseNotFound
	}

	// 2. Construir a entidade de movimentação
	movement := &stock.StockMovement{
		Item:         itemData,
		WarehouseID:  input.WarehouseID,
		Quantity:     input.Quantity,
		MovementType: input.MovementType,
		Reason:       input.Reason,
		RefSource:    input.RefSource,
		MovedAt:      time.Now(),
	}

	// 3. Aplicar a movimentação (a camada de repositório cuidará do lock e da transação)
	// A lógica de validação de estoque negativo e a criação do ledger
	// devem ser tratadas DENTRO da implementação do repositório para garantir
	// que ocorram na mesma transação e sob o mesmo lock pessimista.
	if err := uc.stockRepo.ApplyMovement(ctx, movement); err != nil {
		// O repositório pode retornar erros específicos, como ErrNegativeStock
		return err
	}

	return nil
}
