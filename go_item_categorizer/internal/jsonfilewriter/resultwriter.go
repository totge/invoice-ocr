package jsonfilewriter

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/totge/invoice-oc/go_item_categorizer/internal/domain"
)

type Writer struct {
	filePath string
}

// New validates the path and returns a writer ready to be used.
func New(filePath string) (*Writer, error) {
	if filePath == "" {
		return nil, fmt.Errorf("output file path cannot be empty")
	}

	// Get the directory containing the target file.
	dir := filepath.Dir(filePath)

	// Check if the directory exists and we can write to it.
	// os.Stat gets file info. If it returns an error, the path might not exist.
	info, err := os.Stat(dir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("output directory %q does not exist", dir)
		}
		// Some other error occurred when checking the directory.
		return nil, fmt.Errorf("could not verify output directory %q: %w", dir, err)
	}

	if !info.IsDir() {
		return nil, fmt.Errorf("output path %q is not a directory", dir)
	}

	// A simple way to check for write permissions is to try creating and deleting a temporary file.
	// This is a very robust check.
	tempFile, err := os.CreateTemp(dir, "permission-check-*.tmp")
	if err != nil {
		return nil, fmt.Errorf("no write permissions for output directory %q: %w", dir, err)
	}
	tempFile.Close()
	os.Remove(tempFile.Name())

	// All checks passed!
	return &Writer{filePath: filePath}, nil
}

// WriteResult now uses the pre-validated path.
func (w *Writer) WriteResult(ctx context.Context, receipt *domain.CategorizedReceipt) error {
	// 1. Respect context cancellation.
	// Before performing any work, check if the context has been cancelled.
	select {
	case <-ctx.Done():
		return ctx.Err() // Return the context's error (e.g., context.Canceled)
	default:
		// Context is still valid, proceed.
	}
	
	// File operations are now done here. We are confident this will
	// likely succeed because of the checks in New().
	file, err := os.Create(w.filePath)
	if err != nil {
		return fmt.Errorf("failed to create file %q: %w", w.filePath, err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "    ")

	if err := encoder.Encode(receipt); err != nil {
		return fmt.Errorf("failed to encode and write JSON to file: %w", err)
	}

	return nil
}
