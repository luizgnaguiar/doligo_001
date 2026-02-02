package invoice

import (
	"context"
	"time"

	"github.com/google/uuid"
	"doligo_001/internal/api/dto"
	"doligo_001/internal/domain"
	"doligo_001/internal/domain/invoice"
	item_usecase "doligo_001/internal/usecase/item"
)

type Usecase interface {
	Create(ctx context.Context, req *dto.CreateInvoiceRequest) (*invoice.Invoice, error)
	GetByID(ctx context.Context, id uuid.UUID) (*invoice.Invoice, error)
}

type usecase struct {
	invoiceRepo Repository
	itemRepo    item_usecase.Repository
}

func NewUsecase(invoiceRepo Repository, itemRepo item_usecase.Repository) Usecase {
	return &usecase{
		invoiceRepo: invoiceRepo,
		itemRepo:    itemRepo,
	}
}

func (u *usecase) Create(ctx context.Context, req *dto.CreateInvoiceRequest) (*invoice.Invoice, error) {
	userID, _ := domain.UserIDFromContext(ctx)
	thirdPartyID, _ := uuid.Parse(req.ThirdPartyID)
	invoiceDate, _ := time.Parse("2006-01-02", req.Date)

	newInvoice := &invoice.Invoice{
		ID:           uuid.New(),
		ThirdPartyID: thirdPartyID,
		Number:       req.Number,
		Date:         invoiceDate,
	}

	var totalAmount float64
	var totalCost float64

	for _, lineReq := range req.Lines {
		itemID, _ := uuid.Parse(lineReq.ItemID)
		item, err := u.itemRepo.GetByID(ctx, itemID)
		if err != nil {
			return nil, err
		}

		line := invoice.InvoiceLine{
			ID:          uuid.New(),
			InvoiceID:   newInvoice.ID,
			ItemID:      itemID,
			Description: lineReq.Description,
			Quantity:    lineReq.Quantity,
			UnitPrice:   lineReq.UnitPrice,
			UnitCost:    item.CostPrice,
			TotalAmount: lineReq.Quantity * lineReq.UnitPrice,
			TotalCost:   lineReq.Quantity * item.CostPrice,
		}
		totalAmount += line.TotalAmount
		totalCost += line.TotalCost
		newInvoice.Lines = append(newInvoice.Lines, line)
	}

	newInvoice.TotalAmount = totalAmount
	newInvoice.TotalCost = totalCost
	
	newInvoice.SetCreatedBy(userID)
	newInvoice.SetUpdatedBy(userID)

	err := u.invoiceRepo.Create(ctx, newInvoice)
	if err != nil {
		return nil, err
	}

	return newInvoice, nil
}

func (u *usecase) GetByID(ctx context.Context, id uuid.UUID) (*invoice.Invoice, error) {
	return u.invoiceRepo.FindByID(ctx, id)
}
