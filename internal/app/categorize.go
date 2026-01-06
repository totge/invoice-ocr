package app

import (
	"context"
	"fmt"
	"log/slog"
)

func Categorize(ctx context.Context, inputReader ReceiptReader, productCatalog ProductLister, catAssigner Categorizer, writer ResultWriter) error {

	slog.Info("Starting receipt categorization pipeline")

	slog.Info("Reading structured receipt input")
	receipt, err := inputReader.ReadReceipt(ctx)
	if err != nil {
		return fmt.Errorf("failed to read receipt input: %w", err)
	}
	slog.Info("Input loaded", "items_to_process", len(receipt.Items))

	slog.Info("Fetching product catalog")
	products, err := productCatalog.ListProducts(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch product catalog: %w", err)
	}
	slog.Debug("Catalog loaded", "catalog_size", len(products))

	slog.Info("Categorizing items with LLM...")
	enrichedReceipt, err := catAssigner.Categorize(ctx, receipt, products)
	if err != nil {
		return fmt.Errorf("failed to categorize receipt items: %w", err)
	}
	slog.Info("Categorization complete")

	slog.Info("Writing results")
	err = writer.WriteResult(ctx, enrichedReceipt)
	if err != nil {
		return fmt.Errorf("failed to write final results: %w", err)
	}

	slog.Info("Pipeline finished successfully")
	return nil
}
