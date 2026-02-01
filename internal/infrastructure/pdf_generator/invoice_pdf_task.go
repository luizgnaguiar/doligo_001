package pdf_generator

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/johnfercher/maroto/pkg/consts"
	"github.com/johnfercher/maroto/pkg/pdf"
	"github.com/johnfercher/maroto/pkg/props"
)

// InvoiceData represents the data needed to generate an invoice PDF.
// Placeholder struct - replace with actual domain Invoice entity.
type InvoiceData struct {
	InvoiceID   string
	CustomerName string
	Items       []struct {
		Description string
		Quantity    int
		UnitPrice   float64
		Total       float64
	}
	TotalAmount float64
}

// InvoicePDFTask implements the worker.Task interface for generating invoice PDFs.
type InvoicePDFTask struct {
	Invoice InvoiceData
	OutputPath string // e.g., "invoices/invoice_123.pdf"
}

// Execute generates an invoice PDF using maroto.
func (t *InvoicePDFTask) Execute(ctx context.Context) error {
	log.Printf("Starting Invoice PDF generation for Invoice ID: %s", t.Invoice.InvoiceID)

	select {
	case <-ctx.Done():
		log.Printf("Invoice PDF generation for ID %s cancelled due to context termination: %v", t.Invoice.InvoiceID, ctx.Err())
		return ctx.Err()
	default:
		m := pdf.NewMaroto(consts.Portrait, consts.A4)
		m.SetPageMargins(10, 10, 10)

		m.RegisterHeader(func() {
			m.Row(10, func() {
				m.Col(12, func() {
					m.Text(fmt.Sprintf("Invoice #%s", t.Invoice.InvoiceID), props.Text{
						Size:  18,
						Align: consts.Center,
						Style: consts.Bold,
					})
				})
			})
		})

		m.Row(15, func() {
			m.Col(6, func() {
				m.Text("Customer:", props.Text{Size: 10, Align: consts.Left})
				m.Text(t.Invoice.CustomerName, props.Text{Size: 12, Align: consts.Left, Style: consts.Bold})
			})
			m.Col(6, func() {
				m.Text(fmt.Sprintf("Date: %s", time.Now().Format("02/01/2006")), props.Text{Size: 10, Align: consts.Right})
			})
		})

		m.Line(5)

		headers := []string{"Description", "Quantity", "Unit Price", "Total"}
		contents := [][]string{}
		for _, item := range t.Invoice.Items {
			contents = append(contents, []string{
				item.Description,
				fmt.Sprintf("%d", item.Quantity),
				fmt.Sprintf("%.2f", item.UnitPrice),
				fmt.Sprintf("%.2f", item.Total),
			})
		}

		m.Row(7, func() {
			m.Col(12, func() {
				m.TableList(headers, contents, props.TableList{
					HeaderProp: props.TableListContent{
						Size:      10,
						GridSizes: []uint{6, 2, 2, 2},
					},
					ContentProp: props.TableListContent{
						Size:      9,
						GridSizes: []uint{6, 2, 2, 2},
					},
					Align: consts.Center,
					AlternatedBackground: &props.Color{
						Red:   240,
						Green: 240,
						Blue:  240,
					},
				})
			})
		})

		m.Row(10, func() {
			m.Col(12, func() {
				m.Text(fmt.Sprintf("Total Amount: %.2f", t.Invoice.TotalAmount), props.Text{
					Size:  14,
					Align: consts.Right,
					Style: consts.Bold,
				})
			})
		})

		// Check for context cancellation before saving
		select {
		case <-ctx.Done():
			log.Printf("Invoice PDF generation for ID %s cancelled during saving due to context termination: %v", t.Invoice.InvoiceID, ctx.Err())
			return ctx.Err()
		default:
			err := m.OutputFileAndClose(t.OutputPath)
			if err != nil {
				return fmt.Errorf("could not save Invoice PDF for ID %s: %w", t.Invoice.InvoiceID, err)
			}
			log.Printf("Successfully generated Invoice PDF for Invoice ID: %s at %s", t.Invoice.InvoiceID, t.OutputPath)
			return nil
		}
	}
}