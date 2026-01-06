package app

import (
	"context"
	"fmt"
	"log/slog"
)

func Process(ctx context.Context, reader ReceiptImageReader, extractor Extractor, productCatalog ProductLister, categorizer Categorizer, writer ResultWriter) error {
	slog.Info("Starting receipt extraction pipeline")

	slog.Info("Reading input image")
	image, err := reader.ReadReceiptImage(ctx)
	if err != nil {
		return fmt.Errorf("failed to read receipt input: %w", err)
	}

	slog.Info("Sending image to LLM for data extraction...")
	receipt, err := extractor.ExtractReceipt(ctx, image)
	if err != nil {
		return fmt.Errorf("failed to extarct receipt data from image: %w", err)
	}
	slog.Info("Extraction successful",
		"timestamp", receipt.Timestamp,
		"parsed_total", receipt.ParsedTotal,
		"items_found", len(receipt.Items),
	)

	slog.Info("Extraction complete")

	slog.Info("Starting receipt categorization")

	slog.Info("Fetching product catalog")
	products, err := productCatalog.ListProducts(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch product catalog: %w", err)
	}
	slog.Debug("Catalog loaded", "catalog_size", len(products))

	slog.Info("Categorizing items...")
	enrichedReceipt, err := categorizer.Categorize(ctx, receipt, products)
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
