package terminalwriter

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"text/tabwriter"
	"time"

	"github.com/totge/invoice-oc/go_item_categorizer/internal/app"
	"github.com/totge/invoice-oc/go_item_categorizer/internal/domain"
)

type Writer struct{}

var _ app.SourceListWriter = (*Writer)(nil)
var _ app.ResultWriter = (*Writer)(nil)

func New() *Writer {
	return &Writer{}
}

// WriteSourceList formats and prints the list of source files to stdout.
func (w *Writer) WriteSourceList(ctx context.Context, sources []domain.SourceInfo) error {
	slog.Debug("Writing source list to terminal", "count", len(sources))

	if len(sources) == 0 {
		fmt.Println("No matching source files found.")
		return nil
	}

	// Initialize tabwriter to write to Stdout.
	// minwidth, tabwidth, padding, padchar, flags
	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)

	// Print Header
	fmt.Fprintln(tw, "FILE NAME\tPATH\tSIZE\tMODIFIED")
	fmt.Fprintln(tw, "---------\t----\t----\t--------")

	// Print Rows
	for _, src := range sources {
		// Format the modification time relative to now or absolute
		modTimeStr := src.ModTime.Format(time.DateTime) // Go 1.20+ format: "2006-01-02 15:04:05"

		// Format size human-readably (simple version)
		sizeStr := fmt.Sprintf("%d B", src.Size)
		if src.Size > 1024 {
			sizeStr = fmt.Sprintf("%.1f KB", float64(src.Size)/1024)
		}

		// Write tab-separated columns
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n",
			src.Name,
			src.Reference,
			sizeStr,
			modTimeStr,
		)
	}

	// Flush ensures everything is written to stdout
	return tw.Flush()
}

// WriteSourceList formats and prints the list of source files to stdout.
func (w *Writer) WriteResult(ctx context.Context, receipt *domain.CategorizedReceipt) error {
	slog.Debug("Writing categorized result to terminal", "item_count", len(receipt.Items))

	fmt.Printf("RECEIPT DATE: %s\n", receipt.Timestamp)
	fmt.Printf("PARSED TOTAL: %d\n\n", receipt.ParsedTotal)

	// Initialize tabwriter to write to Stdout.
	// minwidth, tabwidth, padding, padchar, flags
	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)

	fmt.Fprintln(tw, "ITEM NAME\tPRODUCT NAME\tSUBCATEGORY\tMAIN CATEGORY\tCOST GROUP\tPRICE\tDISCOUNT")
	fmt.Fprintln(tw, "---------\t------------\t-----------\t-------------\t----------\t-----\t--------")

	// Print Rows
	for _, item := range receipt.Items {

		// Write tab-separated columns
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%d\t%d\n",
			item.ItemName,
			item.ProductName,
			item.SubCategory,
			item.MainCategory,
			item.CostGroup,
			item.Price,
			item.Discount,
		)
	}

	// Flush ensures everything is written to stdout
	return tw.Flush()
}
