package llmcategorizer

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/totge/invoice-oc/go_item_categorizer/internal/app"
	"github.com/totge/invoice-oc/go_item_categorizer/internal/domain"
	"github.com/totge/invoice-oc/go_item_categorizer/internal/llm"
)

// Categorizer is the main struct that implements the app.Categorizer interface.
type Categorizer struct {
	client    llm.Client // The agnostic LLM client interface from types.go
	modelName string
}

var _ app.Categorizer = (*Categorizer)(nil)

// New is the constructor for our categorizer.
func New(client llm.Client, modelName string) *Categorizer {
	return &Categorizer{
		client:    client,
		modelName: modelName,
	}
}

// Categorize is the high-level public method. It orchestrates the multi-stage
// process by delegating to private methods and then assembles the final result.
func (c *Categorizer) Categorize(ctx context.Context, receipt *domain.Receipt, products []domain.ProductClassification) (*domain.CategorizedReceipt, error) {

	slog.Debug("Categorizer started", "receipt_items", len(receipt.Items), "catalog_size", len(products))

	// 1. Data Preparation: A quick pre-processing step to make lookups easier.
	productsByCostGroup, allCostGroups := c.prepareProductData(products)

	// 2. Delegate to the Stage 1 orchestrator.
	// The result is a map correlating each original item to its assigned cost group.
	itemToCostGroup, err := c.executeStage1(ctx, receipt.Items, allCostGroups)
	if err != nil {
		return nil, fmt.Errorf("categorization failed during stage 1: %w", err)
	}

	// 3. Delegate to the Stage 2 orchestrator.
	// The result is a map correlating each original item to its final, detailed categorization.
	finalMappings, err := c.executeStage2(ctx, receipt.Items, itemToCostGroup, productsByCostGroup)
	if err != nil {
		return nil, fmt.Errorf("categorization failed during stage 2: %w", err)
	}

	// 4. Perform the final, simple assembly.
	result := c.buildFinalReceipt(receipt, finalMappings)

	slog.Info("Categorization complete", "successfully_mapped", len(finalMappings), "total_items", len(receipt.Items))

	return result, nil
}

// --- Private Helper Methods for Final Assembly ---

// prepareProductData is a simple helper to pre-process the product list for efficient use.
// It prepares a map of products by costgroup and a list of unique cost groups
func (c *Categorizer) prepareProductData(products []domain.ProductClassification) (map[string][]domain.ProductClassification, []string) {
	productsByGroup := make(map[string][]domain.ProductClassification)
	costGroupSet := make(map[string]struct{})

	for _, p := range products {
		productsByGroup[p.CostGroup] = append(productsByGroup[p.CostGroup], p)
		costGroupSet[p.CostGroup] = struct{}{}
	}

	allCostGroups := make([]string, 0, len(costGroupSet))
	for cg := range costGroupSet {
		allCostGroups = append(allCostGroups, cg)
	}
	return productsByGroup, allCostGroups
}

// buildFinalReceipt takes the original receipt and the final mappings and constructs the output.
func (c *Categorizer) buildFinalReceipt(originalReceipt *domain.Receipt, stage2Mapping map[string]stage2ResponseItem) *domain.CategorizedReceipt {
	categorizedItems := make([]domain.CategorizedItem, len(originalReceipt.Items))

	for i, item := range originalReceipt.Items {
		// Look up the final, enriched data for this specific item.
		// We use the original item name as the key for correlation.
		result, ok := stage2Mapping[item.ItemName]
		if !ok {
			// Handle cases where the LLM might have failed to categorize an item.
			// We can create a default/uncategorized item.
			categorizedItems[i] = domain.CategorizedItem{
				Item: domain.Item{
					ItemName: item.ItemName,
					Price:    item.Price,
					Discount: item.Discount,
				},
				ProductClassification: domain.ProductClassification{
					ProductName:  "Uncategorized", // Default value
					CostGroup:    "Uncategorized",
					MainCategory: "Uncategorized",
					SubCategory:  "Uncategorized",
					CategoryID:   "Uncategorized",
				},
			}
			continue
		}

		categorizedItems[i] = domain.CategorizedItem{
			Item: domain.Item{
				ItemName: item.ItemName,
				Price:    item.Price,
				Discount: item.Discount,
			},
			ProductClassification: domain.ProductClassification{
				CategoryID:   result.CategoryID,
				ProductName:  result.ProductName,
				CostGroup:    result.CostGroup,
				MainCategory: result.MainCategory,
				SubCategory:  result.Subcategory,
			},
		}
	}

	return &domain.CategorizedReceipt{
		Timestamp:   originalReceipt.Timestamp,
		Items:       categorizedItems,
		ParsedTotal: originalReceipt.ParsedTotal,
	}
}
