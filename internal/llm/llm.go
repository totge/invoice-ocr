package llm

import (
	"context"
	"encoding/json"
)

type Prompt struct {
	SystemInstruction string
	Content           []*PromptContent
	OutputSchema      json.RawMessage
}

type PromptContent struct {
	ContentType ContentType
	Format      string
	Data        []byte
}

type ContentType string

const (
	ContentTypeText  ContentType = "text"
	ContentTypeImage ContentType = "image"
)

type ImageFormat string

const (
	ImageFormatJPEG = "image/jpeg"
	ImageFormatPNG  = "image/png"
)

// Client is the agnostic interface that all LLM providers must implement.
type Client interface {
	GenerateJSON(ctx context.Context, model string, prompt Prompt) (string, error)
}
