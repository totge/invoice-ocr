package llm

import (
	"github.com/google/generative-ai-go/genai"
)

type GeminiModel string

type Prompt struct {
	SystemPrompt *genai.Content
	TaskPrompt   genai.Text
	InuptData    genai.Text
	Examples     genai.Text
	OutputFormat *genai.Schema
}

type ProductInfo interface {
	GetCostGroups() []string
	GetProductListJSON(string) ([]byte, error)
}

type ItemInfo interface {
	GetName() string
	GetPrice() int
	GetDiscount() int
}

// stage1Data holds the dynamic data needed for the stage1_cost_group.tmpl template.
type stage1Input struct {
	CostGroups []string   // Slice of available cost group names
	Items      []ItemInfo // Slice of items from the current receipt
}

var stage1OutputFormat = &genai.Schema{
	Type: genai.TypeArray,
	Items: &genai.Schema{
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
	},
}

type stage1Output struct {
	ItemName  string `json:"item_name"`
	CostGroup string `json:"cost_group"`
}

type stage2Input struct {
	CostGroup string
	Items     []ItemInfo
	Products  string
}

var stage2OutputFormat = &genai.Schema{
	Type: genai.TypeArray,
	Items: &genai.Schema{
		Type: genai.TypeObject,
		Properties: map[string]*genai.Schema{
			"item_name": {
				Type:        genai.TypeString,
				Description: "original name of the item, exactly as it was provided in the input",
				Nullable:    false,
			},
			"category_id": {
				Type:        genai.TypeString,
				Description: "the category_id of the best matching row, exactly as it was provided in the input",
				Nullable:    false,
			},
			"cost_group": {
				Type:        genai.TypeString,
				Description: "name of the cost group, exactly as it was provided in the input",
				Nullable:    false,
			},
			"main_category": {
				Type:        genai.TypeString,
				Description: "main category of the best fitting product from the product list",
				Nullable:    false,
			},
			"subcategory": {
				Type:        genai.TypeString,
				Description: "subcategory of the best fitting product from the product list",
				Nullable:    false,
			},
			"product_name": {
				Type:        genai.TypeString,
				Description: "best fitting general product name, selected from the provided product list",
				Nullable:    false,
			},
		},
	},
}

type stage2Output struct {
	CategoryId   string `json:"category_id"`
	ItemName     string `json:"item_name"`
	CostGroup    string `json:"cost_group"`
	MainCategory string `json:"main_category"`
	Subcategory  string `json:"subcategory"`
	ProductName  string `json:"product_name"`
}

type CategoryHierarchy struct {
	CostGroup    string
	MainCategory string
	Subcategory  string
	ProductName  string
}
