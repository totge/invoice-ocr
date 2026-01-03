package app

import (
	"context"
	"fmt"
	"log/slog"
)

func Extract(ctx context.Context, reader ReceiptImageReader, extractor Extractor, writer ReceiptWriter) error {

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

	slog.Info("Writing extracted data to output")
	err = writer.WriteReceipt(ctx, receipt)
	if err != nil {
		return fmt.Errorf("failed to write data from receipt: %w", err)
	}

	slog.Info("Pipeline finished successfully")
	return nil
}
