package categorizer

import (
	"github.com/totge/invoice-oc/go_item_categorizer/internal/llm"
	"github.com/totge/invoice-oc/go_item_categorizer/internal/receipt"
)

func CategorizeReceipt(rec receipt.Receipt, categoryMapping map[string]llm.CategoryHierarchy) CategorizedReceipt {
	categorizedItems := make([]CategorizedItem, 0, len(rec.Items))

	for _, item := range rec.Items {
		category := categoryMapping[item.Name]

		categorizedItems = append(categorizedItems, CategorizedItem{
			OriginalName: item.Name,
			Price:        item.Price,
			Discount:     item.Discount,
			CostGroup:    category.CostGroup,
			MainCategory: category.MainCategory,
			Subcategory:  category.Subcategory,
			ProductName:  category.ProductName,
		})
	}

	return CategorizedReceipt{
		Timestamp:   rec.Timestamp,
		Items:       categorizedItems,
		ParsedTotal: rec.ParsedTotal,
	}
}
