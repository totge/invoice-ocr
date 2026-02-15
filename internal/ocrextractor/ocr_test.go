package ocrextractor

import "testing"

func TestMapToDomain(t *testing.T) {
	// mapToDomain doesn't use the LLM client, so nil is fine.
	op := &OCRPipeline{}

	t.Run("normal discount", func(t *testing.T) {
		raw := &rawReceipt{
			Timestamp:   "2025-01-15T10:30:00",
			ParsedTotal: 900,
			Items: []rawItem{
				{Name: "Milk", FullPrice: 1000, Discount: -100},
			},
		}
		result := op.mapToDomain(raw)

		if result.Timestamp != "2025-01-15T10:30:00" {
			t.Errorf("expected timestamp %q, got %q", "2025-01-15T10:30:00", result.Timestamp)
		}
		if result.ParsedTotal != 900 {
			t.Errorf("expected ParsedTotal=900, got %d", result.ParsedTotal)
		}
		if len(result.Items) != 1 {
			t.Fatalf("expected 1 item, got %d", len(result.Items))
		}

		item := result.Items[0]
		if item.ItemName != "Milk" {
			t.Errorf("expected item name %q, got %q", "Milk", item.ItemName)
		}
		// Paid price = 1000 + (-100) = 900
		if item.Price != 900 {
			t.Errorf("expected price=900, got %d", item.Price)
		}
		// Discount stored as positive: -1 * (-100) = 100
		if item.Discount != 100 {
			t.Errorf("expected discount=100, got %d", item.Discount)
		}
	})

	t.Run("no discount", func(t *testing.T) {
		raw := &rawReceipt{
			Timestamp:   "2025-01-15",
			ParsedTotal: 500,
			Items: []rawItem{
				{Name: "Bread", FullPrice: 500, Discount: 0},
			},
		}
		result := op.mapToDomain(raw)
		item := result.Items[0]

		if item.Price != 500 {
			t.Errorf("expected price=500, got %d", item.Price)
		}
		if item.Discount != 0 {
			t.Errorf("expected discount=0, got %d", item.Discount)
		}
	})

	t.Run("positive discount edge case", func(t *testing.T) {
		raw := &rawReceipt{
			Timestamp:   "2025-01-15",
			ParsedTotal: 1050,
			Items: []rawItem{
				{Name: "Special", FullPrice: 1000, Discount: 50},
			},
		}
		result := op.mapToDomain(raw)
		item := result.Items[0]

		// Paid price = 1000 + 50 = 1050
		if item.Price != 1050 {
			t.Errorf("expected price=1050, got %d", item.Price)
		}
		// Discount stored: -1 * 50 = -50
		if item.Discount != -50 {
			t.Errorf("expected discount=-50, got %d", item.Discount)
		}
	})

	t.Run("multiple items preserved in order", func(t *testing.T) {
		raw := &rawReceipt{
			Timestamp:   "2025-01-15",
			ParsedTotal: 2000,
			Items: []rawItem{
				{Name: "First", FullPrice: 1000, Discount: -100},
				{Name: "Second", FullPrice: 800, Discount: 0},
				{Name: "Third", FullPrice: 300, Discount: -50},
			},
		}
		result := op.mapToDomain(raw)

		if len(result.Items) != 3 {
			t.Fatalf("expected 3 items, got %d", len(result.Items))
		}
		names := []string{result.Items[0].ItemName, result.Items[1].ItemName, result.Items[2].ItemName}
		if names[0] != "First" || names[1] != "Second" || names[2] != "Third" {
			t.Errorf("items out of order: %v", names)
		}
	})

	t.Run("empty items list", func(t *testing.T) {
		raw := &rawReceipt{
			Timestamp:   "2025-01-15",
			ParsedTotal: 0,
			Items:       []rawItem{},
		}
		result := op.mapToDomain(raw)

		if len(result.Items) != 0 {
			t.Errorf("expected 0 items, got %d", len(result.Items))
		}
		if result.Timestamp != "2025-01-15" {
			t.Errorf("expected timestamp to pass through")
		}
	})
}
