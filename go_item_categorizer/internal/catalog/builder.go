package catalog

import (
	"context"
	"log"
	"strings"

	"github.com/totge/invoice-oc/go_item_categorizer/internal/appsheet"
)

type ProductClassification struct {
	ProductName  string
	CostGroup    string
	MainCategory string
	SubCategory  string
	CategoryID   string
}

type ProductCatalog map[string][]ProductClassification

func (p *ProductCatalog) GetCostGroups() []string {
	costGroups := make([]string, len(*p))

	for k := range *p {
		costGroups = append(costGroups, k)
	}

	return costGroups
}

func BuildProductCatalog(ctx context.Context, client *appsheet.Client) (ProductCatalog, error) {

	// --- 2. Fetch Categories from AppSheet ---
	log.Printf("Fetching categories from table '%s'...", string(appsheet.TableCategories))
	// Prepare the data structure to hold the results
	var categoryData appsheet.AppSheetTableData[appsheet.Category]

	// Call the client method, passing the specific table name and the data holder
	err := client.ReadRecords(ctx, appsheet.TableCategories, &categoryData)
	if err != nil {
		// This error includes request errors OR the decoding error IF you fixed ReadRecords
		log.Fatalf("FATAL: Failed to fetch categories: %v", err)
	}
	// Access the data via the struct field
	categories := categoryData.Data

	categoryMap := make(map[string]appsheet.Category)
	for _, cat := range categories {
		categoryMap[cat.CategoryId] = cat
	}

	log.Printf("Fetching expenses from table '%s'...", string(appsheet.TableExpenses))
	// Prepare the data structure to hold the results
	var expenseData appsheet.AppSheetTableData[appsheet.Expense]

	// Call the client method, passing the specific table name and the data holder
	err = client.ReadRecords(ctx, appsheet.TableExpenses, &expenseData)
	if err != nil {
		// This error includes request errors OR the decoding error IF you fixed ReadRecords
		log.Fatalf("FATAL: Failed to fetch expenses: %v", err)
	}
	// Access the data via the struct field
	expenses := expenseData.Data

	// --- 3. Deduplicate Expenses by ProductName + CategoryID ---
	log.Println("Deduplicating expenses by product name and category ID...")
	// Use a map where the key uniquely identifies the product+category combo.
	// The value can just be the category ID needed for joining.
	type expenseKey struct { // Define a local struct for the composite key
		ProductName string
		CategoryId  string
	}
	uniqueExpenseProducts := make(map[expenseKey]struct{}) // Value is empty struct{} - memory efficient presence check

	for _, exp := range expenses {
		// Basic validation/cleaning (optional, but good practice)
		trimmedName := strings.TrimSpace(exp.Name)
		trimmedCatID := strings.TrimSpace(exp.CategoryId)

		if trimmedName != "" && trimmedCatID != "" {
			key := expenseKey{ProductName: trimmedName, CategoryId: trimmedCatID}
			uniqueExpenseProducts[key] = struct{}{} // Add/overwrite entry
		}
	}
	log.Printf("Found %d unique product name + category ID combinations.", len(uniqueExpenseProducts))

	catalogResult := make(ProductCatalog)

	for expKey := range uniqueExpenseProducts {
		catInfo, found := categoryMap[expKey.CategoryId]
		if !found {
			log.Printf("Warning: Category ID '%s' (for product '%s') not found in category map. Skipping.", expKey.CategoryId, expKey.ProductName)
			continue
		}

		productInfo := ProductClassification{
			ProductName: expKey.ProductName,
			// Create the path - ensure order matches LLM expectation
			CategoryID:   expKey.CategoryId,
			CostGroup:    catInfo.CostGroup,
			MainCategory: catInfo.MainCategory,
			SubCategory:  catInfo.SubCategory,
		}

		catalogResult[productInfo.CostGroup] = append(catalogResult[productInfo.CostGroup], productInfo)

	}

	return catalogResult, nil
}
