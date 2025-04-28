package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/google/generative-ai-go/genai"
	"github.com/totge/invoice-oc/go_item_categorizer/internal/catalog"
	"github.com/totge/invoice-oc/go_item_categorizer/internal/receipt"
)

func AssignCategoryData(client *genai.Client, ctx context.Context, items []receipt.Item, products catalog.ProductCatalog) (CategorizedReceipt, error) {
	var enrichedReceipt CategorizedReceipt

	costGorups := products.GetCostGroups()

	stg1Input := stage1Input{
		CostGroups: costGorups,
		Items:      items,
	}

	// TODO: this is the actual code
	stg1Prompt, err := buildStage1Prompt(stg1Input)
	if err != nil {
		return enrichedReceipt, err
	}

	stg1Resp := generateContent(client, ctx, stg1Prompt)

	processedResp, err := processResponse[stage1Output](stg1Resp)
	if err != nil {
		return enrichedReceipt, err
	}

	fmt.Println(processedResp)

	// process response -> create cateorized item list from

	return enrichedReceipt, nil
}

// TODO: error handling in this
func processResponse[TargetT any](response *genai.GenerateContentResponse) ([]TargetT, error) {
	var processedOutput []TargetT

	log.Printf("response candidates: %d", len(response.Candidates))
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
