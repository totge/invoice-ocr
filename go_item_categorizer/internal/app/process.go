package app

import (
	"context"
	"fmt"
)

func Process(ctx context.Context, inputReader ReceiptReader, productCatalog ProductLister, catAssigner Categorizer, writer ResultWriter) error {

	receipt, err := inputReader.ReadReceipt(ctx)
	if err != nil {
		return fmt.Errorf("failed to read receipt input: %w", err)
	}
	products, err := productCatalog.ListProducts(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch product catalog: %w", err)
	}

	enrichedReceipt, err := catAssigner.Categorize(ctx, receipt, products)
	if err != nil {
		return fmt.Errorf("failed to categorize receipt items: %w", err)
	}

	err = writer.WriteResult(ctx, enrichedReceipt)
	if err != nil {
		return fmt.Errorf("failed to write final results: %w", err)
	}

	return nil
}
