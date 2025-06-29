package app

import "context"

func Process(ctx context.Context, inputReader ReceiptReader, productCatalog ProductLister, catAssigner Categorizer, writer ResultWriter) error {

	receipt, err := inputReader.ReadReceipt(ctx)
	if err != nil {
		// TODO: logging or error formatting needed
		return err
	}
	products, err := productCatalog.ListProducts(ctx)
	if err != nil {
		// TODO: logging or error formatting needed
		return err
	}

	enrichedReceipt, err := catAssigner.Categorize(ctx, receipt, products)
	if err != nil {
		// TODO: logging or error formatting needed
		return err
	}

	err = writer.WriteResult(ctx, enrichedReceipt)
	if err != nil {
		// TODO: logging or error formatting needed
		return err
	}

	return nil
}