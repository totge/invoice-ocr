package llm

import (
	"context"
	"encoding/json"
)

type Prompt struct {
	SystemInstruction string
	Text              string
	OutputSchema      json.RawMessage
}

// Client is the agnostic interface that all LLM providers must implement.
type Client interface {
	GenerateJSON(ctx context.Context, model string, prompt Prompt) (string, error)
}
