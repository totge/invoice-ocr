package ocrextractor

import (
	"embed"
	"fmt"
	"log/slog"
	"strings"
	"text/template"

	"github.com/totge/invoice-oc/go_item_categorizer/internal/domain"
	"github.com/totge/invoice-oc/go_item_categorizer/internal/llm"
)

//go:embed templates/*.tmpl
var templatesFS embed.FS

// Singleton instance
var builder *promptBuilder

type promptBuilder struct {
	templates *template.Template
}

func init() {
	// loading template files
	parsed, err := template.New("ocrPrompts").ParseFS(templatesFS, "templates/*.tmpl")
	if err != nil {
		panic(fmt.Sprintf("ocr: failed to parse prompt templates: %v", err))
	}
	builder = &promptBuilder{templates: parsed}
}

// buildOCRPrompt constructs the prompt for the receipt extraction.
// It combines the static text instructions with the dynamic image data.
func (pb *promptBuilder) buildOCRPrompt(image *domain.ImageSource) (llm.Prompt, error) {

	slog.Debug("Building OCR prompt",
		"image_format", image.Format,
		"image_size_bytes", len(image.Data),
	)

	// 1. Render the System Instruction
	systemInstruction, err := pb.render("system_instructions.tmpl", nil)
	if err != nil {
		return llm.Prompt{}, fmt.Errorf("failed to render system instruction: %w", err)
	}

	// 2. Render the Task Text
	taskText, err := pb.render("ocr_task.tmpl", nil)
	if err != nil {
		return llm.Prompt{}, fmt.Errorf("failed to render task prompt: %w", err)
	}

	// 3. Assemble the Content (Text + Image)
	content := []*llm.PromptContent{
		{
			ContentType: llm.ContentTypeText,
			Data:        []byte(taskText),
		},
		{
			ContentType: llm.ContentTypeImage,
			Format:      image.Format,
			Data:        image.Data,
		},
	}

	// 4. Return the final Prompt
	return llm.Prompt{
		SystemInstruction: systemInstruction,
		Content:           content,
		OutputSchema:      ocrSchema, // Global defined in types.go
	}, nil
}

func (pb *promptBuilder) render(templateName string, data any) (string, error) {
	tmpl := pb.templates.Lookup(templateName)
	if tmpl == nil {
		return "", fmt.Errorf("template %q not found", templateName)
	}

	var sb strings.Builder
	if err := tmpl.Execute(&sb, data); err != nil {
		return "", fmt.Errorf("error executing template %q: %w", templateName, err)
	}

	return sb.String(), nil
}
