package main // Declares this as an executable program

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/totge/invoice-oc/go_item_categorizer/internal/llm"
	"github.com/totge/invoice-oc/go_item_categorizer/internal/receipt"
)

// Config struct to hold application configuration
type Config struct {
	GeminiApiKey    string
	AppSheetApiKey  string
	AppSheetAppId   string
	AppSheetBaseUrl string // Optional, might have a default
}

// main is the entry point of the application
func main() {
	fmt.Println("Initializing application...")

	// loadig environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
		os.Exit(1)
	}

	config, err := loadConfig()
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	fmt.Println(config.AppSheetBaseUrl)

	fmt.Println("Receipt Categorizer Starting...")

	// --- File Reading Responsibility (in main) ---
	// 1. Get file path (e.g., hardcoded for now, later from flags)
	filePath := "./testdata/16000333862025032623918.json" // Relative to execution dir

	// 2. Read the file content
	log.Printf("Reading receipt data from: %s\n", filePath)
	jsonData, err := os.ReadFile(filePath) // Reads the whole file into memory
	if err != nil {
		log.Fatalf("FATAL: Failed to read file %s: %v", filePath, err)
	}
	// --- End File Reading ---

	// --- Parsing Responsibility (call internal package) ---
	log.Println("Parsing receipt data...")
	// 3. Pass the data to the dedicated parser function
	parsedReceipt, err := receipt.ParseReceipt(jsonData)
	if err != nil {
		log.Fatalf("FATAL: Failed to parse receipt data from %s: %v", filePath, err)
	}
	// --- End Parsing ---

	log.Printf("Successfully parsed receipt from %s with %d items.\n",
		parsedReceipt.Timestamp, len(parsedReceipt.Items)) // Adjust field names based on your struct

	// ... Next steps: Process parsedReceipt ...

	// // --- 3. Initialize AppSheet Client ---
	// log.Println("Initializing AppSheet client...")
	// appsheetClient, err := appsheet.NewClient(config.AppSheetBaseUrl, config.AppSheetAppId, config.AppSheetApiKey)
	// if err != nil {
	// 	log.Fatalf("FATAL: Failed to create AppSheet client: %v", err)
	// }
	// log.Println("AppSheet client initialized.")

	// // --- 4. Create Context for API Calls ---
	// // Use a background context for now, or add a timeout if needed
	// // Example: 2 minute timeout for fetching *all* initial AppSheet data
	// ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	// defer cancel() // Important: release context resources when main exits

	// log.Println("Building product catalog")
	// productCatalog, err := catalog.BuildProductCatalog(ctx, appsheetClient)
	// if err != nil {
	// 	log.Println("Error happend when building the catalog")
	// }
	// log.Printf("Catalog built with %d cost groups\n", len(productCatalog))

	// --- 5. Prompt genai to get the categorized items
	// ctx = context.Background()
	// geminiClient, err := genai.NewClient(ctx, option.WithAPIKey(config.GeminiApiKey))
	// if err != nil {
	// 	log.Fatalf("FATAL: Failed to create Gemini client: %v", err)
	// }
	// resp := llm.GenerateContent(*geminiClient, ctx)

	// for _, cand := range resp.Candidates {
	// 	if cand.Content != nil {
	// 		for _, part := range cand.Content.Parts {
	// 			fmt.Println(part)
	// 		}
	// 	}
	// }
	// fmt.Println("---")

	renderedTemplate, err := llm.RenderTemplate()
	if err != nil {
		log.Fatalf("FATAL: Failed to render template: %v", err)
	}

	log.Printf("Rendered template:\n\n%s\n", renderedTemplate)

}

func loadConfig() (Config, error) {
	allEnvVarsFound := true
	var c Config

	geminiApiKey := os.Getenv("GEMINI_API_KEY")
	if geminiApiKey == "" {
		log.Fatal("FATAL: GEMINI_API_KEY environment variable not set.")
		allEnvVarsFound = false
	}

	appSheetApiKey := os.Getenv("APPSHEET_API_KEY")
	if appSheetApiKey == "" {
		fmt.Println("FATAL: APPSHEET_API_KEY environment variable not set.")
		allEnvVarsFound = false
	}

	appSheetAppId := os.Getenv("APPSHEET_APP_ID")
	if appSheetAppId == "" {
		fmt.Println("FATAL: APPSHEET_APP_ID environment variable not set.")
		allEnvVarsFound = false
	}

	appSheetBaseUrl := os.Getenv("APPSHEET_BASE_URL")
	if appSheetBaseUrl == "" {
		fmt.Println("FATAL: APPSHEET_BASE_URL environment variable not set.")
		allEnvVarsFound = false
	}

	if !allEnvVarsFound {
		return c, errors.New("missing config from environment")
	}

	c.AppSheetApiKey = appSheetApiKey
	c.AppSheetAppId = appSheetAppId
	c.AppSheetBaseUrl = appSheetBaseUrl
	c.GeminiApiKey = geminiApiKey

	return c, nil
}
