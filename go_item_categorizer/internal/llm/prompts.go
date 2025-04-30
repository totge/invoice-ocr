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

// // getSystemInstructions formats the system prompt.
// // Currently takes no dynamic data, but could be adapted if needed.
// func getSystemInstructions() (string, error) {
// 	// The system instructions template currently doesn't require dynamic data.
// 	// Pass nil as the data argument to formatPrompt.
// 	prompt, err := formatPrompt("system_instructions.tmpl", nil)
// 	if err != nil {
// 		// Wrap the error for context
// 		return "", fmt.Errorf("failed to format system instructions: %w", err)
// 	}
// 	return prompt, nil
// }

// func formatCostGroupPrompt(items []receipt.Item, allCostGroups []string) (string, error) {
// 	// Basic input validation
// 	if len(items) == 0 {
// 		return "", fmt.Errorf("cannot format cost group prompt: no items provided")
// 	}
// 	if len(allCostGroups) == 0 {
// 		return "", fmt.Errorf("cannot format cost group prompt: no cost groups provided")
// 	}

// 	// Create and populate the data structure needed by the template.
// 	data := stage1Data{
// 		Items:      items,
// 		CostGroups: allCostGroups,
// 	}

// 	// Execute the template using the helper function.
// 	prompt, err := formatPrompt("stage1_cost_group.tmpl", &data) // Pass pointer to data struct
// 	if err != nil {
// 		// Wrap the error
// 		return "", fmt.Errorf("failed to format cost group prompt: %w", err)
// 	}
// 	return prompt, nil
// }

// // formatDetailedMatchPrompt prepares the data and formats the prompt for Stage 2.
// func formatDetailedMatchPrompt(item receipt.Item, assignedCostGroup string, candidates []catalog.ProductClassification) (string, error) {
// 	// Basic input validation
// 	if item.Name == "" { // Assuming Name is the description field
// 		return "", fmt.Errorf("cannot format detailed match prompt: item description is empty")
// 	}
// 	if assignedCostGroup == "" {
// 		return "", fmt.Errorf("cannot format detailed match prompt: assigned cost group is empty")
// 	}
// 	// Note: Allowing empty candidates might be valid if a cost group has no products yet.
// 	// The template handles the display, the caller handles the logic if no candidates exist.
// 	// if len(candidates) == 0 {
// 	// 	return "", fmt.Errorf("cannot format detailed match prompt: no candidate products provided for cost group '%s'", assignedCostGroup)
// 	// }

// 	// Create and populate the data structure for the template.
// 	data := stage2Data{
// 		Item:              item,
// 		AssignedCostGroup: assignedCostGroup,
// 		Candidates:        candidates,
// 	}

// 	// Execute the template.
// 	prompt, err := formatPrompt("stage2_detailed_match.tmpl", &data) // Pass pointer
// 	if err != nil {
// 		// Wrap the error
// 		return "", fmt.Errorf("failed to format detailed match prompt: %w", err)
// 	}
// 	return prompt, nil
// }

// // AssignCostGroups takes a list of receipt items and all known cost groups,
// // prompts the LLM to assign a cost group to each item, and returns a map
// // mapping the original item description to the assigned cost group name.
// // TODO: how should the resoults be returned
// func AssignCostGroups(client genai.Client, ctx context.Context, items []receipt.Item, allCostGroups []string) (map[string]string, error) {
// 	log.Printf("Starting Stage 1: Assigning Cost Groups for %d items...", len(items))

// 	if len(items) == 0 {
// 		return nil, fmt.Errorf("no items provided to assign cost groups")
// 	}
// 	if len(allCostGroups) == 0 {
// 		return nil, fmt.Errorf("no cost groups provided for assignment")
// 	}

// 	// 1. Get prompts
// 	systemPrompt, err := getSystemInstructions()
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to get system instructions: %w", err)
// 	}
// 	userPrompt, err := formatCostGroupPrompt(items, allCostGroups)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to format cost group prompt: %w", err)
// 	}

// 	// 2. --- PSEUDOCODE: Call Gemini API ---

