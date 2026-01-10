package localfilelister

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/totge/invoice-oc/go_item_categorizer/internal/app"
	"github.com/totge/invoice-oc/go_item_categorizer/internal/domain"
)

type Lister struct {
	directory string
}

var _ app.SourceLister = (*Lister)(nil)

func New(dir string) *Lister {
	slog.Debug("Initializing local file lister", "directory", dir)
	return &Lister{directory: dir}
}

func (l *Lister) ListSources(ctx context.Context) ([]domain.SourceInfo, error) {

	slog.Debug("Listing files in directory", "path", l.directory)

	entries, err := os.ReadDir(l.directory)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	var sources []domain.SourceInfo
	for _, e := range entries {
		if e.IsDir() {
			continue
		}

		info, err := e.Info()
		if err != nil {
			// Log failure for the file and continue.
			slog.Warn("Failed to get file info, skipping", "file", e.Name(), "error", err)
			continue
		}

		sources = append(sources, domain.SourceInfo{
			Name:      e.Name(),
			Reference: filepath.Join(l.directory, e.Name()),
			Extension: filepath.Ext(e.Name()),
			Size:      info.Size(),
			ModTime:   info.ModTime(),
		})
	}

	slog.Debug("Directory listing complete", "files_found", len(sources))
	return sources, nil
}
