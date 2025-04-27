package llm

import (
	"github.com/google/generative-ai-go/genai"
	"github.com/totge/invoice-oc/go_item_categorizer/internal/receipt"
)

type GeminiModel string

// TODO:DO I need this type??
type prompt struct {
	systemPrompt *genai.Content
	taskPrompt genai.Text
	inuptData genai.Text
	examples genai.Text
	outputFormat *genai.Schema
}


// stage1Data holds the dynamic data needed for the stage1_cost_group.tmpl template.
type stage1Input struct {
	CostGroups []string       // Slice of available cost group names
	Items      []receipt.Item // Slice of items from the current receipt
}

var stage1OutputFormat = &genai.Schema{
	Type: genai.TypeObject,
	Properties: map[string]*genai.Schema{
		"item_name": {
			Type:        genai.TypeString,
			Description: "original name of the item, exactly as it was provided in the input",
			Nullable:    false,
		},
		"cost_group": {
			Type:        genai.TypeString,
			Description: "original name of best corresponding cost group, exactly as it was provided in the input",
			Nullable:    false,
		},
	},
}

// // stage2Data holds the dynamic data needed for the stage2_detailed_match.tmpl template.
// type stage2Input struct {
// 	Item              receipt.Item                    // The specific receipt item being processed
// 	AssignedCostGroup string                          // The cost group assigned in stage 1
// 	Candidates        []catalog.ProductClassification // Filtered list of existing products in that cost group
// }

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
	Timestamp   string            `json:"datetime"`
	Items       []CategorizedItem `json:"items"`
	ParsedTotal int               `json:"parsed_total"`
}
