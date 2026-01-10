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
		sizeStr := w.formatSize(src.Size)

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

func (w *Writer) formatSize(bytes int64) string {
	const unit = 1024.0 // Use 1000.0 if you prefer SI units
	if bytes < int64(unit) {
		return fmt.Sprintf("%d B", bytes)
	}

	div := unit
	exp := 0
	// Loop until the number is small enough, or we run out of units
	for n := bytes / int64(unit); n >= int64(unit); n /= int64(unit) {
		div *= unit
		exp++
	}

	// 'K' is index 0 in the suffix list for 1024^1
	suffixes := []string{"KB", "MB", "GB", "TB"}

	// Safety check for huge numbers
	if exp >= len(suffixes) {
		exp = len(suffixes) - 1
	}

	return fmt.Sprintf("%.1f %s", float64(bytes)/float64(div), suffixes[exp])
}
