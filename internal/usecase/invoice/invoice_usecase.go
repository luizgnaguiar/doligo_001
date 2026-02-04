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
	"doligo_001/internal/infrastructure/worker"
	item_usecase "doligo_001/internal/usecase/item"

	"github.com/google/uuid"
)

type usecase struct {
	invoiceRepo Repository
	itemRepo    item_usecase.Repository
	pdfGen      pdf.Generator
	emailSender email.EmailSender
	workerPool  *worker.WorkerPool
}

func NewUsecase(invoiceRepo Repository, itemRepo item_usecase.Repository, pdfGen pdf.Generator, emailSender email.EmailSender, workerPool *worker.WorkerPool) Usecase {
	return &usecase{
		invoiceRepo: invoiceRepo,
		itemRepo:    itemRepo,
		pdfGen:      pdfGen,
		emailSender: emailSender,
		workerPool:  workerPool,
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
	var totalTax float64

	for _, lineReq := range req.Lines {
		itemID, _ := uuid.Parse(lineReq.ItemID)
		item, err := u.itemRepo.GetByID(ctx, itemID)
		if err != nil {
			return nil, err
		}

		// Calculate tax (assuming TaxRate is percentage, e.g. 10 for 10%)
		taxAmount := lineReq.UnitPrice * (lineReq.TaxRate / 100)
		netPrice := lineReq.UnitPrice + taxAmount
		lineTotalAmount := lineReq.Quantity * netPrice
		lineTotalTax := lineReq.Quantity * taxAmount

		line := invoice.InvoiceLine{
			ID:          uuid.New(),
			InvoiceID:   newInvoice.ID,
			ItemID:      itemID,
			Description: lineReq.Description,
			Quantity:    lineReq.Quantity,
			UnitPrice:   lineReq.UnitPrice,
			UnitCost:    item.CostPrice,
			TaxRate:     lineReq.TaxRate,
			TaxAmount:   taxAmount,
			NetPrice:    netPrice,
			TotalAmount: lineTotalAmount,
			TotalCost:   lineReq.Quantity * item.CostPrice,
		}
		totalAmount += line.TotalAmount
		totalCost += line.TotalCost
		totalTax += lineTotalTax
		newInvoice.Lines = append(newInvoice.Lines, line)
	}

	newInvoice.TotalAmount = totalAmount
	newInvoice.TotalCost = totalCost
	newInvoice.TotalTax = totalTax

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

func (u *usecase) QueueInvoicePDFGeneration(ctx context.Context, invoiceID uuid.UUID) error {
	// 1. Fetch the invoice to ensure it exists
	inv, err := u.invoiceRepo.FindByID(ctx, invoiceID)
	if err != nil {
		return fmt.Errorf("failed to find invoice: %w", err)
	}

	// 2. Update status to processing
	inv.PDFStatus = "processing"
	if err := u.invoiceRepo.Update(ctx, inv); err != nil {
		return fmt.Errorf("failed to update invoice status: %w", err)
	}

	// 3. Submit to worker pool
	task := &InvoicePDFTask{
		InvoiceID: invoiceID,
		Usecase:   u,
	}

	if err := u.workerPool.Submit(task); err != nil {
		// If submission fails, try to revert status or mark as failed
		inv.PDFStatus = "failed"
		_ = u.invoiceRepo.Update(ctx, inv)
		return fmt.Errorf("failed to submit PDF task: %w", err)
	}

	return nil
}