// 	model := client.GenerativeModel(string(Gemini20Flash))
// 	resp, err := model.GenerateContent(ctx, genai.Text(systemPrompt), genai.Text(userPrompt))
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	res
// 	// resp, err := c.generateContentWithHistory(ctx, systemPrompt, userPrompt)
// 	// This helper would:
// 	// - Use c.sdkClient (your initialized SDK client)
// 	// - Construct the appropriate input for the SDK's GenerateContent method,
// 	//   likely involving creating []*sdk.Content parts with roles "user", "model", "user"
// 	//   using the systemPrompt and userPrompt strings.
// 	// - Set generation config (temperature, etc.) if needed.
// 	// - Call the SDK's GenerateContent method.
// 	// - Return the SDK's response object and any error.
// 	// Replace placeholder below with your actual implementation:
// 	var respText string // Placeholder for the text content extracted from response
// 	var apiErr error    // Placeholder for API error
// 	// Example call structure (needs real SDK types/methods):
// 	// sdkResp, apiErr := c.callGeminiSDK(ctx, systemPrompt, userPrompt)
// 	// if apiErr == nil {
// 	//     respText = extractTextFromResponse(sdkResp) // Use helper from previous steps
// 	// }
// 	respText = `{"T/P TEJ 1.5L": "Élelmiszer", "NIVEA TUSFÜRDŐ": "Drogéria", "CSIRK£M€LL FILÉ": "Élelmiszer"}` // !! REPLACE WITH ACTUAL API CALL & TEXT EXTRACTION !!
// 	apiErr = nil                                                                                               // !! REPLACE !!

// 	// --- End Pseudocode ---

// 	if apiErr != nil {
// 		return nil, fmt.Errorf("gemini API call for cost groups failed: %w", apiErr)
// 	}

// 	if respText == "" {
// 		return nil, fmt.Errorf("gemini API returned empty response for cost groups")
// 	}

// 	// 3. Parse the JSON response
// 	costGroupMap, err := parseCostGroupResponse(respText)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to parse cost group assignment response: %w", err)
// 	}

// 	// 4. Optional: Validate response keys match input item names? (Could be complex)
// 	log.Printf("Successfully assigned cost groups to %d items.", len(costGroupMap))
// 	return costGroupMap, nil
// }

// // FindBestMatch takes a single receipt item, its assigned cost group, and a list
// // of candidate products within that group. It prompts the LLM to select the best
// // match and returns a pointer to the chosen catalog.ProductKnowledge entry,
// // or nil if no suitable match was chosen (choice '0').
// func (c *Client) FindBestMatch(ctx context.Context, item receipt.Item, assignedCostGroup string, candidates []catalog.ProductKnowledge) (*catalog.ProductKnowledge, error) {
// 	log.Printf("Starting Stage 2: Finding best match for item '%s' in cost group '%s'...", item.Name, assignedCostGroup)

// 	if item.Name == "" {
// 		return nil, fmt.Errorf("item name is empty")
// 	}
// 	// Allow empty candidates? Or handle upstream? If handled here:
// 	if len(candidates) == 0 {
// 		log.Printf("No candidates provided for item '%s' in cost group '%s'. Returning no match.", item.Name, assignedCostGroup)
// 		return nil, nil // No candidates means no match possible
// 	}

// 	// 1. Get prompts
// 	systemPrompt, err := getSystemInstructions()
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to get system instructions: %w", err)
// 	}
// 	userPrompt, err := formatDetailedMatchPrompt(item, assignedCostGroup, candidates)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to format detailed match prompt: %w", err)
// 	}

// 	// 2. --- PSEUDOCODE: Call Gemini API ---
// 	// resp, err := c.generateContentWithHistory(ctx, systemPrompt, userPrompt)
// 	// Similar to above, use your SDK client and helper to make the call.
// 	// Replace placeholder below with your actual implementation:
// 	var respText string // Placeholder for the text content extracted from response
// 	var apiErr error    // Placeholder for API error
// 	respText = "1"      // !! REPLACE WITH ACTUAL API CALL & TEXT EXTRACTION !! Example output
// 	apiErr = nil        // !! REPLACE !!
// 	// --- End Pseudocode ---

// 	if apiErr != nil {
// 		return nil, fmt.Errorf("gemini API call for best match failed for item '%s': %w", item.Name, apiErr)
// 	}

// 	if respText == "" {
// 		return nil, fmt.Errorf("gemini API returned empty response for best match for item '%s'", item.Name)
// 	}

// 	// 3. Parse the numeric choice response
// 	choiceIndex, err := parseChoiceNumberResponse(respText)
// 	if err != nil {
// 		// Log the problematic response text
// 		log.Printf("Error parsing choice number response for item '%s'. Response: '%s', Error: %v", item.Name, respText, err)
// 		return nil, fmt.Errorf("failed to parse numeric choice response for item '%s': %w", item.Name, err)
// 	}

