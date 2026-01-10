package sourcefilter

import (
	"log/slog"
	"strings"

	"github.com/totge/invoice-oc/go_item_categorizer/internal/app"
	"github.com/totge/invoice-oc/go_item_categorizer/internal/domain"
)

var _ app.SourceFilter = (*Filter)(nil)

// Supported extensions could be defined here or injected
var validExtensions = map[string]bool{
	".json": true,
	".jpg":  true,
	".jpeg": true,
	".png":  true,
}

type Filter struct {
	LimitNumber      int
	SupportedFormats map[string]bool
}

func New(limit int) *Filter {
	sanitizedLimit := limit
	if limit < 0 {
		slog.Warn("Received negative limit, treating as unlimited", "limit", limit)
		sanitizedLimit = 0
	}

	slog.Debug("Initializing source filter", "limit", sanitizedLimit)

	return &Filter{
		LimitNumber:      sanitizedLimit,
		SupportedFormats: validExtensions,
	}
}

func (f *Filter) Filter(rawList []domain.SourceInfo) []domain.SourceInfo {
	slog.Debug("Filtering source list", "input_count", len(rawList))

	var filtered []domain.SourceInfo

	var keys []string
	for k := range f.SupportedFormats {
		keys = append(keys, k)
	}
	slog.Debug("supported formats", "formats", keys)

	// filter for supported file formats
	for _, item := range rawList {
		ext := strings.ToLower(item.Extension)
		if f.SupportedFormats[ext] {
			filtered = append(filtered, item)
		}
	}

	droppedCount := len(rawList) - len(filtered)
	slog.Debug("Extension filtering complete",
		"kept_count", len(filtered),
		"dropped_count", droppedCount,
	)

	// limiting output
	if f.LimitNumber > 0 && len(filtered) > f.LimitNumber {
		originalCount := len(filtered)
		filtered = filtered[:f.LimitNumber]
		slog.Debug("List truncated by limit",
			"limit", f.LimitNumber,
			"truncated_items", originalCount-len(filtered),
		)
	}

	return filtered
}
