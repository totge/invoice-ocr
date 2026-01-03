package ocrextractor

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/totge/invoice-oc/go_item_categorizer/internal/app"
	"github.com/totge/invoice-oc/go_item_categorizer/internal/domain"
	"github.com/totge/invoice-oc/go_item_categorizer/internal/llm"
)

type OCRPipeline struct {
	client    llm.Client
	modelName string
}

var _ app.Extractor = (*OCRPipeline)(nil)

func New(client llm.Client, modelName string) *OCRPipeline {
	return &OCRPipeline{
		client:    client,
		modelName: modelName,
	}
}

func (op *OCRPipeline) ExtractReceipt(ctx context.Context, raw *domain.ImageSource) (*domain.Receipt, error) {

	slog.Debug("Starting OCR extraction pipeline", "model", op.modelName)

	// 1. Build the prompt, using the global builder instance
	prompt, err := builder.buildOCRPrompt(raw)
	if err != nil {
		return nil, fmt.Errorf("failed to build OCR prompt: %w", err)
	}

	// 2. Call the LLM
	slog.Debug("Sending OCR request to LLM")

	jsonResponse, err := op.client.GenerateJSON(ctx, op.modelName, prompt)
	if err != nil {
		return nil, fmt.Errorf("OCR LLM generation failed: %w", err)
	}

	// 3. Unmarshal into our INTERNAL raw format first
	var rawResult rawReceipt
	if err := json.Unmarshal([]byte(jsonResponse), &rawResult); err != nil {
		return nil, fmt.Errorf("failed to decode OCR response JSON: %w", err)
	}

	slog.Debug("OCR raw response received",
		"items_found", len(rawResult.Items),
		"parsed_total", rawResult.ParsedTotal,
	)

	// 4. Map the raw result to the Domain format (performing calculations)
	return op.mapToDomain(&rawResult), nil
}

// mapToDomain handles the conversion from the LLM's simple as-is recipt data
// to the application's domain data definition. This mianly involves calculating the price paid for products.
func (op *OCRPipeline) mapToDomain(raw *rawReceipt) *domain.Receipt {
	domainItems := make([]domain.Item, len(raw.Items))

	// TODO: Add a warning if the discount is not a negative number
	for i, rItem := range raw.Items {
		// Calculate the actual paid price.
		// Assumption: Discount is a negative number (e.g. -100).
		if rItem.Discount > 0 {
			slog.Warn("Detected positive discount value from OCR",
				"item_name", rItem.Name,
				"discount_value", rItem.Discount,
				"recommendation", "Check receipt to verify if this is a surcharge or LLM error",
			)
		}
		// Paid Price = Full Price + Discount (e.g. 1000 + (-100) = 900).
		paidPrice := rItem.FullPrice + rItem.Discount

		domainItems[i] = domain.Item{
			ItemName: rItem.Name,
			Price:    paidPrice,           // The calculated final price
			Discount: -1 * rItem.Discount, // Store discunt amount as positive integer
		}
	}

	return &domain.Receipt{
		Timestamp:   raw.Timestamp,
		Items:       domainItems,
		ParsedTotal: raw.ParsedTotal,
	}
}
