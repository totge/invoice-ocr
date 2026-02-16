package csvcatalog

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeCSV(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "catalog.csv")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestListProducts(t *testing.T) {
	ctx := context.Background()

	t.Run("valid csv", func(t *testing.T) {
		path := writeCSV(t, `product_name,cost_group,main_category,subcategory,category_id
Milk,Dairy,Food,Dairy Products,C1
Bread,Bakery,Food,Bread & Rolls,C2
`)
		catalog := New(path)
		products, err := catalog.ListProducts(ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(products) != 2 {
			t.Fatalf("expected 2 products, got %d", len(products))
		}
	})

	t.Run("deduplication", func(t *testing.T) {
		path := writeCSV(t, `product_name,cost_group,main_category,subcategory,category_id
Milk,Dairy,Food,Dairy Products,C1
Milk,Dairy,Food,Dairy Products,C1
Bread,Bakery,Food,Bread & Rolls,C2
`)
		catalog := New(path)
		products, err := catalog.ListProducts(ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(products) != 2 {
			t.Errorf("expected 2 products after dedup, got %d", len(products))
		}
	})

	t.Run("same product different categories", func(t *testing.T) {
		path := writeCSV(t, `product_name,cost_group,main_category,subcategory,category_id
Milk,Dairy,Food,Dairy Products,C1
Milk,Beverages,Drinks,Cold Drinks,C5
`)
		catalog := New(path)
		products, err := catalog.ListProducts(ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(products) != 2 {
			t.Errorf("expected 2 products (different category_id), got %d", len(products))
		}
	})

	t.Run("empty csv header only", func(t *testing.T) {
		path := writeCSV(t, "product_name,cost_group,main_category,subcategory,category_id\n")
		catalog := New(path)
		products, err := catalog.ListProducts(ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(products) != 0 {
			t.Errorf("expected 0 products, got %d", len(products))
		}
	})

	t.Run("missing header", func(t *testing.T) {
		path := writeCSV(t, "")
		catalog := New(path)
		_, err := catalog.ListProducts(ctx)
		if err == nil {
			t.Fatal("expected error for empty file")
		}
		if !strings.Contains(err.Error(), "missing header") {
			t.Errorf("expected error about missing header, got: %v", err)
		}
	})

	t.Run("wrong header columns", func(t *testing.T) {
		path := writeCSV(t, "name,group,category,sub,id\nMilk,Dairy,Food,Dairy Products,C1\n")
		catalog := New(path)
		_, err := catalog.ListProducts(ctx)
		if err == nil {
			t.Fatal("expected error for wrong header")
		}
		if !strings.Contains(err.Error(), "expected") {
			t.Errorf("expected error about header mismatch, got: %v", err)
		}
	})

	t.Run("rows with empty fields skipped", func(t *testing.T) {
		path := writeCSV(t, `product_name,cost_group,main_category,subcategory,category_id
Milk,Dairy,Food,Dairy Products,C1
,,,,
Bread,Bakery,Food,,C2
`)
		catalog := New(path)
		products, err := catalog.ListProducts(ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(products) != 1 {
			t.Errorf("expected 1 product (2 rows skipped), got %d", len(products))
		}
	})

	t.Run("file does not exist", func(t *testing.T) {
		catalog := New("/nonexistent/catalog.csv")
		_, err := catalog.ListProducts(ctx)
		if err == nil {
			t.Fatal("expected error for nonexistent file")
		}
	})

	t.Run("whitespace trimmed", func(t *testing.T) {
		path := writeCSV(t, `product_name,cost_group,main_category,subcategory,category_id
 Milk , Dairy , Food , Dairy Products , C1
`)
		catalog := New(path)
		products, err := catalog.ListProducts(ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(products) != 1 {
			t.Fatalf("expected 1 product, got %d", len(products))
		}
		if products[0].ProductName != "Milk" {
			t.Errorf("expected trimmed product name %q, got %q", "Milk", products[0].ProductName)
		}
	})
}
