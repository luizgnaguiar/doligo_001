// Package thirdparty contains the use case for managing third parties.
// It orchestrates the business logic, coordinating with the repository
// and handling auditing.
package thirdparty

import (
	"context"
	"doligo_001/internal/api/dto"
	"doligo_001/internal/domain"
	"doligo_001/internal/domain/thirdparty"
	"github.com/google/uuid"
)

// Usecase defines the contract for third party business logic.
type Usecase interface {
	Create(ctx context.Context, req *dto.CreateThirdPartyRequest) (*thirdparty.ThirdParty, error)
	GetByID(ctx context.Context, id uuid.UUID) (*thirdparty.ThirdParty, error)
	Update(ctx context.Context, id uuid.UUID, req *dto.UpdateThirdPartyRequest) (*thirdparty.ThirdParty, error)
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context) ([]*thirdparty.ThirdParty, error)
}

type usecase struct {
	repo thirdparty.Repository
}

// NewUsecase creates a new third party usecase.
func NewUsecase(repo thirdparty.Repository) Usecase {
	return &usecase{repo: repo}
}

// Create handles the creation of a new third party.
func (u *usecase) Create(ctx context.Context, req *dto.CreateThirdPartyRequest) (*thirdparty.ThirdParty, error) {
	userID, _ := domain.UserIDFromContext(ctx)

	tp := &thirdparty.ThirdParty{
		ID:    uuid.New(),
		Name:  req.Name,
		Email: req.Email,
		Type:  thirdparty.ThirdPartyType(req.Type),
		IsActive: true,
	}
	tp.SetCreatedBy(userID)
	tp.SetUpdatedBy(userID)

	if err := u.repo.Create(ctx, tp); err != nil {
		return nil, err
	}
	return tp, nil
}

// GetByID retrieves a single third party by its ID.
func (u *usecase) GetByID(ctx context.Context, id uuid.UUID) (*thirdparty.ThirdParty, error) {
	return u.repo.GetByID(ctx, id)
}

// Update handles the update of an existing third party.
func (u *usecase) Update(ctx context.Context, id uuid.UUID, req *dto.UpdateThirdPartyRequest) (*thirdparty.ThirdParty, error) {
	userID, _ := domain.UserIDFromContext(ctx)

	tp, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	tp.Name = req.Name
	tp.Email = req.Email
	tp.Type = thirdparty.ThirdPartyType(req.Type)
	tp.IsActive = req.IsActive
	tp.SetUpdatedBy(userID)

	if err := u.repo.Update(ctx, tp); err != nil {
		return nil, err
	}
	return tp, nil
}

// Delete handles the deletion of a third party.
func (u *usecase) Delete(ctx context.Context, id uuid.UUID) error {
	return u.repo.Delete(ctx, id)
}

// List retrieves all third parties.
func (u *usecase) List(ctx context.Context) ([]*thirdparty.ThirdParty, error) {
	return u.repo.List(ctx)
}
