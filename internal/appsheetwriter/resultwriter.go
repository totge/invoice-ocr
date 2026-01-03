package appsheetwriter

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/totge/invoice-oc/go_item_categorizer/internal/app"
	"github.com/totge/invoice-oc/go_item_categorizer/internal/appsheet"
	"github.com/totge/invoice-oc/go_item_categorizer/internal/domain"
)

type StageWriter interface {
	WriteExpenseStage(context.Context, []appsheet.ExpenseStage) error
}

type Writer struct {
	client StageWriter
}

var _ app.ResultWriter = (*Writer)(nil)

func (w *Writer) WriteResult(ctx context.Context, receipt *domain.CategorizedReceipt) error {

	slog.Debug("Transforming receipt for AppSheet export",
		"receipt_timestamp", receipt.Timestamp,
		"total_amount", receipt.ParsedTotal,
		"item_count", len(receipt.Items),
	)

	stagedExpenses := make([]appsheet.ExpenseStage, 0, len(receipt.Items))

	for _, item := range receipt.Items {
		expense := appsheet.ExpenseStage{
			ReceiptId:    receipt.Timestamp + " - " + strconv.Itoa(receipt.ParsedTotal) + " HUF",
			ExpenseDate:  receipt.Timestamp,
			CostGroup:    item.CostGroup,
			MainCategory: item.MainCategory,
			SubCategory:  item.SubCategory,
			Name:         item.ProductName,
			Amount:       item.Price,
			OriginalName: item.ItemName,
			Approved:     false,
		}
		stagedExpenses = append(stagedExpenses, expense)
	}

	err := w.client.WriteExpenseStage(ctx, stagedExpenses)
	if err != nil {
		return fmt.Errorf("appsheet client failed to write the results: %w", err)
	}

	slog.Debug("Receipt successfully exported to AppSheet")

	return nil
}

func New(writer StageWriter) *Writer {
	return &Writer{client: writer}
}
