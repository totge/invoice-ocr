package llm

import (
	"encoding/json"
	"log"

	"github.com/google/generative-ai-go/genai"
)

// TODO: This is meant to be the high level api to interact with this package
// TODO: It should only call functions and methods and not do any other work
// TODO: Anything that is more than that should have its delegated, testable function or method
func AssignCategoryData(client ContentGenerator, items []ItemInfo, products ProductInfo) (map[string]CategoryHierarchy, error) {

	// --- 1. Prepare the inputs and prompt for stage1
	costGroups := products.GetCostGroups()

	stg1Input := stage1Input{
		CostGroups: costGroups,
		Items:      items,
	}

	stg1Prompt, err := buildStage1Prompt(stg1Input)
	if err != nil {
		return nil, err
	}

	// --- 2. Generate the llm output for stage2
	// stg1Resp := generateContent(client, ctx, stg1Prompt)
	stg1Resp, err := client.GenerateContent(stg1Prompt)
	if err != nil {
		return nil, err
	}

	// --- 3. Process the stage 1 output
	processedResp, err := processResponse[stage1Output](stg1Resp)
	if err != nil {
		return nil, err
	}

	// --- 4. Prepare the stage 2 inputs
	stage2InputList, err := createStage2Input(items, products, processedResp)
	if err != nil {
		return nil, err
	}

	// --- 5. Make the prompt and generate llm output for all stage 2 input
	enrichedItems := make([]stage2Output, 0, len(items))
	// run stage two for each stage2 input
	for _, input := range stage2InputList {
		stage2prompt, err := buildStage2Prompt(input)
		if err != nil {
			return nil, err
		}

		stg2Resp, err := client.GenerateContent(stage2prompt)
		if err != nil {
			return nil, err
		}
		processedResp, err := processResponse[stage2Output](stg2Resp)
		if err != nil {
			return nil, err
		}
		enrichedItems = append(enrichedItems, processedResp...)
	}

	// --- 6. Build the final return value from the stage2 output
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

func createStage2Input(items []ItemInfo, products ProductInfo, procesedStage1output []stage1Output) ([]stage2Input, error) {

	costGroupMapping := createItemCostGroupMapping(procesedStage1output)

	stage2GroupedItems := make(map[string][]ItemInfo)

	// grouping original items by cost group
	for _, item := range items {
		itemCostGroup := costGroupMapping[item.GetName()]

		// adding slice for new cost groups
		if itemList, ok := stage2GroupedItems[itemCostGroup]; ok {
			stage2GroupedItems[itemCostGroup] = append(itemList, item)
		} else {
			stage2GroupedItems[itemCostGroup] = []ItemInfo{item}
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
