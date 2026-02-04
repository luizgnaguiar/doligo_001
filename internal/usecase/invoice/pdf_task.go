package invoice

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

type InvoicePDFTask struct {
	InvoiceID uuid.UUID
	Usecase   *usecase
}

func (t *InvoicePDFTask) Execute(ctx context.Context) error {
	// 1. Fetch the invoice
	inv, err := t.Usecase.invoiceRepo.FindByIDWithDetails(ctx, t.InvoiceID)
	if err != nil {
		return fmt.Errorf("failed to fetch invoice %s: %w", t.InvoiceID, err)
	}

	// 2. Generate PDF
	pdfBytes, err := t.Usecase.pdfGen.Generate(ctx, inv)
	if err != nil {
		// Update status to failed
		inv.PDFStatus = "failed"
		_ = t.Usecase.invoiceRepo.Update(ctx, inv)
		return fmt.Errorf("failed to generate PDF for invoice %s: %w", t.InvoiceID, err)
	}

	// 3. Save PDF locally
	storageDir := "storage/pdfs"
	if err := os.MkdirAll(storageDir, 0755); err != nil {
		return fmt.Errorf("failed to create storage directory: %w", err)
	}

	filename := fmt.Sprintf("invoice-%s.pdf", inv.Number)
	filePath := filepath.Join(storageDir, filename)
	if err := os.WriteFile(filePath, pdfBytes, 0644); err != nil {
		return fmt.Errorf("failed to write PDF file: %w", err)
	}

	// 4. Update Invoice status and URL
	inv.PDFStatus = "completed"
	inv.PDFUrl = filePath
	if err := t.Usecase.invoiceRepo.Update(ctx, inv); err != nil {
		return fmt.Errorf("failed to update invoice status: %w", err)
	}

	return nil
}
