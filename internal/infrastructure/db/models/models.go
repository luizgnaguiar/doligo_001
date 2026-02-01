// Package models contains the GORM data models that map to the database schema.
// These models are used by the infrastructure layer (repositories) to interact
// with the database. They are kept separate from the domain entities to decouple
// the application's core logic from the persistence implementation details.
// This separation allows the database schema to evolve independently of the
// domain model and vice-versa.
package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Auditable defines a common interface for models that require auditing fields.
// This ensures that CreatedBy and UpdatedBy fields can be set polymorphically.
type Auditable interface {
	SetCreatedBy(userID uuid.UUID)
	SetUpdatedBy(userID uuid.UUID)
}

// BaseModel is an abstract base model that provides common fields for all GORM models.
// It includes an auto-incrementing ID, a UUID for public reference, and timestamps.
// The gorm.Model is embedded to provide ID, CreatedAt, UpdatedAt, and DeletedAt.
type BaseModel struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
	CreatedBy uuid.UUID      `gorm:"type:uuid"`
	UpdatedBy uuid.UUID      `gorm:"type:uuid"`
}

// User model represents the database schema for users.
type User struct {
	BaseModel
	FirstName string `gorm:"size:100;not null"`
	LastName  string `gorm:"size:100"`
	Email     string `gorm:"size:255;not null;uniqueIndex"`
	Password  string `gorm:"size:255;not null"`
	IsActive  bool   `gorm:"default:true"`
	Roles     []Role `gorm:"many2many:user_roles;"`
}

// SetCreatedBy sets the CreatedBy field for the User model.
func (u *User) SetCreatedBy(userID uuid.UUID) {
	u.CreatedBy = userID
}

// SetUpdatedBy sets the UpdatedBy field for the User model.
func (u *User) SetUpdatedBy(userID uuid.UUID) {
	u.UpdatedBy = userID
}

// Role model represents the database schema for roles.
type Role struct {
	BaseModel
	Name        string       `gorm:"size:100;not null;uniqueIndex"`
	Description string       `gorm:"size:255"`
	Permissions []Permission `gorm:"many2many:role_permissions;"`
	Users       []User       `gorm:"many2many:user_roles;"`
}

// Permission model represents the database schema for permissions.
type Permission struct {
	BaseModel
	Name        string `gorm:"size:100;not null;uniqueIndex"`
	Description string `gorm:"size:255"`
	Roles       []Role `gorm:"many2many:role_permissions;"`
}

// ThirdParty model represents the database schema for customers and suppliers.
type ThirdParty struct {
	BaseModel
	Name      string `gorm:"size:255;not null;index"`
	Email     string `gorm:"size:255;not null;uniqueIndex"`
	Type      string `gorm:"size:50;not null"` // 'CUSTOMER' or 'SUPPLIER'
	IsActive  bool   `gorm:"default:true"`
	CreatedByUser User `gorm:"foreignKey:CreatedBy"`
	UpdatedByUser User `gorm:"foreignKey:UpdatedBy"`
}

// Item model represents the database schema for products and services.
type Item struct {
	BaseModel
	Name        string  `gorm:"size:255;not null;index"`
	Description string
	Type        string  `gorm:"size:50;not null"` // 'STORABLE' or 'SERVICE'
	CostPrice   float64 `gorm:"type:numeric(15,4);default:0.0"`
	SalePrice   float64 `gorm:"type:numeric(15,4);default:0.0"`
	AverageCost float64 `gorm:"type:numeric(15,4);default:0.0"`
	IsActive    bool    `gorm:"default:true"`
	CreatedByUser User `gorm:"foreignKey:CreatedBy"`
	UpdatedByUser User `gorm:"foreignKey:UpdatedBy"`
}

// Warehouse model represents the database schema for a stock warehouse.
type Warehouse struct {
	BaseModel
	Name     string `gorm:"size:255;not null;uniqueIndex"`
	IsActive bool   `gorm:"default:true"`
	Bins     []Bin  `gorm:"foreignKey:WarehouseID"`
}

// Bin model represents a specific storage bin within a warehouse.
type Bin struct {
	BaseModel
	Name        string    `gorm:"size:255;not null"`
	WarehouseID uuid.UUID `gorm:"type:uuid;not null"`
	Warehouse   Warehouse `gorm:"foreignKey:WarehouseID"`
	IsActive    bool      `gorm:"default:true"`
}

// Stock model represents the current quantity of an item at a specific location.
// It serves as the single source of truth for current inventory levels.
type Stock struct {
	ItemID      uuid.UUID `gorm:"type:uuid;primaryKey"`
	WarehouseID uuid.UUID `gorm:"type:uuid;primaryKey"`
	BinID       uuid.UUID `gorm:"type:uuid;primaryKey;default:'00000000-0000-0000-0000-000000000000'"` // Use a zero UUID for non-binned stock
	Quantity    float64   `gorm:"type:numeric(15,4);not null;default:0.0"`
	UpdatedAt   time.Time
	Item        Item      `gorm:"foreignKey:ItemID"`
	Warehouse   Warehouse `gorm:"foreignKey:WarehouseID"`
	Bin         Bin       `gorm:"foreignKey:BinID"`
}

// StockMovement model represents a record of stock moving in or out.
type StockMovement struct {
	ID          uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	ItemID      uuid.UUID  `gorm:"type:uuid;not null;index"`
	WarehouseID uuid.UUID  `gorm:"type:uuid;not null;index"`
	BinID       *uuid.UUID `gorm:"type:uuid;index"`
	Type        string     `gorm:"size:10;not null"` // 'IN' or 'OUT'
	Quantity    float64    `gorm:"type:numeric(15,4);not null"`
	Reason      string     `gorm:"size:255"`
	HappenedAt  time.Time  `gorm:"not null"`
	CreatedBy   uuid.UUID  `gorm:"type:uuid"`
	Item        Item       `gorm:"foreignKey:ItemID"`
	Warehouse   Warehouse  `gorm:"foreignKey:WarehouseID"`
	Bin         *Bin       `gorm:"foreignKey:BinID"`
	CreatedByUser User     `gorm:"foreignKey:CreatedBy"`
}

// StockLedger model is an immutable, append-only log of all stock transactions.
type StockLedger struct {
	ID              uuid.UUID     `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	StockMovementID uuid.UUID     `gorm:"type:uuid;not null;index"`
	ItemID          uuid.UUID     `gorm:"type:uuid;not null;index"`
	WarehouseID     uuid.UUID     `gorm:"type:uuid;not null;index"`
	BinID           *uuid.UUID    `gorm:"type:uuid;index"`
	MovementType    string        `gorm:"size:10;not null"`
	QuantityChange  float64       `gorm:"type:numeric(15,4);not null"`
	QuantityBefore  float64       `gorm:"type:numeric(15,4);not null"`
	QuantityAfter   float64       `gorm:"type:numeric(15,4);not null"`
	Reason          string        `gorm:"size:255"`
	HappenedAt      time.Time     `gorm:"not null"`
	RecordedAt      time.Time     `gorm:"not null;default:now()"`
	RecordedBy      uuid.UUID     `gorm:"type:uuid"`
	Item            Item          `gorm:"foreignKey:ItemID"`
	Warehouse       Warehouse     `gorm:"foreignKey:WarehouseID"`
	Bin             *Bin          `gorm:"foreignKey:BinID"`
	RecordedByUser  User          `gorm:"foreignKey:RecordedBy"`
	StockMovement   StockMovement `gorm:"foreignKey:StockMovementID"`
}
