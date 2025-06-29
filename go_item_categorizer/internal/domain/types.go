package domain

type Item struct {
	ItemName string `json:"name"`
	Price int `json:"price"`
	Discount int `json:"discount"`
}

type Receipt struct {
	Timestamp string `json:"datetime"`
	Items []Item `json:"items"`
	ParsedTotal int `json:"parsed_total"`
}

type ProductClassification struct {
	ProductName  string `json:"product_name"`
	CostGroup    string `json:"cost_group"`
	MainCategory string `json:"main_category"`
	SubCategory  string `json:"subcategory"`
	CategoryID   string `json:"category_id"`
}

// TODO: it is probably rendundant and ProductClassification should be used instead
type CategoryHierarchy struct {
	CostGroup    string
	MainCategory string
	Subcategory  string
	ProductName  string
}

// TODO: it might be better to just embed item and product classification
// defenetly a better option here, and the actual target writer can transform it to local types if needed for any reason
type CategorizedItem struct {
	Item
	ProductClassification
}

type CategorizedReceipt struct {
	Timestamp   string            `json:"datetime"`
	Items       []CategorizedItem `json:"items"`
	ParsedTotal int               `json:"parsed_total"`
}