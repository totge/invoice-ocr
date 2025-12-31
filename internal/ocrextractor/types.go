package ocrextractor

import "encoding/json"

type rawReceipt struct {
	Timestamp   string    `json:"timestamp"`
	Items       []rawItem `json:"items"`
	ParsedTotal int       `json:"parsed_total"`
}

type rawItem struct {
	Name      string `json:"name"`
	FullPrice int    `json:"full_price"`
	Discount  int    `json:"discount"`
}

var ocrSchema = json.RawMessage(`{
    "type": "OBJECT",
    "properties": {
        "timestamp": { 
            "type": "STRING", 
            "description": "ISO 8601 timestamp of purchase" 
        },
        "parsed_total": { 
            "type": "INTEGER", 
            "description": "Total amount paid in HUF" 
        },
        "items": {
            "type": "ARRAY",
            "items": {
                "type": "OBJECT",
                "properties": {
                    "name": { 
                        "type": "STRING",
                        "description": "Name of the product"
                    },
                    "full_price": { 
                        "type": "INTEGER",
                        "description": "The full price of the item before discount"
                    },
                    "discount": { 
                        "type": "INTEGER",
                        "description": "Discount amount (negative) or 0"
                    }
                },
                "required": ["name", "full_price", "discount"]
            }
        }
    },
    "required": ["timestamp", "parsed_total", "items"]
}`)
