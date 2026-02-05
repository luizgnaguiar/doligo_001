package invoice

import (
	"context"
	"doligo_001/internal/domain/invoice"
	"doligo_001/internal/infrastructure/email"
	"fmt"
	"time"

	"doligo_001/internal/api/dto"
	"doligo_001/internal/api/middleware"
	"doligo_001/internal/domain"
	"doligo_001/internal/infrastructure/pdf"
	"doligo_001/internal/infrastructure/worker"
	audit_uc "doligo_001/internal/usecase"
	item_usecase "doligo_001/internal/usecase/item"

	"github.com/google/uuid"
)

type usecase struct {
	invoiceRepo    Repository
	itemRepo       item_usecase.Repository
	pdfGen         pdf.Generator
	emailSender    email.EmailSender
	workerPool     *worker.WorkerPool
	auditService   audit_uc.AuditService
	pdfStoragePath string
}

func NewUsecase(invoiceRepo Repository, itemRepo item_usecase.Repository, pdfGen pdf.Generator, emailSender email.EmailSender, workerPool *worker.WorkerPool, auditService audit_uc.AuditService, pdfStoragePath string) Usecase {
	return &usecase{
		invoiceRepo:    invoiceRepo,
		itemRepo:       itemRepo,
		pdfGen:         pdfGen,
		emailSender:    emailSender,
		workerPool:     workerPool,
		auditService:   auditService,
		pdfStoragePath: pdfStoragePath,
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
			CreatedBy:   userID,
			UpdatedBy:   userID,
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
		inv.PDFErrorMessage = err.Error()
		_ = u.invoiceRepo.Update(ctx, inv)
		return fmt.Errorf("failed to submit PDF task: %w", err)
	}

	return nil
}

func (u *usecase) GetPDFStatus(ctx context.Context, id uuid.UUID) (*dto.InvoicePDFStatusResponse, error) {
	inv, err := u.invoiceRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	userID, _ := domain.UserIDFromContext(ctx)
	permissions, _ := domain.PermissionsFromContext(ctx)

	isOwner := inv.CreatedBy == userID
	hasPermission := false
	for _, p := range permissions {
		if p == "INVOICE_READ" {
			hasPermission = true
			break
		}
	}

	if !isOwner && !hasPermission {
		return nil, fmt.Errorf("forbidden: you do not have permission to access this invoice")
	}

	var relativeURL string
	if inv.PDFStatus == "completed" {
		relativeURL = fmt.Sprintf("/api/v1/invoices/%s/pdf", id)
	}

	return &dto.InvoicePDFStatusResponse{
		Status:       inv.PDFStatus,
		PDFUrl:       relativeURL,
		ErrorMessage: inv.PDFErrorMessage,
	}, nil
}

func (u *usecase) GetPDFPath(ctx context.Context, id uuid.UUID) (string, error) {
	inv, err := u.invoiceRepo.FindByID(ctx, id)
	if err != nil {
		return "", err
	}

	userID, _ := domain.UserIDFromContext(ctx)
	permissions, _ := domain.PermissionsFromContext(ctx)

	isOwner := inv.CreatedBy == userID
	hasPermission := false
	for _, p := range permissions {
		if p == "INVOICE_READ" {
			hasPermission = true
			break
		}
	}

	if !isOwner && !hasPermission {
		return "", fmt.Errorf("forbidden: you do not have permission to access this invoice")
	}

	if inv.PDFStatus != "completed" {
		return "", fmt.Errorf("PDF not ready")
	}

	return inv.PDFUrl, nil
}

func (u *usecase) Delete(ctx context.Context, id uuid.UUID) error {
	userID, _ := domain.UserIDFromContext(ctx)
	
	oldInvoice, err := u.invoiceRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if err := u.invoiceRepo.Delete(ctx, id); err != nil {
		return err
	}

	corrID, _ := middleware.FromContext(ctx)
	u.auditService.Log(ctx, userID, "invoice", id.String(), "DELETE", oldInvoice, nil, corrID)

	return nil
}
