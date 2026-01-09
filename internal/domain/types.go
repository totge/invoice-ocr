package domain

import "time"

type SourceInfo struct {
	Name      string // Display name (e.g. "receipt_001.jpg")
	Reference string // ID of file that could be passed to --input (e.g. "/tmp/receipts/receipt_001.jpg")
	Extension string
	Size      int64
	ModTime   time.Time
}

type ImageSource struct {
	Format string
	Data   []byte
}

type Item struct {
	ItemName string `json:"name"`
	Price    int    `json:"price"`
	Discount int    `json:"discount"`
}

type Receipt struct {
	Timestamp   string `json:"datetime"`
	Items       []Item `json:"items"`
	ParsedTotal int    `json:"parsed_total"`
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
