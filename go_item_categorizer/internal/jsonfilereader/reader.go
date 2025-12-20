package jsonfilereader

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"os"

	"github.com/totge/invoice-oc/go_item_categorizer/internal/app"
	"github.com/totge/invoice-oc/go_item_categorizer/internal/domain"
)

type Reader struct {
	filePath string
}

var _ app.ReceiptReader = (*Reader)(nil)

func (r *Reader) ReadReceipt(ctx context.Context) (*domain.Receipt, error) {
	jsonData, err := os.Open(r.filePath)
	// TODO: add context to the error
	if err != nil {
		log.Fatalf("FATAL: Failed to open file %s: %v", r.filePath, err)
		return nil, err
	}

	data, err := io.ReadAll(jsonData)
	if err != nil {
		return nil, err
	}

	var receipt domain.Receipt

	err = json.Unmarshal(data, &receipt)
	if err != nil {
		return nil, err
	}

	return &receipt, nil
}

func NewReader(filePath string) *Reader {
	return &Reader{filePath: filePath}
}
