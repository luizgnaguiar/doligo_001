package thirdparty

import (
	"context"
	"time"
)

// ThirdPartyType define se a entidade é um Cliente ou Fornecedor.
type ThirdPartyType string

const (
	// TypeCustomer indica que o terceiro é um Cliente.
	TypeCustomer ThirdPartyType = "CUSTOMER"
	// TypeSupplier indica que o terceiro é um Fornecedor.
	TypeSupplier ThirdPartyType = "SUPPLIER"
)

// ThirdParty representa um cliente ou fornecedor no sistema.
// É a entidade central para qualquer pessoa física ou jurídica com a qual a empresa se relaciona.
type ThirdParty struct {
	ID             int64
	Name           string // Nome fantasia ou nome social
	LegalName      string // Razão social ou nome completo de pessoa física
	FederalTaxID   string // CPF ou CNPJ, sem formatação
	Email          string
	PhoneNumber    string
	Type           ThirdPartyType
	CreatedAt      time.Time
	UpdatedAt      time.Time
	DeletedAt      *time.Time
	CreatedBy      *int64
	UpdatedBy      *int64
}

// ThirdPartyRepository define a interface de persistência para Terceiros.
// Os métodos devem ser implementados pela camada de infraestrutura.
type ThirdPartyRepository interface {
	// Create insere um novo terceiro no banco de dados.
	Create(ctx context.Context, thirdParty *ThirdParty) error
	// FindByID busca um terceiro pelo seu ID.
	FindByID(ctx context.Context, id int64) (*ThirdParty, error)
	// Update atualiza os dados de um terceiro existente.
	Update(ctx context.Context, thirdParty *ThirdParty) error
	// Delete marca um terceiro como removido (soft delete).
	Delete(ctx context.Context, id int64) error
}
