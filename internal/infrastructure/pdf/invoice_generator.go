package pdf

import (
	"context"
	"fmt"
	"doligo_001/internal/domain/invoice"
	"github.com/johnfercher/maroto/pkg/consts"
	"github.com/johnfercher/maroto/pkg/pdf"
	"github.com/johnfercher/maroto/pkg/props"
)

// Generator defines the interface for a PDF document generator.
type Generator interface {
	Generate(ctx context.Context, invoice *invoice.Invoice) ([]byte, error)
}

// marotoGenerator is an implementation of Generator that uses the Maroto library.
type marotoGenerator struct{}

// NewMarotoGenerator creates a new instance of a Maroto-based PDF generator.
func NewMarotoGenerator() Generator {
	return &marotoGenerator{}
}

// Generate creates a PDF for a given invoice and returns its content as a byte slice.
func (g *marotoGenerator) Generate(ctx context.Context, inv *invoice.Invoice) ([]byte, error) {
	m := pdf.NewMaroto(consts.Portrait, consts.A4)
	m.SetPageMargins(10, 15, 10)

	if err := g.buildHeader(m, inv); err != nil {
		return nil, err
	}
	// Pass context to buildBody
	if err := g.buildBody(ctx, m, inv); err != nil {
		return nil, err
	}
	g.buildFooter(m)

	// Check for cancellation before the final, potentially expensive, output step.
	select {
	case <-ctx.Done():
		return nil, ctx.Err() // Return context error if cancelled
	default:
		// continue
	}

	// Return the PDF as a byte array instead of writing to a file.
	bytes, err := m.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to generate PDF bytes: %w", err)
	}

	return bytes.Bytes(), nil
}

func (g *marotoGenerator) buildHeader(m pdf.Maroto, inv *invoice.Invoice) error {
	m.Row(20, func() {
		m.Col(6, func() {
			m.Text("INVOICE", props.Text{
				Top:   3,
				Style: consts.Bold,
				Size:  24,
				Align: consts.Left,
			})
		})
		m.Col(6, func() {
			m.Text(fmt.Sprintf("Invoice #%s", inv.Number), props.Text{Top: 5, Align: consts.Right})
			m.Text(fmt.Sprintf("Date: %s", inv.Date.Format("2006-01-02")), props.Text{Top: 10, Align: consts.Right})
		})
	})
	return nil
}

// buildBody now accepts a context to allow for cancellation.
func (g *marotoGenerator) buildBody(ctx context.Context, m pdf.Maroto, inv *invoice.Invoice) error {
	m.Line(10)

	// Customer Information
	m.Row(12, func() {
		m.Col(12, func() {
			m.Text("Bill To:", props.Text{Style: consts.Bold})
			if inv.ThirdParty != nil {
				m.Text(inv.ThirdParty.Name, props.Text{Top: 5})
				m.Text(inv.ThirdParty.Email, props.Text{Top: 10})
			} else {
				m.Text("N/A", props.Text{Top: 5})
			}
		})
	})

	m.Line(10)

	// Invoice Lines Table
	headers := []string{"Description", "Quantity", "Unit Price", "Total"}
	var contents [][]string
	for _, line := range inv.Lines {
		// Check for cancellation on each line item. This is crucial for large invoices.
		select {
		case <-ctx.Done():
			return ctx.Err() // Return context error if cancelled
		default:
			// continue processing
		}

		contents = append(contents, []string{
			line.Description,
			fmt.Sprintf("%.2f", line.Quantity),
			fmt.Sprintf("%.2f", line.UnitPrice),
			fmt.Sprintf("%.2f", line.TotalAmount),
		})
	}

	m.TableList(headers, contents, props.TableList{
		HeaderProp: props.TableListContent{
			Size:      9,
			GridSizes: []uint{6, 2, 2, 2},
		},
		ContentProp: props.TableListContent{
			Size:      9,
			GridSizes: []uint{6, 2, 2, 2},
		},
		Align: consts.Center,
		HeaderContentSpace: 1,
		LineProp: props.Line{
			Style: consts.Dotted,
			Width: 0.5,
		},
	})

	m.Row(20, func() {
		m.ColSpace(7)
		m.Col(5, func() {
			m.Text(fmt.Sprintf("Total: %.2f", inv.TotalAmount), props.Text{
				Top:   5,
				Size:  12,
				Style: consts.Bold,
				Align: consts.Right,
			})
		})
	})

	return nil
}

func (g *marotoGenerator) buildFooter(m pdf.Maroto) {
	m.RegisterFooter(func() {
		m.Row(10, func() {
			m.Col(12, func() {
				m.Text("Thank you for your business.", props.Text{
					Top:   5,
					Size:  8,
					Align: consts.Center,
					Style: consts.Italic,
				})
			})
		})
	})
}
