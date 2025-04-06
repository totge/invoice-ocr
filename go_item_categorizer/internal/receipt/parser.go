package receipt

import "encoding/json"

func ParseReceipt(jsonData []byte) (*Receipt, error){
	var r Receipt
	err := json.Unmarshal(jsonData, &r)

	return &r, err
}
