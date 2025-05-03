package llm

import (
	"bytes"
	"embed"
	"fmt"
	"log"
	"strings"
	"text/template"

	"github.com/google/generative-ai-go/genai"
	"github.com/totge/invoice-oc/go_item_categorizer/internal/receipt"
)

type PromptData struct {
	Task      string
	InputData Input
	Examples  []string
}

type Input struct {
	Items      []string
	Categories []string
}

//go:embed templates/*.tmpl
var promptTemplates embed.FS // Embeds the templates directory content

var parsedTemplates *template.Template

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
	// Create a new template, add helper functions, then parse all embedded files matching the pattern.
	parsedTemplates, err = template.New("geminiPrompts"). // Give the template collection a name
								Funcs(funcMap).                              // Attach the helper functions
								ParseFS(promptTemplates, "templates/*.tmpl") // Parse from embedded FS

	if err != nil {
		// If templates fail to parse, the application cannot function correctly.
		// Using log.Fatalf ensures the error is printed and the app exits.
		log.Fatalf("FATAL: Failed to parse prompt templates: %v", err)
	}

	log.Println("Gemini prompt templates loaded and parsed successfully.")
}

//TODO: only here for testing
func RenderTemplate() (prompt, error) {

	// templateData := PromptData{
	// 	Task: "Do something",
	// 	InputData: Input{
	// 		Items:      []string{"Narancs", "Sertés darálthús", "Tejes kifli"},
	// 		Categories: []string{"Élelmiszer", "Sport", "Háztartás"},
	// 	},
	// 	Examples: []string{
	// 		"Tejes kifli -> Élelmiszer",
	// 		"Tejhabosító gép -> Háztartás",
	// 	},
	// }

	stage1TemplateData := stage1Input{
		CostGroups: []string{
			"Élelmiszer",
			"Sport",
			"Háztartás",
		},
		Items: []receipt.Item{
			{
				Name: "TARTOS TEJ 2,8%",
				Price: 738,
				Discount: 0,
			},
			{
				Name: "ALMA, GALA KG",
				Price: 445,
				Discount: 0,
			},
			{
				Name: "NARANCS KG",
				Price: 593,
				Discount: 0,
			},
		},
	}

	// renderedTemplate, err := renderTemplate("template_test.tmpl", templateData)
	// if err != nil {
	// 	return "", err
	// }

	// renderedTemplate, err := renderTemplate("stage1_TASK.tmpl", struct{}{})
	// if err != nil {
	// 	return "", err
	// }

	// return renderedTemplate, nil

	var stage1promptstr string

	p, err := buildStage1Prompt(stage1TemplateData)
	if err != nil {
		return p, err
	}

	stage1promptstr += string(p.taskPrompt) + "\n"
	stage1promptstr += string(p.inuptData) + "\n"
	stage1promptstr += string(p.examples) + "\n"

	return p, nil

}

func buildStage1Prompt(inputData stage1Input) (prompt, error) {
	var assambledPrompt prompt

	systemPrompt, err := renderTemplate("system_instructions.tmpl", struct{}{})
	if err != nil {
		return assambledPrompt, err
	}

	taskPrompt, err := renderTemplate("stage1_TASK.tmpl", struct{}{})
	if err != nil {
		return assambledPrompt, err
	}

	inputPrompt, err := renderTemplate("stage1_INPUT.tmpl", inputData)
	if err != nil {
		return assambledPrompt, err
	}

	examplesPrompt, err := renderTemplate("stage1_EXAMPLES.tmpl", struct{}{})
	if err != nil {
		return assambledPrompt, err
	}

	assambledPrompt.systemPrompt = &genai.Content{Parts: []genai.Part{genai.Text(systemPrompt)}}
	assambledPrompt.taskPrompt = genai.Text(taskPrompt)
	assambledPrompt.inuptData = genai.Text(inputPrompt)
	assambledPrompt.examples = genai.Text(examplesPrompt)
	assambledPrompt.outputFormat = stage1OutputFormat

	return assambledPrompt, nil
}

func buildStage2Prompt(inputData stage2Input) (prompt, error) {
	var assambledPrompt prompt

	systemPrompt, err := renderTemplate("system_instructions.tmpl", struct{}{})
	if err != nil {
		return assambledPrompt, err
	}

	taskPrompt, err := renderTemplate("stage2_TASK.tmpl", inputData)
	if err != nil {
		return assambledPrompt, err
	}

	inputPrompt, err := renderTemplate("stage2_INPUT.tmpl", inputData)
	if err != nil {
		return assambledPrompt, err
	}

	examplesPrompt, err := renderTemplate("stage2_EXAMPLES.tmpl", struct{}{})
	if err != nil {
		return assambledPrompt, err
	}

	assambledPrompt.systemPrompt = &genai.Content{Parts: []genai.Part{genai.Text(systemPrompt)}}
	assambledPrompt.taskPrompt = genai.Text(taskPrompt)
	assambledPrompt.inuptData = genai.Text(inputPrompt)
	assambledPrompt.examples = genai.Text(examplesPrompt)
	assambledPrompt.outputFormat = stage2OutputFormat

	return assambledPrompt, nil
}

// renderTemplate executes a named template with the given data struct.
// templateName should be the base filename of the template (e.g., "stage1_cost_group.tmpl").
// data should be a pointer to the corresponding data struct (e.g., &stage1Data{}).
func renderTemplate(templateName string, data interface{}) (string, error) {
	// Safety check in case init() somehow failed silently (shouldn't happen with log.Fatalf)
	if parsedTemplates == nil {
		return "", fmt.Errorf("internal error: prompt templates not initialized")
	}

	// Lookup the specific template by its filename within the parsed collection.
	tmpl := parsedTemplates.Lookup(templateName)
	if tmpl == nil {
		// This indicates a programming error (e.g., typo in templateName)
		return "", fmt.Errorf("template '%s' not found in parsed templates", templateName)
	}

	// Create a buffer to capture the output of the template execution.
	var filledPrompt bytes.Buffer

	// Execute the template, writing the output to the buffer and passing the data struct.
	err := tmpl.Execute(&filledPrompt, data)
	if err != nil {
		// Error occurred during template execution (e.g., data mismatch, bad helper call)
		return "", fmt.Errorf("failed to execute template '%s': %w", templateName, err)
	}

	// Return the generated prompt string from the buffer.
	return filledPrompt.String(), nil
}