package sourcefilter

import (
	"strings"

	"github.com/totge/invoice-oc/go_item_categorizer/internal/app"
	"github.com/totge/invoice-oc/go_item_categorizer/internal/domain"
)

var _ app.SourceFilter = (*Filter)(nil)

type Filter struct {
	LimitNumber      int
	SupportedFormats map[string]bool
}

func New(limit int, supportedExtensions ...string) *Filter {
	supportedFormats := make(map[string]bool, len(supportedExtensions))
	for _, format := range supportedExtensions {
		supportedFormats[format] = true
	}

	return &Filter{LimitNumber: limit, SupportedFormats: supportedFormats}
}

func (f *Filter) Filter(rawList []domain.SourceInfo) []domain.SourceInfo {
	var filtered []domain.SourceInfo

	// filter for supported file formats
	for _, item := range rawList {
		ext := strings.ToLower(item.Extension)
		if f.SupportedFormats[ext] {
			filtered = append(filtered, item)
		}
	}

	// limiting output
	if f.LimitNumber > 0 && len(filtered) > f.LimitNumber {
		filtered = filtered[:f.LimitNumber]
	}

	return filtered
}
