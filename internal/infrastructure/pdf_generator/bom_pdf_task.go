package pdf_generator

import (
	"context"
	"fmt"
	"log"

	"github.com/johnfercher/maroto/pkg/consts"
	"github.com/johnfercher/maroto/pkg/pdf"
	"github.com/johnfercher/maroto/pkg/props"
)

// BomData represents the data needed to generate a BOM Explosion PDF.
// Placeholder struct - replace with actual domain BOM entity.
type BomData struct {
	BomID         string
	ProductName    string
	Components    []struct {
		ComponentName string
		Quantity      int
		UnitOfMeasure string
	}
}

// BomPDFTask implements the worker.Task interface for generating BOM Explosion PDFs.
type BomPDFTask struct {
	Bom BomData
	OutputPath string // e.g., "boms/bom_explosion_456.pdf"
}

// Execute generates a BOM Explosion PDF using maroto.
func (t *BomPDFTask) Execute(ctx context.Context) error {
	log.Printf("Starting BOM Explosion PDF generation for BOM ID: %s", t.Bom.BomID)

	select {
	case <-ctx.Done():
		log.Printf("BOM Explosion PDF generation for ID %s cancelled due to context termination: %v", t.Bom.BomID, ctx.Err())
		return ctx.Err()
	default:
		m := pdf.NewMaroto(consts.Portrait, consts.A4)
		m.SetPageMargins(10, 10, 10)

		m.RegisterHeader(func() {
			m.Row(10, func() {
				m.Col(12, func() {
					m.Text(fmt.Sprintf("BOM Explosion for Product: %s", t.Bom.ProductName), props.Text{
						Size:  18,
						Align: consts.Center,
						Style: consts.Bold,
					})
				})
			})
		})

		m.Row(15, func() {
			m.Col(12, func() {
				m.Text(fmt.Sprintf("BOM ID: %s", t.Bom.BomID), props.Text{Size: 10, Align: consts.Left})
			})
		})

		m.Line(5)

		headers := []string{"Component Name", "Quantity", "Unit of Measure"}
		contents := [][]string{}
		for _, component := range t.Bom.Components {
			contents = append(contents, []string{
				component.ComponentName,
				fmt.Sprintf("%d", component.Quantity),
				component.UnitOfMeasure,
			})
		}

		m.Row(7, func() {
			m.Col(12, func() {
				m.TableList(headers, contents, props.TableList{
					HeaderProp: props.TableListContent{
						Size:      10,
						GridSizes: []uint{6, 3, 3},
					},
					ContentProp: props.TableListContent{
						Size:      9,
						GridSizes: []uint{6, 3, 3},
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

		// Check for context cancellation before saving
		select {
		case <-ctx.Done():
			log.Printf("BOM Explosion PDF generation for ID %s cancelled during saving due to context termination: %v", t.Bom.BomID, ctx.Err())
			return ctx.Err()
		default:
			err := m.OutputFileAndClose(t.OutputPath)
			if err != nil {
				return fmt.Errorf("could not save BOM Explosion PDF for ID %s: %w", t.Bom.BomID, err)
			}
			log.Printf("Successfully generated BOM Explosion PDF for BOM ID: %s at %s", t.Bom.BomID, t.OutputPath)
			return nil
		}
	}
}