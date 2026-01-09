package localfilelister

import (
	"context"
	"fmt"
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
	return &Lister{directory: dir}
}

func (l *Lister) ListSources(ctx context.Context) ([]domain.SourceInfo, error) {
	entries, err := os.ReadDir(l.directory)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	var sources []domain.SourceInfo
	for _, e := range entries {
		if e.IsDir() {
			continue
		}

		// Simple filtering for known extensions could happen here

		sources = append(sources, domain.SourceInfo{
			Name:      e.Name(),
			Reference: filepath.Join(l.directory, e.Name()),
		})
	}
	return sources, nil
}
