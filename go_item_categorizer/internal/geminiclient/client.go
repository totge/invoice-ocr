package geminiclient

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/generative-ai-go/genai"
	"github.com/totge/invoice-oc/go_item_categorizer/internal/llm"
	"google.golang.org/api/option"
)

type Client struct {
	genaiClient *genai.Client
}

var _ llm.Client = (*Client)(nil)

func (c *Client) GenerateJSON(ctx context.Context, model string, prompt llm.Prompt) (string, error) {
	geminiModel := c.genaiClient.GenerativeModel(model)
	geminiModel.ResponseMIMEType = "application/json" // Enforce JSON output

	// --- TRANSLATION LOGIC ---
	geminiModel.SystemInstruction = &genai.Content{Parts: []genai.Part{genai.Text(prompt.SystemInstruction)}}


	// TODO: maybe separate the translation to a separate methode so that it is easily testable
	if prompt.OutputSchema != nil {
		var genaiSchema genai.Schema

		err := json.Unmarshal(prompt.OutputSchema, &genaiSchema)
		if err != nil {
			// This indicates a developer error (the JSON schema string is malformed).
			return "", fmt.Errorf("failed to unmarshal app's internal json schema into genai.Schema: %w", err)
		}

		// Assign the successfully translated schema to the model.
		geminiModel.ResponseSchema = &genaiSchema
	}

	// Convert our message history into genai.Content format.
	// The genai library wants a flat list of Parts for its GenerateContent call.
	// We will build this flat list from our structured messages.
	genaiPrompt := genai.Text(prompt.Text)

	resp, err := geminiModel.GenerateContent(ctx, genaiPrompt)
	if err != nil {
		return "", fmt.Errorf("gemini generation failed: %w", err)
	}

	// Extract the raw text from the response.
	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("gemini returned no content")
	}

	// TODO: Verify behavior, is it only returns 1 part with the text?
	if textPart, ok := resp.Candidates[0].Content.Parts[0].(genai.Text); ok {
		return string(textPart), nil
	}

	return "", fmt.Errorf("gemini response did not contain text")
}

// func (c *Client) translateOutputSchema(outputSchema json.RawMessage) (*genai.Schema, error) {
// 	output

// }

func New(ctx context.Context, apiKey string) (*Client, error) {
	var client Client
	geminiClient, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, err
	}

	client.genaiClient = geminiClient

	return &client, nil
}
