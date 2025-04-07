package main // Declares this as an executable program

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/totge/invoice-oc/go_item_categorizer/internal/receipt"
)

// Config struct to hold application configuration
type Config struct {
	GeminiAPIKey    string
	AppSheetAPIKey  string
	AppSheetAppID   string
	AppSheetBaseURL string // Optional, might have a default
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

	fmt.Println(config.AppSheetBaseURL)

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

	if allEnvVarsFound == false {
		return c, errors.New("missing config from environment")
	}

	c.AppSheetAPIKey = appSheetApiKey
	c.AppSheetAppID = appSheetAppId
	c.AppSheetBaseURL = appSheetBaseUrl
	c.GeminiAPIKey = geminiApiKey

	return c, nil
}
