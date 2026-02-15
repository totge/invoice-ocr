package geminiclient

import (
	"testing"

	"github.com/totge/invoice-oc/go_item_categorizer/internal/llm"
)

func TestTranslatePrompt(t *testing.T) {
	c := &Client{}

	t.Run("text only", func(t *testing.T) {
		prompt := llm.Prompt{
			Content: []*llm.PromptContent{
				{ContentType: llm.ContentTypeText, Data: []byte("hello world")},
			},
		}
		result, err := c.translatePrompt(prompt)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result) != 1 {
			t.Fatalf("expected 1 content block, got %d", len(result))
		}
		parts := result[0].Parts
		if len(parts) != 1 {
			t.Fatalf("expected 1 part, got %d", len(parts))
		}
		if parts[0].Text != "hello world" {
			t.Errorf("expected text %q, got %q", "hello world", parts[0].Text)
		}
	})

	t.Run("image only", func(t *testing.T) {
		imageData := []byte{0xFF, 0xD8, 0xFF}
		prompt := llm.Prompt{
			Content: []*llm.PromptContent{
				{ContentType: llm.ContentTypeImage, Data: imageData, Format: "image/jpeg"},
			},
		}
		result, err := c.translatePrompt(prompt)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		parts := result[0].Parts
		if len(parts) != 1 {
			t.Fatalf("expected 1 part, got %d", len(parts))
		}
		if parts[0].InlineData == nil {
			t.Fatal("expected InlineData to be set")
		}
		if parts[0].InlineData.MIMEType != "image/jpeg" {
			t.Errorf("expected MIME type %q, got %q", "image/jpeg", parts[0].InlineData.MIMEType)
		}
		if len(parts[0].InlineData.Data) != 3 {
			t.Errorf("expected 3 bytes of image data, got %d", len(parts[0].InlineData.Data))
		}
	})

	t.Run("mixed text and image", func(t *testing.T) {
		prompt := llm.Prompt{
			Content: []*llm.PromptContent{
				{ContentType: llm.ContentTypeText, Data: []byte("describe this image")},
				{ContentType: llm.ContentTypeImage, Data: []byte{0x89, 0x50}, Format: "image/png"},
			},
		}
		result, err := c.translatePrompt(prompt)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		parts := result[0].Parts
		if len(parts) != 2 {
			t.Fatalf("expected 2 parts, got %d", len(parts))
		}
		if parts[0].Text != "describe this image" {
			t.Errorf("first part should be text, got %q", parts[0].Text)
		}
		if parts[1].InlineData == nil {
			t.Error("second part should have InlineData")
		}
	})

	t.Run("unknown content type returns error", func(t *testing.T) {
		prompt := llm.Prompt{
			Content: []*llm.PromptContent{
				{ContentType: "video", Data: []byte("data")},
			},
		}
		_, err := c.translatePrompt(prompt)
		if err == nil {
			t.Fatal("expected error for unknown content type")
		}
	})

	t.Run("empty content list", func(t *testing.T) {
		prompt := llm.Prompt{Content: []*llm.PromptContent{}}
		result, err := c.translatePrompt(prompt)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		parts := result[0].Parts
		if len(parts) != 0 {
			t.Errorf("expected 0 parts for empty content, got %d", len(parts))
		}
	})
}
