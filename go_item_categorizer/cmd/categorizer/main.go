package main // Declares this as an executable program

import (
	"fmt"
	"log"
	"os"

	"github.com/totge/invoice-oc/go_item_categorizer/internal/receipt"
)

// main is the entry point of the application
func main() {
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
