package receipt


type Item struct {
	Name string `json:"name"`
	Price int `json:"price"`
	Discount int `json:"discount"`
}

func (i *Item) GetName() string{
	return i.Name
}

func (i *Item) GetPrice() int{
	return i.Price
}

func (i *Item) GetDiscount() int{
	return i.Discount
}


type Receipt struct {
	Timestamp string `json:"datetime"`
	Items []Item `json:"items"`
	ParsedTotal int `json:"parsed_total"`
}

