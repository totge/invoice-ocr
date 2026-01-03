package jsonfilereader

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/totge/invoice-oc/go_item_categorizer/internal/app"
	"github.com/totge/invoice-oc/go_item_categorizer/internal/domain"
)

type Reader struct {
	filePath string
}

var _ app.ReceiptReader = (*Reader)(nil)

func (r *Reader) ReadReceipt(ctx context.Context) (*domain.Receipt, error) {
	slog.Debug("Reading receipt from JSON file", "path", r.filePath)

	file, err := os.Open(r.filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open json file %q: %w", r.filePath, err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read json data: %w", err)
	}
	slog.Debug("File read successfully", "size_bytes", len(data))

	var receipt domain.Receipt

	err = json.Unmarshal(data, &receipt)
	if err != nil {
		return nil, err
	}
	slog.Debug("Receipt unmarshalled successfully",
		"timestamp", receipt.Timestamp,
		"item_count", len(receipt.Items),
	)

	return &receipt, nil
}

func NewReader(filePath string) *Reader {
	return &Reader{filePath: filePath}
}
