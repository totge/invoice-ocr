package app

import (
	"context"

	"github.com/totge/invoice-oc/go_item_categorizer/internal/domain"
)

// SourceLister defines the capability to list available input files from a source system
type SourceLister interface {
	ListSources(ctx context.Context) ([]domain.SourceInfo, error)
}

// SourceFilter defines the capability to filter a list of input files
type SourceFilter interface {
	Filter([]domain.SourceInfo) []domain.SourceInfo
}

// SourceListWriter defines the capability to write a list of sources to a target
type SourceListWriter interface {
	WriteSourceList(context.Context, []domain.SourceInfo) error
}

type ReceiptImageReader interface {
	ReadReceiptImage(context.Context) (*domain.ImageSource, error)
}

type Extractor interface {
	ExtractReceipt(context.Context, *domain.ImageSource) (*domain.Receipt, error)
}

// ReceiptWriter defines the capability to write a raw/extracted receipt.
type ReceiptWriter interface {
	WriteReceipt(context.Context, *domain.Receipt) error
}

type ReceiptReader interface {
	ReadReceipt(context.Context) (*domain.Receipt, error)
}

type ProductLister interface {
	ListProducts(context.Context) ([]domain.ProductClassification, error)
}

type Categorizer interface {
	Categorize(context.Context, *domain.Receipt, []domain.ProductClassification) (*domain.CategorizedReceipt, error)
}

type ResultWriter interface {
	WriteResult(context.Context, *domain.CategorizedReceipt) error
}
