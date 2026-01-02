package imagereader

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"

	"github.com/totge/invoice-oc/go_item_categorizer/internal/app"
	"github.com/totge/invoice-oc/go_item_categorizer/internal/domain"
)

type Reader struct {
	filePath string
}

var _ app.ReceiptImageReader = (*Reader)(nil)

func (r *Reader) ReadReceiptImage(ctx context.Context) (*domain.ImageSource, error) {

	slog.Debug("Reading receipt image from file", "path", r.filePath)

	file, err := os.Open(r.filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open image file %q: %w", r.filePath, err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read image data: %w", err)
	}
	slog.Debug("Image file read successfully", "size_bytes", len(data))

	// detecting MIME type
	mimeType := http.DetectContentType(data)
	slog.Debug("Image format detected", "mime_type", mimeType)
	
	return &domain.ImageSource{
		Data:   data,
		Format: mimeType,
	}, nil
}

func NewReader(filePath string) *Reader {

	return &Reader{filePath: filePath}
}
