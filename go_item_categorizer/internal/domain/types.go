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

type CategorizedItem struct {
	Item
	ProductClassification
}

type CategorizedReceipt struct {
	Timestamp   string            `json:"datetime"`
	Items       []CategorizedItem `json:"items"`
	ParsedTotal int               `json:"parsed_total"`
}