package llm

import (
	"context"
	"log"

	"github.com/google/generative-ai-go/genai"
)

func GenerateContent(client genai.Client, ctx context.Context) *genai.GenerateContentResponse {
	model := client.GenerativeModel("gemini-2.0-flash")
	// model.SetMaxOutputTokens(100)
	model.ResponseMIMEType = "application/json"

	// TODO: add these schemas as consts
	model.ResponseSchema = &genai.Schema{
		Type: genai.TypeObject,
		Properties: map[string]*genai.Schema{
			"item_name": &genai.Schema{
				Type:        genai.TypeString,
				Description: "original name of the item, exactly as it was provided in the input",
				Nullable:    false,
			},
			"cost_group": &genai.Schema{
				Type:        genai.TypeString,
				Description: "original name of best corresponding cost group, exactly as it was provided in the input",
				Nullable:    false,
			},
		},
	}

	// model.SystemInstruction()
	resp, err := model.GenerateContent(ctx, genai.Text("How does AI work?"))
	if err != nil {
		log.Fatal(err)
	}
	return resp // helper function for printing content parts
}
