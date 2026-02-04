package bom

import (
	"context"
	"errors"
	"testing"
	"time"

	"doligo_001/internal/domain/bom"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// fakeBomRepository is a simple fake for the bom.Repository for testing.
type fakeBomRepository struct {
	boms map[uuid.UUID]*bom.BillOfMaterials
}

// newFakeBomRepository initializes a new fake repository.
func newFakeBomRepository() *fakeBomRepository {
	return &fakeBomRepository{
		boms: make(map[uuid.UUID]*bom.BillOfMaterials),
	}
}

func (f *fakeBomRepository) WithTx(tx *gorm.DB) bom.Repository {
	return f
}

func (f *fakeBomRepository) Create(ctx context.Context, bom *bom.BillOfMaterials) error {
	if _, exists := f.boms[bom.ID]; exists {
		return errors.New("bom with this ID already exists")
	}
	f.boms[bom.ID] = bom
	return nil
}

func (f *fakeBomRepository) GetByID(ctx context.Context, id uuid.UUID) (*bom.BillOfMaterials, error) {
	if bom, exists := f.boms[id]; exists {
		return bom, nil
	}
	return nil, errors.New("bom not found")
}

func (f *fakeBomRepository) GetByProductID(ctx context.Context, productID uuid.UUID) (*bom.BillOfMaterials, error) {
	for _, b := range f.boms {
		if b.ProductID == productID {
			return b, nil
		}
	}
	return nil, errors.New("bom not found for product")
}

func (f *fakeBomRepository) Update(ctx context.Context, bom *bom.BillOfMaterials) error {
	if _, exists := f.boms[bom.ID]; !exists {
		return errors.New("bom not found")
	}
	f.boms[bom.ID] = bom
	return nil
}

func (f *fakeBomRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if _, exists := f.boms[id]; !exists {
		return errors.New("bom not found")
	}
	delete(f.boms, id)
	return nil
}

func (f *fakeBomRepository) List(ctx context.Context) ([]*bom.BillOfMaterials, error) {
	var bomList []*bom.BillOfMaterials
	for _, b := range f.boms {
		bomList = append(bomList, b)
	}
	return bomList, nil
}

func TestBomUsecase_GetBOMByID(t *testing.T) {
	repo := newFakeBomRepository()
	usecase := NewBOMUsecase(nil, repo, nil, nil, nil, nil, nil, nil)

	bomID := uuid.New()
	productID := uuid.New()
	expectedBom := &bom.BillOfMaterials{
		ID:        bomID,
		ProductID: productID,
		Name:      "Test BOM",
		IsActive:  true,
		CreatedAt: time.Now(),
	}
	repo.Create(context.Background(), expectedBom)

	tests := []struct {
		name    string
		bomID   uuid.UUID
		wantErr bool
		err     error
	}{
		{
			name:    "happy path - bom found",
			bomID:   bomID,
			wantErr: false,
		},
		{
			name:    "error - bom not found",
			bomID:   uuid.New(),
			wantErr: true,
			err:     errors.New("bom not found"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			retrievedBom, err := usecase.GetBOMByID(context.Background(), tt.bomID)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetBOMByID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && retrievedBom.ID != tt.bomID {
				t.Errorf("GetBOMByID() got = %v, want %v", retrievedBom.ID, tt.bomID)
			}
			
			if tt.wantErr && err.Error() != tt.err.Error() {
				t.Errorf("GetBOMByID() error = %v, want %v", err, tt.err)
			}
		})
	}
}
