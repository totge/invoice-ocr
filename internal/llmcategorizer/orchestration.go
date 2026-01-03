package llmcategorizer

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/totge/invoice-oc/go_item_categorizer/internal/domain"
)

// executeStage1 handles the entire process for the first stage.
// It returns a map correlating each item name to its assigned cost group.
func (c *Categorizer) executeStage1(ctx context.Context, items []domain.Item, allCostGroups []string) (map[string]string, error) {

	slog.Info("Starting Stage 1: Cost Group Assignment", "item_count", len(items))

	// 1. Build the prompt using promptBuilder.
	prompt, err := builder.buildStage1Prompt(items, allCostGroups)
	if err != nil {
		return nil, fmt.Errorf("failed to build stage 1 prompt: %w", err)
	}

	// 2. Call the LLM.
	slog.Debug("Sending Stage 1 prompt to LLM", "model", c.modelName)

	jsonResponse, err := c.client.GenerateJSON(ctx, c.modelName, prompt)
	if err != nil {
		return nil, fmt.Errorf("llm generation for stage 1 failed: %w", err)
	}

	// 3. Unmarshal the response.
	var stage1Results []stage1ResponseItem
	if err := json.Unmarshal([]byte(jsonResponse), &stage1Results); err != nil {
		return nil, fmt.Errorf("failed to decode stage 1 llm response json: %w", err)
	}

	slog.Debug("Stage 1 response received", "categorized_items", len(stage1Results))

	// 4. Correlate the results into a simple map for the next stage.
	itemToCostGroup := make(map[string]string, len(stage1Results))
	for _, result := range stage1Results {
		itemToCostGroup[result.ItemName] = result.CostGroup
	}

	return itemToCostGroup, nil
}

// executeStage2 handles the batching and execution of the second stage.
// It returns the final map correlating item names to their detailed categorization.
func (c *Categorizer) executeStage2(ctx context.Context, items []domain.Item, stage1ItemMapping map[string]string, productsByCostGroup map[string][]domain.ProductClassification) (map[string]stage2ResponseItem, error) {

	slog.Info("Starting Stage 2: Detailed Product Matching")

	// 1. Group items for batching by the cost group assigned in stage 1.
	itemsByCostGroup := make(map[string][]domain.Item)
	// This requires mapping the names back to original items. We can improve this if needed.

	for _, item := range items {
		costGroup, ok := stage1ItemMapping[item.ItemName]
		if !ok {

			slog.Warn("Item missed in Stage 1, assigning to 'Unassigned'", "item_name", item.ItemName)
			// collecting items not found in stage1 response to a separate costgroup
			costGroup = "Unassigned"
		}

		if itemList, ok := itemsByCostGroup[costGroup]; ok {
			itemsByCostGroup[costGroup] = append(itemList, item)
		} else {
			itemsByCostGroup[costGroup] = []domain.Item{item}
		}
	}

	finalMappings := make(map[string]stage2ResponseItem)

	// 2. Loop through each batch and call the LLM.
	for costGroup, itemsInGroup := range itemsByCostGroup {
		productCandidates := productsByCostGroup[costGroup]

		// handling case of no candidates (eg. 'Unassigned' costgroup)
		if len(productCandidates) == 0 {
			slog.Warn("No product candidates found for cost group, skipping Stage 2 for batch", "cost_group", costGroup, "item_count", len(itemsInGroup))
			// leave these items out, they will be handled as 'Uncategorized' later
			continue
		}

		slog.Debug("Processing Stage 2 Batch", "cost_group", costGroup, "item_count", len(itemsInGroup))
		
		// 2a. Build the prompt for this batch.
		prompt, err := builder.buildStage2Prompt(costGroup, itemsInGroup, productCandidates)
		if err != nil {
			return nil, fmt.Errorf("failed to build stage 2 prompt for cost group %s: %w", costGroup, err)
		}

		// 2b. Call the LLM.
		jsonResponse, err := c.client.GenerateJSON(ctx, c.modelName, prompt)
		if err != nil {
			return nil, fmt.Errorf("llm generation for stage 2 (cost group %s) failed: %w", costGroup, err)
		}

		// 2c. Unmarshal and correlate the results for this batch.
		var stage2Results []stage2ResponseItem
		if err := json.Unmarshal([]byte(jsonResponse), &stage2Results); err != nil {
			return nil, fmt.Errorf("failed to decode stage 2 llm response for cost group %s: %w", costGroup, err)
		}
		for _, result := range stage2Results {
			finalMappings[result.ItemName] = result
		}
	}

	return finalMappings, nil
}
