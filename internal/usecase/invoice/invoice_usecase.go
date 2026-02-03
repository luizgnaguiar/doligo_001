package invoice

import (
	"context"
	"doligo_001/internal/domain/invoice"
	"doligo_001/internal/infrastructure/email"
	"fmt"
	"time"

	"doligo_001/internal/api/dto"
	"doligo_001/internal/domain"
	"doligo_001/internal/infrastructure/pdf"
	item_usecase "doligo_001/internal/usecase/item"

	"github.com/google/uuid"
)

type usecase struct {
	invoiceRepo Repository
	itemRepo    item_usecase.Repository
	pdfGen      pdf.Generator
	emailSender email.EmailSender
}

func NewUsecase(invoiceRepo Repository, itemRepo item_usecase.Repository, pdfGen pdf.Generator, emailSender email.EmailSender) Usecase {
	return &usecase{
		invoiceRepo: invoiceRepo,
		itemRepo:    itemRepo,
		pdfGen:      pdfGen,
		emailSender: emailSender,
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

	// Send email
	go func() {
		emailCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		u.emailSender.Send(emailCtx, "test@example.com", "New Invoice Created", fmt.Sprintf("Invoice %s has been created.", newInvoice.Number))
	}()

	return newInvoice, nil
}

func (u *usecase) GetByID(ctx context.Context, id uuid.UUID) (*invoice.Invoice, error) {
	return u.invoiceRepo.FindByID(ctx, id)
}

func (u *usecase) GenerateInvoicePDF(ctx context.Context, invoiceID uuid.UUID) ([]byte, string, error) {
	// 1. Fetch the full invoice data
	inv, err := u.invoiceRepo.FindByIDWithDetails(ctx, invoiceID)
	if err != nil {
		return nil, "", fmt.Errorf("failed to find invoice: %w", err)
	}

	// 2. Generate the PDF
	pdfBytes, err := u.pdfGen.Generate(ctx, inv)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate PDF: %w", err)
	}

	// 3. Create a filename
	filename := fmt.Sprintf("invoice-%s.pdf", inv.Number)

	return pdfBytes, filename, nil
}
