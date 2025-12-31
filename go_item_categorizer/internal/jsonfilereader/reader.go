package jsonfilereader

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/totge/invoice-oc/go_item_categorizer/internal/app"
	"github.com/totge/invoice-oc/go_item_categorizer/internal/domain"
)

type Reader struct {
	filePath string
}

var _ app.ReceiptReader = (*Reader)(nil)

func (r *Reader) ReadReceipt(ctx context.Context) (*domain.Receipt, error) {
	file, err := os.Open(r.filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open json file %q: %w", r.filePath, err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read json data: %w", err)
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
