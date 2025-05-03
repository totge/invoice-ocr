package llm

import (
	"context"
	"encoding/json"
	"log"

	"github.com/google/generative-ai-go/genai"
	"github.com/totge/invoice-oc/go_item_categorizer/internal/catalog"
	"github.com/totge/invoice-oc/go_item_categorizer/internal/receipt"
)

func AssignCategoryData(client *genai.Client, ctx context.Context, items []receipt.Item, products catalog.ProductCatalog) (map[string]CategoryHierarchy, error) {

	costGroups := products.GetCostGroups()

	stg1Input := stage1Input{
		CostGroups: costGroups,
		Items:      items,
	}

	stg1Prompt, err := buildStage1Prompt(stg1Input)
	if err != nil {
		return nil, err
	}

	stg1Resp := generateContent(client, ctx, stg1Prompt)

	processedResp, err := processResponse[stage1Output](stg1Resp)
	if err != nil {
		return nil, err
	}

	stage2InputList, err := createStage2Input(items, products, processedResp)
	if err != nil {
		return nil, err
	}

	enrichedItems := make([]stage2Output, 0, len(items))
	// run stage two for each stage2 input
	for _, input := range stage2InputList {
		stage2prompt, err := buildStage2Prompt(input)
		if err != nil {
			return nil, err
		}

		stg2Resp := generateContent(client, ctx, stage2prompt)
		processedResp, err := processResponse[stage2Output](stg2Resp)
		if err != nil {
			return nil, err
		}
		enrichedItems = append(enrichedItems, processedResp...)
	}

	categoryMapping := buildCategoryMapping(enrichedItems)

	return categoryMapping, nil
}

func buildCategoryMapping(llmEnrichedData []stage2Output) map[string]CategoryHierarchy {
	categoryMapping := make(map[string]CategoryHierarchy, len(llmEnrichedData))

	for _, item := range llmEnrichedData {
		categoryMapping[item.ItemName] = CategoryHierarchy{
			CostGroup:    item.CostGroup,
			MainCategory: item.MainCategory,
			Subcategory:  item.Subcategory,
			ProductName:  item.ProductName,
		}
	}

	return categoryMapping
}

func createStage2Input(items []receipt.Item, products catalog.ProductCatalog, procesedStage1output []stage1Output) ([]stage2Input, error) {

	costGroupMapping := createItemCostGroupMapping(procesedStage1output)

	stage2GroupedItems := make(map[string][]receipt.Item)

	// grouping original items by cost group
	for _, item := range items {
		itemCostGroup := costGroupMapping[item.Name]

		// adding slice for new cost groups
		if itemList, ok := stage2GroupedItems[itemCostGroup]; ok {
			stage2GroupedItems[itemCostGroup] = append(itemList, item)
		} else {
			stage2GroupedItems[itemCostGroup] = []receipt.Item{item}
		}
	}

	stage2Inputs := make([]stage2Input, 0, len(stage2GroupedItems))
	// creating stage2 input structs
	for costGroup, itemList := range stage2GroupedItems {
		productList, err := products.GetProductListJSON(costGroup)
		if err != nil {
			return stage2Inputs, err
		}
		stage2Inputs = append(stage2Inputs, stage2Input{
			CostGroup: costGroup,
			Items:     itemList,
			Products:  string(productList),
		})
	}

	return stage2Inputs, nil
}

func createItemCostGroupMapping(response []stage1Output) map[string]string {
	mapping := make(map[string]string, len(response))
	for _, item := range response {
		mapping[item.ItemName] = item.CostGroup
	}
	return mapping
}

// TODO: error handling in this
// unmarshals the response into the provided type
func processResponse[TargetT any](response *genai.GenerateContentResponse) ([]TargetT, error) {
	var processedOutput []TargetT

	// log.Printf("response candidates: %d", len(response.Candidates))
	for _, c := range response.Candidates {
		for _, part := range c.Content.Parts {
			if txt, ok := part.(genai.Text); ok {
				var processedOutputPart []TargetT

				if err := json.Unmarshal([]byte(txt), &processedOutputPart); err != nil {
					log.Printf("Error happened unmarshalling part:\n%s\n", txt)
					continue
				}

				processedOutput = append(processedOutput, processedOutputPart...)
			}
		}
	}
	return processedOutput, nil
}

// helper function to interact with the llm
func generateContent(client *genai.Client, ctx context.Context, p prompt) *genai.GenerateContentResponse {
	// TODO: model type should be some kind of config parameter
	model := client.GenerativeModel("gemini-2.0-flash")
	// model.SetMaxOutputTokens(100)
	model.ResponseMIMEType = "application/json"

	model.ResponseSchema = p.outputFormat
	model.SystemInstruction = p.systemPrompt

	// model.SystemInstruction()
	resp, err := model.GenerateContent(ctx, p.taskPrompt, p.examples, p.inuptData)
	if err != nil {
		// TODO: This is probably bad, should return to caller
		log.Fatal(err)
	}
	return resp // helper function for printing content parts
}
