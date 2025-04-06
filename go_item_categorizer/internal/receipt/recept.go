package receipt


type Item struct {
	Name string `json:"name"`
	Price int `json:"price"`
	Discount int `json:"discount"`
}

type Receipt struct {
	Timestamp string `json:"datetime"`
	Items []Item `json:"items"`
	ParsedTotal int `json:"parsed_total"`
}

