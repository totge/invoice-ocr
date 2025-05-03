package catalog

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/totge/invoice-oc/go_item_categorizer/internal/appsheet"
)

type ProductClassification struct {
	ProductName  string `json:"product_name"`
	CostGroup    string `json:"cost_group"`
	MainCategory string `json:"main_category"`
	SubCategory  string `json:"subcategory"`
	CategoryID   string `json:"category_id"`
}

type ProductCatalog map[string][]ProductClassification

func (p *ProductCatalog) GetCostGroups() []string {
	costGroups := make([]string, 0, len(*p))

	for k := range *p {
		costGroups = append(costGroups, k)
	}

	return costGroups
}

func (p *ProductCatalog) GetProductListJSON(costGroup string) ([]byte, error) {

	products, ok := (*p)[costGroup]
	if !ok {
		log.Printf("cost group '%s' not found in catalog\n", costGroup)
		return nil, fmt.Errorf("cost group '%s' not found", costGroup)
	}

	serializedData, err := json.Marshal(products)
	if err != nil {
		// Handle potential errors during JSON marshalling
		log.Printf("error marshalling products for cost group '%s': %v\n", costGroup, err)
		return nil, fmt.Errorf("failed to marshal products for cost group '%s': %w", costGroup, err)
	}

	return serializedData, nil
}

func BuildProductCatalog(ctx context.Context, client *appsheet.Client) (ProductCatalog, error) {

	// --- 2. Fetch Categories from AppSheet ---
	log.Printf("Fetching categories from table '%s'...", string(appsheet.TableCategories))
	// Prepare the data structure to hold the results

	// Call the client method, passing the specific table name and the data holder
	categories, err := appsheet.ReadRecords[appsheet.Category](client, ctx, appsheet.TableCategories)
	if err != nil {
		// This error includes request errors OR the decoding error IF you fixed ReadRecords
		log.Fatalf("FATAL: Failed to fetch categories: %v", err)
	}

	categoryMap := make(map[string]appsheet.Category)
	for _, cat := range categories {
		categoryMap[cat.CategoryId] = cat
	}

	log.Printf("Fetching expenses from table '%s'...", string(appsheet.TableExpenses))
	// Prepare the data structure to hold the results

	// Call the client method, passing the specific table name and the data holder
	expenses, err := appsheet.ReadRecords[appsheet.Expense](client, ctx, appsheet.TableExpenses)
	if err != nil {
		// This error includes request errors OR the decoding error IF you fixed ReadRecords
		log.Fatalf("FATAL: Failed to fetch expenses: %v", err)
	}

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
