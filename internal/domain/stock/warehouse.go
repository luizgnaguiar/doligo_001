package stock

import (
	"context"
	"time"
)

// Warehouse representa um local físico ou virtual de armazenamento de estoque.
// Pode ser um armazém principal, uma loja, um veículo de serviço ou uma área de quarentena.
type Warehouse struct {
	ID        int64
	Name      string    // Nome do armazém (ex: "Armazém Principal", "Loja Filial A")
	IsVirtual bool      // `true` para locais não-físicos (ex: "Em Trânsito", "Ajuste de Inventário")
	Address   string    // Endereço físico (rua, número, etc.)
	City      string
	State     string
	ZipCode   string
	Country   string
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time // Suporte a soft-delete
	CreatedBy *int64
	UpdatedBy *int64
}

// WarehouseRepository define a interface de persistência para armazéns.
type WarehouseRepository interface {
	Create(ctx context.Context, warehouse *Warehouse) error
	FindByID(ctx context.Context, id int64) (*Warehouse, error)
	Update(ctx context.Context, warehouse *Warehouse) error
}