// 	// 4. Map choice index back to the candidate
// 	if choiceIndex == 0 {
// 		log.Printf("LLM indicated no suitable match (choice 0) for item '%s'.", item.Name)
// 		return nil, nil // Explicit "None" choice
// 	}

// 	// Adjust for 1-based indexing from the prompt vs 0-based slice index
// 	candidateIndex := choiceIndex - 1
// 	if candidateIndex < 0 || candidateIndex >= len(candidates) {
// 		log.Printf("Error: LLM returned out-of-bounds choice index %d for item '%s'. Max index: %d. Response: '%s'", choiceIndex, item.Name, len(candidates), respText)
// 		return nil, fmt.Errorf("LLM returned invalid choice index %d for item '%s'", choiceIndex, item.Name)
// 	}

// 	chosenCandidate := &candidates[candidateIndex]
// 	log.Printf("Found best match for item '%s': Product '%s'", item.Name, chosenCandidate.ProductName)
// 	return chosenCandidate, nil
// }

// // --- Helper functions for parsing responses ---

// // parseCostGroupResponse parses the JSON string expected from Stage 1.
// func parseCostGroupResponse(jsonString string) (map[string]string, error) {
// 	// Clean potential markdown fences
// 	jsonString = strings.TrimSpace(jsonString)
// 	jsonString = strings.TrimPrefix(jsonString, "```json")
// 	jsonString = strings.TrimSuffix(jsonString, "```")
// 	jsonString = strings.TrimSpace(jsonString)

// 	if jsonString == "" {
// 		return nil, fmt.Errorf("cannot parse empty string as JSON")
// 	}

// 	var result map[string]string
// 	err := json.Unmarshal([]byte(jsonString), &result)
// 	if err != nil {
// 		// Log the string that failed to parse for debugging
// 		log.Printf("DEBUG: Failed JSON string for cost group response: %s", jsonString)
// 		return nil, fmt.Errorf("failed to unmarshal cost group JSON: %w", err)
// 	}
// 	return result, nil
// }

// // parseChoiceNumberResponse parses the single number expected from Stage 2.
// func parseChoiceNumberResponse(text string) (int, error) {
// 	cleanedText := strings.TrimSpace(text)
// 	if cleanedText == "" {
// 		return -1, fmt.Errorf("cannot parse empty string as choice number")
// 	}

// 	// Attempt to parse as an integer
// 	choice, err := strconv.Atoi(cleanedText)
// 	if err != nil {
// 		return -1, fmt.Errorf("response '%s' is not a valid integer choice: %w", cleanedText, err)
// 	}

// 	return choice, nil
// }

// // --- Placeholder/Pseudocode for actual API call wrapper ---
// /*
// // generateContentWithHistory is an example *internal* helper wrapping the SDK
// func (c *Client) generateContentWithHistory(ctx context.Context, systemPrompt, userPrompt string) (*sdk.GenerateContentResponse, error) {
//     model := c.sdkClient.GenerativeModel(c.modelName)
//     // model.GenerationConfig = ... // Set config if needed

//     // Construct content using SDK types (e.g., sdk.Content, sdk.Text)
//     // This structure depends heavily on how the specific SDK handles system prompts.
//     requestContent := []*sdk.Content{
//         {Role: "user", Parts: []sdk.Part{sdk.Text(systemPrompt)}},
//         {Role: "model", Parts: []sdk.Part{sdk.Text("OK")}}, // Acknowledge system prompt
//         {Role: "user", Parts: []sdk.Part{sdk.Text(userPrompt)}},
//     }

//     resp, err := model.GenerateContent(ctx, requestContent...)
//     if err != nil {
//         // TODO: Check for specific SDK error types if available
//         return nil, fmt.Errorf("sdk GenerateContent failed: %w", err)
//     }
//     // TODO: Check resp.PromptFeedback for safety blocks/errors
//     return resp, nil
// }

// // extractTextFromResponse extracts plain text from the SDK response object
// func extractTextFromResponse(resp *sdk.GenerateContentResponse) string {
//     // Implementation depends on the SDK's response structure
//     // Usually involves iterating resp.Candidates[0].Content.Parts
//     // ...
//     return "extracted text placeholder"
// }
// */
