package llmcategorizer

import "encoding/json"

var stage1Schema = json.RawMessage(`{
    "type": "array",
    "items": {
        "type": "object",
        "properties": {
            "item_name": { 
				"type": "string",
				"description": "original name of the item, exactly as it was provided in the input",
				"nullable": false
			},
            "cost_group": { 
				"type": "string",
				"description": "original name of best corresponding cost group, exactly as it was provided in the input",
				"nullable": false
			}
        },
        "required": ["item_name", "cost_group"]
    }
}`)

type stage1ResponseItem struct {
	ItemName  string `json:"item_name"`
	CostGroup string `json:"cost_group"`
}

var stage2Schema = json.RawMessage(`{
    "type": "array",
    "items": {
        "type": "object",
        "properties": {
            "item_name": { 
				"type": "string",
				"description": "original name of the item, exactly as it was provided in the input",
				"nullable": false
			},
            "category_id": { 
				"type": "string",
				"description": "the category_id of the best matching row, exactly as it was provided in the input",
				"nullable": false
			},
            "cost_group": { 
				"type": "string",
				"description": "name of the cost group, exactly as it was provided in the input",
				"nullable": false
			},
            "main_category": { 
				"type": "string",
				"description": "main category of the best fitting product from the product list",
				"nullable": false
			},
            "subcategory": { 
				"type": "string",
				"description": "subcategory of the best fitting product from the product list",
				"nullable": false
			},
            "product_name": { 
				"type": "string",
				"description": "best fitting general product name, selected from the provided product list",
				"nullable": false
			}
        },
        "required": ["item_name", "category_id", "cost_group", "main_category", "subcategory", "product_name"]
    }
}`)

type stage2ResponseItem struct {
	CategoryID   string `json:"category_id"`
	ItemName     string `json:"item_name"`
	CostGroup    string `json:"cost_group"`
	MainCategory string `json:"main_category"`
	Subcategory  string `json:"subcategory"`
	ProductName  string `json:"product_name"`
}
