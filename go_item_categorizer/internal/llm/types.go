package llm

import (
	"github.com/totge/invoice-oc/go_item_categorizer/internal/catalog"
	"github.com/totge/invoice-oc/go_item_categorizer/internal/receipt"
)

type GeminiModel string 
// stage1Data holds the dynamic data needed for the stage1_cost_group.tmpl template.
type stage1Data struct {
	CostGroups []string       // Slice of available cost group names
	Items      []receipt.Item // Slice of items from the current receipt
}

// stage2Data holds the dynamic data needed for the stage2_detailed_match.tmpl template.
type stage2Data struct {
	Item              receipt.Item                    // The specific receipt item being processed
	AssignedCostGroup string                          // The cost group assigned in stage 1
	Candidates        []catalog.ProductClassification // Filtered list of existing products in that cost group
}

type CategorizedItem struct {
	Name         string `json:"name"`
	Price        int    `json:"price"`
	Discount     int    `json:"discount"`
	CategoryId   string `json:"category_id"`
	MainCategory string `json:"main_category"`
	Subcategory  string `json:"subcategory"`
	ProductName  string `json:"product_name"`
}
type CategorizedReceipt struct {
	Timestamp   string `json:"datetime"`
	Items       []CategorizedItem `json:"items"`
	ParsedTotal int    `json:"parsed_total"`
}
