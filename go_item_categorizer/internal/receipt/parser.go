package receipt

import (
	"encoding/json"
	"io"
)

func ParseReceipt(jsonData io.Reader) (*Receipt, error){
	var r Receipt

	data, err := io.ReadAll(jsonData)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(data, &r)
	if err != nil {
		return nil, err
	}

	return &r, err
}
