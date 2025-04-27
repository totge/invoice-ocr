package llm

import (
	"context"
	"log"

	"github.com/google/generative-ai-go/genai"
)




func generateContent(client genai.Client, ctx context.Context, p prompt) *genai.GenerateContentResponse {
	model := client.GenerativeModel("gemini-2.0-flash")
	// model.SetMaxOutputTokens(100)
	model.ResponseMIMEType = "application/json"

	model.ResponseSchema = p.outputFormat
	model.SystemInstruction = p.systemPrompt

	// model.SystemInstruction()
	resp, err := model.GenerateContent(ctx, p.taskPrompt, p.examples, p.inuptData)
	if err != nil {
		log.Fatal(err)
	}
	return resp // helper function for printing content parts
}
