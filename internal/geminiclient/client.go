package geminiclient

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/totge/invoice-oc/go_item_categorizer/internal/llm"
	"google.golang.org/genai"
)

type Client struct {
	genaiClient *genai.Client
}

var _ llm.Client = (*Client)(nil)

// TODO: is response content validation needed? It seems a bit manual validation step i don't know if its reliable enough
// TODO: Think about testability here, feels like this method does a lot
func (c *Client) GenerateJSON(ctx context.Context, model string, prompt llm.Prompt) (string, error) {

	slog.Debug("Generating content with Gemini",
		"model", model,
		"content_parts", len(prompt.Content),
	)

	config := &genai.GenerateContentConfig{
		ResponseMIMEType: "application/json",
	}

	// Setting system prompt for Gemini model
	if prompt.SystemInstruction != "" {
		config.SystemInstruction = &genai.Content{
			Parts: []*genai.Part{{Text: prompt.SystemInstruction}},
		}
	}

	// Setting output format
	if prompt.OutputSchema != nil {
		var genaiSchema genai.Schema

		if err := json.Unmarshal(prompt.OutputSchema, &genaiSchema); err != nil {
			return "", fmt.Errorf("failed to unmarshal app's internal json schema into genai.Schema: %w", err)
		}
		config.ResponseSchema = &genaiSchema

		slog.Debug("Applied JSON schema validation to Gemini request")
	}

	// Convert internal Prompt type into genai.Content format.
	genaiPrompt, err := c.translatePrompt(prompt)
	if err != nil {
		return "", fmt.Errorf("failed to convert prompt to genai types: %w", err)
	}

	// Generate content by the LLM
	resp, err := c.genaiClient.Models.GenerateContent(ctx, model, genaiPrompt, config)
	if err != nil {
		return "", fmt.Errorf("gemini generation failed: %w", err)
	}

	// Extract the raw text from the response.
	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("gemini returned no content")
	}

	text := resp.Text()

	slog.Debug("Gemini response received", "response_length", len(text))
	slog.Debug("Token usage", "prompt_tokens", resp.UsageMetadata.PromptTokenCount)

	return text, nil

}

func (c *Client) translatePrompt(prompt llm.Prompt) ([]*genai.Content, error) {
	parts := make([]*genai.Part, 0, len(prompt.Content))

	for _, c := range prompt.Content {
		switch c.ContentType {
		case llm.ContentTypeText:
			parts = append(parts, &genai.Part{Text: string(c.Data)})
		case llm.ContentTypeImage:
			parts = append(parts, &genai.Part{InlineData: &genai.Blob{Data: c.Data, MIMEType: c.Format}})
		default:
			return nil, fmt.Errorf("unable to handle prompt type %s", c.ContentType)
		}
	}

	return []*genai.Content{{Parts: parts}}, nil

}

func New(ctx context.Context, apiKey string) (*Client, error) {

	slog.Debug("Initializing Gemini client")

	var client Client

	geminiClient, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, err
	}

	client.genaiClient = geminiClient

	return &client, nil
}
