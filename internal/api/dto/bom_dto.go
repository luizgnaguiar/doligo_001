package dto

import "github.com/google/uuid"

// CreateBOMRequest represents the request body for creating a new Bill of Materials.
type CreateBOMRequest struct {
	ProductID  uuid.UUID            `json:"product_id" validate:"required"`
	Name       string               `json:"name" validate:"required"`
	IsActive   bool                 `json:"is_active"`
	Components []BOMComponentRequest `json:"components" validate:"required,min=1"`
}

// BOMComponentRequest represents a single component within a BOM creation request.
type BOMComponentRequest struct {
	ComponentItemID uuid.UUID `json:"component_item_id" validate:"required"`
	Quantity        float64   `json:"quantity" validate:"required,gt=0"`
	UnitOfMeasure   string    `json:"unit_of_measure" validate:"required"`
	IsActive        bool      `json:"is_active"`
}

// BOMResponse represents the response body for a Bill of Materials.
type BOMResponse struct {
	ID         uuid.UUID             `json:"id"`
	ProductID  uuid.UUID             `json:"product_id"`
	Name       string                `json:"name"`
	IsActive   bool                  `json:"is_active"`
	Components []BOMComponentResponse `json:"components"`
	CreatedAt  string                `json:"created_at"`
	UpdatedAt  string                `json:"updated_at"`
	CreatedBy  uuid.UUID             `json:"created_by"`
	UpdatedBy  uuid.UUID             `json:"updated_by"`
}

// BOMComponentResponse represents a single component within a BOM response.
type BOMComponentResponse struct {
	ID              uuid.UUID `json:"id"`
	BillOfMaterialsID uuid.UUID `json:"bill_of_materials_id"`
	ComponentItemID   uuid.UUID `json:"component_item_id"`
	Quantity        float64   `json:"quantity"`
	UnitOfMeasure   string    `json:"unit_of_measure"`
	IsActive        bool      `json:"is_active"`
	CreatedAt       string    `json:"created_at"`
	UpdatedAt       string    `json:"updated_at"`
	CreatedBy       uuid.UUID `json:"created_by"`
	UpdatedBy       uuid.UUID `json:"updated_by"`
}

// CalculateCostRequest represents the request body for calculating predictive cost.
type CalculateCostRequest struct {
	BOMID      uuid.UUID `json:"bom_id" validate:"required"`
}

// CalculateCostResponse represents the response body for predictive cost calculation.
type CalculateCostResponse struct {
	BOMID    uuid.UUID `json:"bom_id"`
	TotalCost float64   `json:"total_cost"`
}

// ProduceItemRequest represents the request body for initiating a production order.
type ProduceItemRequest struct {
	BOMID              uuid.UUID `json:"bom_id" validate:"required"`
	WarehouseID        uuid.UUID `json:"warehouse_id" validate:"required"`
	ProductionQuantity float64   `json:"production_quantity" validate:"required,gt=0"`
}

// ProduceItemResponse represents the response body for a production order.
type ProduceItemResponse struct {
	ProductionRecordID   uuid.UUID `json:"production_record_id"`
	ActualProductionCost float64   `json:"actual_production_cost"`
	Message             string    `json:"message"`
}
