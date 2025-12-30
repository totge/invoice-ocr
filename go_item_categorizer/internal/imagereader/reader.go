package imagereader

import (
	"context"
	"io"
	"log"
	"os"

	"github.com/totge/invoice-oc/go_item_categorizer/internal/app"
	"github.com/totge/invoice-oc/go_item_categorizer/internal/domain"
)

type Reader struct {
	filePath   string
	fileFormat string
}

var _ app.ReceiptImageReader = (*Reader)(nil)

func (r *Reader) ReadReceiptImage(ctx context.Context) (*domain.ImageSource, error) {
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

	return &domain.ImageSource{Data: data}, nil
}

func NewReader(filePath string) *Reader {
	return &Reader{filePath: filePath}
}
