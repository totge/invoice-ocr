package csvcatalog

import (
	"context"
	"encoding/csv"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/totge/invoice-oc/go_item_categorizer/internal/app"
	"github.com/totge/invoice-oc/go_item_categorizer/internal/domain"
)

var expectedHeader = []string{"product_name", "cost_group", "main_category", "subcategory", "category_id"}

// Catalog implements app.ProductLister by reading products from a CSV file.
type Catalog struct {
	filePath string
}

var _ app.ProductLister = (*Catalog)(nil)

// New creates a new CSV-based product catalog.
func New(filePath string) *Catalog {
	return &Catalog{filePath: filePath}
}

// ListProducts reads the CSV file and returns deduplicated product classifications.
func (c *Catalog) ListProducts(_ context.Context) ([]domain.ProductClassification, error) {
	slog.Debug("Loading product catalog from CSV", "path", c.filePath)

	file, err := os.Open(c.filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open catalog CSV: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to parse catalog CSV: %w", err)
	}

	if len(records) == 0 {
		return nil, fmt.Errorf("catalog CSV is empty (missing header)")
	}

	// Validate header
	header := records[0]
	if err := validateHeader(header); err != nil {
		return nil, err
	}

	// Parse rows with deduplication
	type productKey struct {
		ProductName string
		CategoryID  string
	}
	unique := make(map[productKey]domain.ProductClassification)

	for i, row := range records[1:] {
		lineNum := i + 2 // 1-indexed, skip header

		if len(row) < 5 {
			slog.Warn("Skipping CSV row with insufficient columns", "line", lineNum)
			continue
		}

		productName := strings.TrimSpace(row[0])
		costGroup := strings.TrimSpace(row[1])
		mainCategory := strings.TrimSpace(row[2])
		subCategory := strings.TrimSpace(row[3])
		categoryID := strings.TrimSpace(row[4])

		if productName == "" || costGroup == "" || mainCategory == "" || subCategory == "" || categoryID == "" {
			slog.Warn("Skipping CSV row with empty required fields", "line", lineNum)
			continue
		}

		key := productKey{ProductName: productName, CategoryID: categoryID}
		if _, exists := unique[key]; !exists {
			unique[key] = domain.ProductClassification{
				ProductName:  productName,
				CostGroup:    costGroup,
				MainCategory: mainCategory,
				SubCategory:  subCategory,
				CategoryID:   categoryID,
			}
		}
	}

	products := make([]domain.ProductClassification, 0, len(unique))
	for _, p := range unique {
		products = append(products, p)
	}

	slog.Debug("CSV catalog loaded", "product_count", len(products))
	return products, nil
}

func validateHeader(header []string) error {
	if len(header) < len(expectedHeader) {
		return fmt.Errorf("catalog CSV header must have %d columns (%s), got %d",
			len(expectedHeader), strings.Join(expectedHeader, ", "), len(header))
	}
	for i, expected := range expectedHeader {
		got := strings.TrimSpace(strings.ToLower(header[i]))
		if got != expected {
			return fmt.Errorf("catalog CSV header column %d: expected %q, got %q", i+1, expected, header[i])
		}
	}
	return nil
}
