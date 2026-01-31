package stock

import (
	"context"
	"time"
)

// Bin representa uma localização específica dentro de um Warehouse.
// Exemplos: uma prateleira, uma gaveta, uma posição de palete (ex: A-01-B-03).
// A granularidade do controle de estoque é definida pelo uso (ou não) de Bins.
type Bin struct {
	ID          int64
	WarehouseID int64  // Chave estrangeira para o Warehouse pai
	Code        string // Código único do Bin dentro do armazém (ex: "PR-05-A")
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// BinRepository define a interface de persistência para Bins.
type BinRepository interface {
	Create(ctx context.Context, bin *Bin) error
	FindByID(ctx context.Context, id int64) (*Bin, error)
	// FindByCode busca um Bin pelo seu código dentro de um armazém específico.
	FindByCode(ctx context.Context, warehouseID int64, code string) (*Bin, error)
	Update(ctx context.Context, bin *Bin) error
}
