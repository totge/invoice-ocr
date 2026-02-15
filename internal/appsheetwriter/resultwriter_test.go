package appsheetwriter

import (
	"context"
	"testing"

	"github.com/totge/invoice-oc/go_item_categorizer/internal/appsheet"
	"github.com/totge/invoice-oc/go_item_categorizer/internal/domain"
)

// mockStageWriter captures the expenses passed to WriteExpenseStage.
type mockStageWriter struct {
	captured []appsheet.ExpenseStage
}

func (m *mockStageWriter) WriteExpenseStage(_ context.Context, expenses []appsheet.ExpenseStage) error {
	m.captured = expenses
	return nil
}

func TestWriteResult(t *testing.T) {
	t.Run("single item transformation", func(t *testing.T) {
		mock := &mockStageWriter{}
		w := New(mock)

		receipt := &domain.CategorizedReceipt{
			Timestamp:   "2025-01-15",
			ParsedTotal: 900,
			Items: []domain.CategorizedItem{
				{
					Item: domain.Item{ItemName: "Tej 2.8%", Price: 900, Discount: 0},
					ProductClassification: domain.ProductClassification{
						ProductName:  "Milk",
						CostGroup:    "Dairy",
						MainCategory: "Food",
						SubCategory:  "Dairy Products",
						CategoryID:   "C1",
					},
				},
			},
		}

		err := w.WriteResult(context.Background(), receipt)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(mock.captured) != 1 {
			t.Fatalf("expected 1 expense, got %d", len(mock.captured))
		}

		exp := mock.captured[0]
		if exp.ReceiptId != "2025-01-15 - 900 HUF" {
			t.Errorf("expected ReceiptId=%q, got %q", "2025-01-15 - 900 HUF", exp.ReceiptId)
		}
		if exp.ExpenseDate != "2025-01-15" {
			t.Errorf("expected ExpenseDate=%q, got %q", "2025-01-15", exp.ExpenseDate)
		}
		if exp.CostGroup != "Dairy" {
			t.Errorf("expected CostGroup=%q, got %q", "Dairy", exp.CostGroup)
		}
		if exp.MainCategory != "Food" {
			t.Errorf("expected MainCategory=%q, got %q", "Food", exp.MainCategory)
		}
		if exp.SubCategory != "Dairy Products" {
			t.Errorf("expected SubCategory=%q, got %q", "Dairy Products", exp.SubCategory)
		}
		if exp.Name != "Milk" {
			t.Errorf("expected Name=%q, got %q", "Milk", exp.Name)
		}
		if exp.Amount != 900 {
			t.Errorf("expected Amount=900, got %d", exp.Amount)
		}
		if exp.OriginalName != "Tej 2.8%" {
			t.Errorf("expected OriginalName=%q, got %q", "Tej 2.8%", exp.OriginalName)
		}
		if exp.Approved != false {
			t.Error("expected Approved=false")
		}
	})

	t.Run("multiple items", func(t *testing.T) {
		mock := &mockStageWriter{}
		w := New(mock)

		receipt := &domain.CategorizedReceipt{
			Timestamp:   "2025-01-15",
			ParsedTotal: 2500,
			Items: []domain.CategorizedItem{
				{
					Item: domain.Item{ItemName: "Item A", Price: 1000, Discount: 0},
					ProductClassification: domain.ProductClassification{
						ProductName: "Product A", CostGroup: "G1", MainCategory: "C1", SubCategory: "S1",
					},
				},
				{
					Item: domain.Item{ItemName: "Item B", Price: 1500, Discount: 200},
					ProductClassification: domain.ProductClassification{
						ProductName: "Product B", CostGroup: "G2", MainCategory: "C2", SubCategory: "S2",
					},
				},
			},
		}

		err := w.WriteResult(context.Background(), receipt)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(mock.captured) != 2 {
			t.Fatalf("expected 2 expenses, got %d", len(mock.captured))
		}

		// All items should share the same ReceiptId
		expectedID := "2025-01-15 - 2500 HUF"
		for i, exp := range mock.captured {
			if exp.ReceiptId != expectedID {
				t.Errorf("item[%d]: expected ReceiptId=%q, got %q", i, expectedID, exp.ReceiptId)
			}
			if exp.Approved != false {
				t.Errorf("item[%d]: expected Approved=false", i)
			}
		}

		if mock.captured[0].Name != "Product A" || mock.captured[1].Name != "Product B" {
			t.Errorf("unexpected product names: %q, %q", mock.captured[0].Name, mock.captured[1].Name)
		}
	})
}
