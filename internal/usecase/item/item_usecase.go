// Package item contains the use case for managing items.
package item

import (
	"context"
	"doligo_001/internal/api/dto"
	"doligo_001/internal/api/middleware"
	"doligo_001/internal/domain"
	"doligo_001/internal/domain/item"
	uc "doligo_001/internal/usecase"
	"github.com/google/uuid"
)

// Usecase defines the contract for item business logic.
type Usecase interface {
	Create(ctx context.Context, req *dto.CreateItemRequest) (*item.Item, error)
	GetByID(ctx context.Context, id uuid.UUID) (*item.Item, error)
	Update(ctx context.Context, id uuid.UUID, req *dto.UpdateItemRequest) (*item.Item, error)
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context) ([]*item.Item, error)
}

type usecase struct {
	repo         item.Repository
	auditService uc.AuditService
}

// NewUsecase creates a new item usecase.
func NewUsecase(repo item.Repository, auditService uc.AuditService) Usecase {
	return &usecase{
		repo:         repo,
		auditService: auditService,
	}
}

// Create handles the creation of a new item.
func (u *usecase) Create(ctx context.Context, req *dto.CreateItemRequest) (*item.Item, error) {
	userID, _ := domain.UserIDFromContext(ctx)

	i := &item.Item{
		ID:          uuid.New(),
		Name:        req.Name,
		Description: req.Description,
		Type:        item.ItemType(req.Type),
		CostPrice:   req.CostPrice,
		SalePrice:   req.SalePrice,
		IsActive:    true,
	}
	i.SetCreatedBy(userID)
	i.SetUpdatedBy(userID)

	if err := u.repo.Create(ctx, i); err != nil {
		return nil, err
	}

	corrID, _ := middleware.FromContext(ctx)
	u.auditService.Log(ctx, userID, "item", i.ID.String(), "CREATE", nil, i, corrID)

	return i, nil
}

// GetByID retrieves a single item by its ID.
func (u *usecase) GetByID(ctx context.Context, id uuid.UUID) (*item.Item, error) {
	return u.repo.GetByID(ctx, id)
}

// Update handles the update of an existing item.
func (u *usecase) Update(ctx context.Context, id uuid.UUID, req *dto.UpdateItemRequest) (*item.Item, error) {
	userID, _ := domain.UserIDFromContext(ctx)

	oldItem, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	// Create a shallow copy for audit
	oldValues := *oldItem

	i := oldItem
	i.Name = req.Name
	i.Description = req.Description
	i.Type = item.ItemType(req.Type)
	i.CostPrice = req.CostPrice
	i.SalePrice = req.SalePrice
	i.IsActive = req.IsActive
	i.SetUpdatedBy(userID)

	if err := u.repo.Update(ctx, i); err != nil {
		return nil, err
	}

	corrID, _ := middleware.FromContext(ctx)
	u.auditService.Log(ctx, userID, "item", i.ID.String(), "UPDATE", oldValues, i, corrID)

	return i, nil
}

// Delete handles the deletion of an item.
func (u *usecase) Delete(ctx context.Context, id uuid.UUID) error {
	userID, _ := domain.UserIDFromContext(ctx)
	oldItem, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if err := u.repo.Delete(ctx, id); err != nil {
		return err
	}

	corrID, _ := middleware.FromContext(ctx)
	u.auditService.Log(ctx, userID, "item", id.String(), "DELETE", oldItem, nil, corrID)

	return nil
}

// List retrieves all items.
func (u *usecase) List(ctx context.Context) ([]*item.Item, error) {
	return u.repo.List(ctx)
}
