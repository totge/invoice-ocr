package llm

import (
	"context"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

type ContentGenerator interface {
	GenerateContent(Prompt) (*genai.GenerateContentResponse, error)
}

type Client struct {
	geminiClient *genai.Client
	ctx          context.Context
}

func (c *Client) GenerateContent(p Prompt) (*genai.GenerateContentResponse, error) {
	model := c.geminiClient.GenerativeModel("gemini-2.0-flash")
	// model.SetMaxOutputTokens(100)
	model.ResponseMIMEType = "application/json"
	model.ResponseSchema = p.OutputFormat
	model.SystemInstruction = p.SystemPrompt

	// model.SystemInstruction()
	resp, err := model.GenerateContent(c.ctx, p.TaskPrompt, p.Examples, p.InuptData)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func NewClient(ctx context.Context, apiKey string) (*Client, error) {
	var client Client
	geminiClient, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, err
	}

	client.geminiClient = geminiClient
	client.ctx = ctx

	return &client, nil
}
