package llmcategorizer

import (
	"sort"
	"testing"

	"github.com/totge/invoice-oc/go_item_categorizer/internal/domain"
)

func TestPrepareProductData(t *testing.T) {
	c := &Categorizer{}

	t.Run("groups by cost group", func(t *testing.T) {
		products := []domain.ProductClassification{
			{ProductName: "Milk", CostGroup: "Dairy"},
			{ProductName: "Cheese", CostGroup: "Dairy"},
			{ProductName: "Apple", CostGroup: "Fruit"},
		}
		byGroup, groups := c.prepareProductData(products)

		if len(byGroup) != 2 {
			t.Fatalf("expected 2 groups, got %d", len(byGroup))
		}
		if len(byGroup["Dairy"]) != 2 {
			t.Errorf("expected 2 Dairy products, got %d", len(byGroup["Dairy"]))
		}
		if len(byGroup["Fruit"]) != 1 {
			t.Errorf("expected 1 Fruit product, got %d", len(byGroup["Fruit"]))
		}

		sort.Strings(groups)
		if len(groups) != 2 || groups[0] != "Dairy" || groups[1] != "Fruit" {
			t.Errorf("expected [Dairy, Fruit], got %v", groups)
		}
	})

	t.Run("empty product list", func(t *testing.T) {
		byGroup, groups := c.prepareProductData(nil)
		if len(byGroup) != 0 {
			t.Errorf("expected empty map, got %d entries", len(byGroup))
		}
		if len(groups) != 0 {
			t.Errorf("expected empty groups, got %d", len(groups))
		}
	})

	t.Run("single cost group", func(t *testing.T) {
		products := []domain.ProductClassification{
			{ProductName: "A", CostGroup: "Food"},
			{ProductName: "B", CostGroup: "Food"},
		}
		byGroup, groups := c.prepareProductData(products)

		if len(byGroup) != 1 {
			t.Fatalf("expected 1 group, got %d", len(byGroup))
		}
		if len(groups) != 1 || groups[0] != "Food" {
			t.Errorf("expected [Food], got %v", groups)
		}
	})
}

func TestBuildFinalReceipt(t *testing.T) {
	c := &Categorizer{}

	t.Run("all items mapped", func(t *testing.T) {
		receipt := &domain.Receipt{
			Timestamp:   "2025-01-15",
			ParsedTotal: 1500,
			Items: []domain.Item{
				{ItemName: "Milk", Price: 500, Discount: 0},
				{ItemName: "Bread", Price: 1000, Discount: 100},
			},
		}
		mapping := map[string]stage2ResponseItem{
			"Milk": {
				CategoryID:   "C1",
				ProductName:  "Full Fat Milk",
				CostGroup:    "Dairy",
				MainCategory: "Food",
				Subcategory:  "Dairy Products",
			},
			"Bread": {
				CategoryID:   "C2",
				ProductName:  "White Bread",
				CostGroup:    "Bakery",
				MainCategory: "Food",
				Subcategory:  "Bread & Rolls",
			},
		}

		result := c.buildFinalReceipt(receipt, mapping)

		if result.Timestamp != "2025-01-15" {
			t.Errorf("expected timestamp pass-through, got %q", result.Timestamp)
		}
		if result.ParsedTotal != 1500 {
			t.Errorf("expected ParsedTotal=1500, got %d", result.ParsedTotal)
		}
		if len(result.Items) != 2 {
			t.Fatalf("expected 2 items, got %d", len(result.Items))
		}

		item0 := result.Items[0]
		if item0.ItemName != "Milk" || item0.ProductName != "Full Fat Milk" || item0.CostGroup != "Dairy" {
			t.Errorf("unexpected item[0]: %+v", item0)
		}
		if item0.Price != 500 || item0.Discount != 0 {
			t.Errorf("unexpected item[0] pricing: price=%d, discount=%d", item0.Price, item0.Discount)
		}

		item1 := result.Items[1]
		if item1.ItemName != "Bread" || item1.ProductName != "White Bread" {
			t.Errorf("unexpected item[1]: %+v", item1)
		}
	})

	t.Run("unmapped items get uncategorized defaults", func(t *testing.T) {
		receipt := &domain.Receipt{
			Timestamp:   "2025-01-15",
			ParsedTotal: 500,
			Items: []domain.Item{
				{ItemName: "Mystery Item", Price: 500, Discount: 0},
			},
		}
		mapping := map[string]stage2ResponseItem{} // empty mapping

		result := c.buildFinalReceipt(receipt, mapping)

		if len(result.Items) != 1 {
			t.Fatalf("expected 1 item, got %d", len(result.Items))
		}
		item := result.Items[0]
		if item.ProductName != "Uncategorized" {
			t.Errorf("expected ProductName=%q, got %q", "Uncategorized", item.ProductName)
		}
		if item.CostGroup != "Uncategorized" {
			t.Errorf("expected CostGroup=%q, got %q", "Uncategorized", item.CostGroup)
		}
		if item.MainCategory != "Uncategorized" {
			t.Errorf("expected MainCategory=%q, got %q", "Uncategorized", item.MainCategory)
		}
		// Original item data should still be preserved
		if item.ItemName != "Mystery Item" || item.Price != 500 {
			t.Errorf("original item data not preserved: %+v", item)
		}
	})

	t.Run("partial mapping", func(t *testing.T) {
		receipt := &domain.Receipt{
			Timestamp:   "2025-01-15",
			ParsedTotal: 1000,
			Items: []domain.Item{
				{ItemName: "Known", Price: 500, Discount: 0},
				{ItemName: "Unknown", Price: 500, Discount: 0},
			},
		}
		mapping := map[string]stage2ResponseItem{
			"Known": {
				CategoryID:   "C1",
				ProductName:  "Known Product",
				CostGroup:    "Group",
				MainCategory: "Cat",
				Subcategory:  "Sub",
			},
		}

		result := c.buildFinalReceipt(receipt, mapping)

		if result.Items[0].ProductName != "Known Product" {
			t.Errorf("expected first item mapped, got %q", result.Items[0].ProductName)
		}
		if result.Items[1].ProductName != "Uncategorized" {
			t.Errorf("expected second item uncategorized, got %q", result.Items[1].ProductName)
		}
	})

	t.Run("empty receipt", func(t *testing.T) {
		receipt := &domain.Receipt{
			Timestamp:   "2025-01-15",
			ParsedTotal: 0,
			Items:       []domain.Item{},
		}
		result := c.buildFinalReceipt(receipt, map[string]stage2ResponseItem{})

		if len(result.Items) != 0 {
			t.Errorf("expected 0 items, got %d", len(result.Items))
		}
		if result.Timestamp != "2025-01-15" {
			t.Errorf("expected timestamp preserved")
		}
	})
}
