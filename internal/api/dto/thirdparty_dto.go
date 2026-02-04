// Package dto provides data transfer objects for API communication.
package dto

import (
	"time"
	"github.com/google/uuid"
	"doligo_001/internal/domain/thirdparty"
	"doligo_001/internal/api/sanitizer"
)

// CreateThirdPartyRequest defines the structure for creating a new third party.
type CreateThirdPartyRequest struct {
	Name  string `json:"name" validate:"required,min=2,max=255"`
	Email string `json:"email" validate:"required,email"`
	Type  string `json:"type" validate:"required,oneof=CUSTOMER SUPPLIER"`
}

func (r *CreateThirdPartyRequest) Sanitize() {
	r.Name = sanitizer.SanitizeString(r.Name)
	r.Email = sanitizer.SanitizeString(r.Email)
}

// UpdateThirdPartyRequest defines the structure for updating an existing third party.
type UpdateThirdPartyRequest struct {
	Name     string `json:"name" validate:"required,min=2,max=255"`
	Email    string `json:"email" validate:"required,email"`
	Type     string `json:"type" validate:"required,oneof=CUSTOMER SUPPLIER"`
	IsActive bool   `json:"is_active"`
}

func (r *UpdateThirdPartyRequest) Sanitize() {
	r.Name = sanitizer.SanitizeString(r.Name)
	r.Email = sanitizer.SanitizeString(r.Email)
}

// ThirdPartyResponse defines the structure for a third party response.
type ThirdPartyResponse struct {
	ID        uuid.UUID  `json:"id"`
	Name      string     `json:"name"`
	Email     string     `json:"email"`
	Type      string     `json:"type"`
	IsActive  bool       `json:"is_active"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	CreatedBy uuid.UUID  `json:"created_by"`
	UpdatedBy uuid.UUID  `json:"updated_by"`
}

// NewThirdPartyResponse creates a response DTO from a domain entity.
func NewThirdPartyResponse(tp *thirdparty.ThirdParty) *ThirdPartyResponse {
	return &ThirdPartyResponse{
		ID:        tp.ID,
		Name:      tp.Name,
		Email:     tp.Email,
		Type:      string(tp.Type),
		IsActive:  tp.IsActive,
		CreatedAt: tp.CreatedAt,
		UpdatedAt: tp.UpdatedAt,
		CreatedBy: tp.CreatedBy,
		UpdatedBy: tp.UpdatedBy,
	}
}
