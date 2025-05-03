package categorizer

type CategorizedItem struct {
	OriginalName string `json:"name"`
	Price        int    `json:"price"`
	Discount     int    `json:"discount"`
	CostGroup    string `json:"category_id"`
	MainCategory string `json:"main_category"`
	Subcategory  string `json:"subcategory"`
	ProductName  string `json:"product_name"`
}

type CategorizedReceipt struct {
	Timestamp   string            `json:"datetime"`
	Items       []CategorizedItem `json:"items"`
	ParsedTotal int               `json:"parsed_total"`
}
