package item

import (
	"context"
	"time"
)

// ItemType define se o item é um produto estocável ou um serviço.
type ItemType string

const (
	// TypeStorable indica que o item é um produto físico, sujeito a controle de estoque.
	TypeStorable ItemType = "STORABLE"
	// TypeService indica que o item é um serviço, como "hora-homem", não sujeito a estoque.
	TypeService ItemType = "SERVICE"
)

// Item representa um produto ou serviço comercializável no sistema.
type Item struct {
	ID          int64
	Name        string
	Description string
	Type        ItemType
	Unit        string // Ex: "un", "kg", "h" para hora

	// Campos de Precificação
	CostPrice   float64 // Preço de Custo
	SalePrice   float64 // Preço de Venda
	AverageCost float64 // Custo Médio Ponderado (calculado, não inserido diretamente)

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
	CreatedBy *int64
	UpdatedBy *int64
}

// ItemRepository define a interface de persistência para Itens.
// Os métodos devem ser implementados pela camada de infraestrutura.
type ItemRepository interface {
	// Create insere um novo item no banco de dados.
	Create(ctx context.Context, item *Item) error
	// FindByID busca um item pelo seu ID.
	FindByID(ctx context.Context, id int64) (*Item, error)
	// Update atualiza os dados de um item existente.
	Update(ctx context.Context, item *Item) error
	// Delete marca um item como removido (soft delete).
	Delete(ctx context.Context, id int64) error
}
