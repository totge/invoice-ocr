package llmcategorizer

import (
	"embed"
	"fmt"
	"log"
	"strings"
	"text/template"

	"github.com/totge/invoice-oc/go_item_categorizer/internal/domain"
	"github.com/totge/invoice-oc/go_item_categorizer/internal/llm"
)

//go:embed templates/*.tmpl
var promptTemplates embed.FS // Embeds the templates directory content

type promptBuilder struct {
	// templates is an unexported field, encapsulating the state.
	templates *template.Template
}

var builder *promptBuilder

// init parses all templates when the package is loaded.
func init() {
	// Define helper functions available within templates
	funcMap := template.FuncMap{
		// add function: Allows {{ add $index 1 }} in templates for 1-based indexing
		"add": func(a, b int) int {
			return a + b
		},
		// join function: Allows {{ join .Path " / " }} in templates
		"join": func(s []string, sep string) string {
			// Basic protection against nil slice if needed
			if s == nil {
				return ""
			}
			return strings.Join(s, sep)
		},
	}

	var err error
	// var parsedTemplates *template.Template

	// Create a new template, add helper functions, then parse all embedded files matching the pattern.
	parsedTemplates, err := template.New("geminiPrompts"). // Give the template collection a name
								Funcs(funcMap).                              // Attach the helper functions
								ParseFS(promptTemplates, "templates/*.tmpl") // Parse from embedded FS

	if err != nil {
		// If templates fail to parse, the application cannot function correctly.
		// Using log.Fatalf ensures the error is printed and the app exits.
		log.Fatalf("FATAL: Failed to parse prompt templates: %v", err)
	}

	log.Println("Gemini prompt templates loaded and parsed successfully.")

	builder = &promptBuilder{templates: parsedTemplates}

}

// buildStage1Prompt creates the complete, agnostic prompt for the cost group assignment task.
func (pb *promptBuilder) buildStage1Prompt(items []domain.Item, allCostGroups []string) (llm.Prompt, error) {
	// The data structure passed to the template for rendering.
	templateData := struct {
		Items      []domain.Item
		CostGroups []string
	}{
		Items:      items,
		CostGroups: allCostGroups,
	}

	// Render each part of the prompt individually.
	taskPart, err := pb.render("stage1_TASK.tmpl", nil)
	if err != nil {
		return llm.Prompt{}, err
	}
	inputPart, err := pb.render("stage1_INPUT.tmpl", templateData)
	if err != nil {
		return llm.Prompt{}, err
	}
	examplesPart, err := pb.render("stage1_EXAMPLES.tmpl", nil)
	if err != nil {
		return llm.Prompt{}, err
	}

	// Join the parts into a single Text field.
	var sb strings.Builder
	sb.WriteString(taskPart)
	sb.WriteString("\n\n")
	sb.WriteString(inputPart)
	sb.WriteString("\n\n")
	sb.WriteString(examplesPart)

	systemInstruction, err := pb.render("system_instructions.tmpl", nil)
	if err != nil {
		return llm.Prompt{}, fmt.Errorf("failed to render system instruction: %w", err)
	}

	// Assemble and return the final Prompt struct
	return llm.Prompt{
		SystemInstruction: systemInstruction,
		Text:              sb.String(),
		OutputSchema:      stage1Schema,
	}, nil
}

// buildStage2Prompt creates the prompt for the detailed product matching task.
func (pb *promptBuilder) buildStage2Prompt(costGroup string, itemsInGroup []domain.Item, productCandidates []domain.ProductClassification) (llm.Prompt, error) {
	taskTemplateData := struct {
		costGroup string
	}{
		costGroup: costGroup,
	}

	inputTemplateData := struct {
		Items      []domain.Item
		Candidates []domain.ProductClassification
	}{
		Items:      itemsInGroup,
		Candidates: productCandidates,
	}

	// Render each part of the prompt individually.
	taskPart, err := pb.render("stage2_TASK.tmpl", taskTemplateData)
	if err != nil {
		return llm.Prompt{}, err
	}
	inputPart, err := pb.render("stage2_INPUT.tmpl", inputTemplateData)
	if err != nil {
		return llm.Prompt{}, err
	}
	examplesPart, err := pb.render("stage2_EXAMPLES.tmpl", nil)
	if err != nil {
		return llm.Prompt{}, err
	}

	// Join the parts into a single Text field.
	var sb strings.Builder
	sb.WriteString(taskPart)
	sb.WriteString("\n\n")
	sb.WriteString(inputPart)
	sb.WriteString("\n\n")
	sb.WriteString(examplesPart)

	systemInstruction, err := pb.render("system_instructions.tmpl", nil)
	if err != nil {
		return llm.Prompt{}, fmt.Errorf("failed to render system instruction: %w", err)
	}

	// Assemble and return the final Prompt struct
	return llm.Prompt{
		SystemInstruction: systemInstruction,
		Text:              sb.String(),
		OutputSchema:      stage2Schema,
	}, nil
}

// helper method for executing a specific template.
func (pb *promptBuilder) render(templateName string, data any) (string, error) {
	// Lookup the specific template by its filename.
	tmpl := pb.templates.Lookup(templateName)
	if tmpl == nil {
		// This indicates a developer error
		return "", fmt.Errorf("template %q not found", templateName)
	}

	var sb strings.Builder
	// Execute the template, writing the output directly to the strings.Builder.
	if err := tmpl.Execute(&sb, data); err != nil {
		return "", fmt.Errorf("error executing template %q: %w", templateName, err)
	}

	return sb.String(), nil
}
