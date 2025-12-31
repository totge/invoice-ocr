package app

import (
	"context"
	"fmt"
)

func Extract(ctx context.Context, reader ReceiptImageReader, extractor Extractor, writer ReceiptWriter) error {

	image, err := reader.ReadReceiptImage(ctx)
	if err != nil {
		return fmt.Errorf("failed to read receipt input: %w", err)
	}

	receipt, err := extractor.ExtractReceipt(ctx, image)
	if err != nil {
		return fmt.Errorf("failed to extarct receipt data from image: %w", err)
	}

	err = writer.WriteReceipt(ctx, receipt)
	if err != nil {
		return fmt.Errorf("failed to write data from receipt: %w", err)
	}

	return nil
}
