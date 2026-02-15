# Invoice OCR

A CLI tool written in Go that extracts structured data from receipt images using LLM-based OCR and line item classification.

## How It Works

The tool runs a two-stage pipeline:

1. **Extraction** — A receipt image is sent to the Gemini API, which returns structured data (store name, items, prices, totals, timestamp).
2. **Categorization** — Each extracted item goes through a two-stage LLM classification:
   - *Stage 1*: Assigns items to broad cost groups to narrow the search space.
   - *Stage 2*: Matches items to specific products from a catalog (fetched from AppSheet), batched by cost group.

These stages can be run individually (`extract`, `categorize`) or together (`process`). A `list` command is also available to browse available source files.

Currently only Gemini is supported as LLM service.

## Project Structure

```
cmd/categorizer/          Entry point (thin main.go)
internal/
  cli/                    Cobra command definitions, options structs, validation
  app/                    Application orchestration layer and interface definitions
  domain/                 Core types (Receipt, Item, ProductClassification, etc.)
  ocrextractor/           OCR extraction via LLM
  llmcategorizer/         Two-stage LLM categorization pipeline
  geminiclient/           Gemini API client wrapper
  imagereader/            Reads images from local filesystem
  jsonfilereader/         Reads extracted receipt JSON files
  googledrive/            Google Drive file reader
  localfilelister/        Lists files from local directories
  appsheet/               AppSheet API client
  appsheetcatalog/        Fetches product catalog from AppSheet
  appsheetwriter/         Writes results to AppSheet
  jsonfilewriter/         Writes results to JSON files
  terminalwriter/         Writes results to terminal
  sourcefilter/           Filters file lists by criteria
  llm/                    LLM client interface
testdata/
  raw_receipts/           Sample receipt images
  parsed_receipts/        Sample extracted JSON files
```

## Getting Started

### Prerequisites

- Go 1.24+
- [Task](https://taskfile.dev/) runner (optional, but recommended)
- A Google Gemini API key
- An AppSheet API key and App ID (for categorization and AppSheet output)
- A Google Drive service account key file (only if using Google Drive as a source)

### Configuration

Create a `.env` file in the project root. All variables use the `INVOICE_CATEGORIZER_` prefix:

```env
INVOICE_CATEGORIZER_GEMINI_API_KEY=your-gemini-api-key
INVOICE_CATEGORIZER_GEMINI_MODEL=gemini-2.0-flash
INVOICE_CATEGORIZER_APPSHEET_API_KEY=your-appsheet-api-key
INVOICE_CATEGORIZER_APPSHEET_APP_ID=your-appsheet-app-id
INVOICE_CATEGORIZER_APPSHEET_BASE_URL=https://www.appsheet.com
INVOICE_CATEGORIZER_LOG_LEVEL=error
INVOICE_CATEGORIZER_GOOGLE_DRIVE_KEY_PATH=/path/to/service-account-key.json
```

All values can also be set via environment variables directly or overridden with CLI flags.

### Build

```bash
# Using Task (preferred)
task build

# Or directly
go build -o bin/categorizer ./cmd/categorizer
```

### Usage

```bash
# Extract data from a receipt image
./bin/categorizer extract -i receipt.png -o extracted.json

# Categorize items from an extracted JSON file
./bin/categorizer categorize -i extracted.json -t file -o categorized.json

# Run the full pipeline (extract + categorize) in one step
./bin/categorizer process -i receipt.png -t file -o result.json

# List available files from a local directory
./bin/categorizer list -p ./testdata/raw_receipts/

# List files from Google Drive
./bin/categorizer list -s gdrive -p <folder-id>
```

Use `--help` on any command for all available flags.

### Test

```bash
task test

# Or directly
go test -v ./...
```
