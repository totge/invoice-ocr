package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/totge/invoice-oc/go_item_categorizer/internal/app"
	"github.com/totge/invoice-oc/go_item_categorizer/internal/appsheet"
	"github.com/totge/invoice-oc/go_item_categorizer/internal/appsheetcatalog"
	"github.com/totge/invoice-oc/go_item_categorizer/internal/appsheetwriter"
	"github.com/totge/invoice-oc/go_item_categorizer/internal/geminiclient"
	"github.com/totge/invoice-oc/go_item_categorizer/internal/googledrive"
	"github.com/totge/invoice-oc/go_item_categorizer/internal/imagereader"
	"github.com/totge/invoice-oc/go_item_categorizer/internal/jsonfilewriter"
	"github.com/totge/invoice-oc/go_item_categorizer/internal/llmcategorizer"
	"github.com/totge/invoice-oc/go_item_categorizer/internal/ocrextractor"
)

// NewProcessCmd creates the "process" subcommand for end-to-end receipt processing.
func NewProcessCmd(v *viper.Viper) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "process",
		Short: "Run the full extraction + categorization pipeline",
		Long:  `Process reads a receipt image, extracts data via OCR, categorizes items, and writes the results.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runProcess(cmd, v)
		},
	}

	cmd.Flags().StringP("source", "s", "local", "input source: 'local' or 'gdrive'")
	cmd.Flags().StringP("input", "i", "", "path to the receipt image")
	cmd.Flags().StringP("target", "t", "file", "output target: 'file' or 'appsheet'")
	cmd.Flags().StringP("output", "o", "output.json", "output file path (used when target is 'file')")

	cmd.MarkFlagRequired("input")

	return cmd
}

func runProcess(cmd *cobra.Command, v *viper.Viper) error {
	// Read flags
	inputPath, _ := cmd.Flags().GetString("input")
	source, _ := cmd.Flags().GetString("source")
	outputTarget, _ := cmd.Flags().GetString("target")
	outputPath, _ := cmd.Flags().GetString("output")

	// Validate source
	switch source {
	case "local":
		info, err := os.Stat(inputPath)
		if os.IsNotExist(err) {
			return fmt.Errorf("input file does not exist: %s", inputPath)
		} else if err != nil {
			return fmt.Errorf("failed to check input file %s: %w", inputPath, err)
		} else if info.IsDir() {
			return fmt.Errorf("input path is a directory: %s", inputPath)
		}
	case "gdrive":
		// no preemptive validation
	default:
		return fmt.Errorf("invalid --source: %s (must be 'local' or 'gdrive')", source)
	}

	// Validate target
	switch outputTarget {
	case "file", "appsheet":
		// valid
	default:
		return fmt.Errorf("invalid --target: %s (must be 'file' or 'appsheet')", outputTarget)
	}

	if outputTarget == "file" {
		if outputPath == "" {
			return fmt.Errorf("--output is required when target is 'file'")
		}
		if inputPath == outputPath {
			return fmt.Errorf("--input and --output paths cannot be the same")
		}
	}

	// Read config from viper
	geminiApiKey := v.GetString("gemini-api-key")
	geminiModel := v.GetString("gemini-model")
	gdriveKeyPath := v.GetString("google-drive-key-path")
	appsheetApiKey := v.GetString("appsheet-api-key")
	appsheetAppId := v.GetString("appsheet-app-id")
	appsheetBaseUrl := v.GetString("appsheet-base-url")

	if geminiApiKey == "" {
		return fmt.Errorf("Gemini API key is required (set INVOICE_CATEGORIZER_GEMINI_API_KEY or --gemini-api-key)")
	}
	if geminiModel == "" {
		geminiModel = "gemini-2.0-flash"
	}
	if appsheetApiKey == "" {
		return fmt.Errorf("AppSheet API key is required (set INVOICE_CATEGORIZER_APPSHEET_API_KEY or --appsheet-api-key)")
	}
	if appsheetAppId == "" {
		return fmt.Errorf("AppSheet App ID is required (set INVOICE_CATEGORIZER_APPSHEET_APP_ID or --appsheet-app-id)")
	}

	slog.Debug("Process command configuration",
		"source", source,
		"input", inputPath,
		"target", outputTarget,
		"output", outputPath,
		"model", geminiModel,
	)

	ctx, cancel := context.WithTimeout(cmd.Context(), 3*time.Minute)
	defer cancel()

	// Construct dependencies

	// Input reader
	var reader app.ReceiptImageReader
	switch source {
	case "local":
		reader = imagereader.NewReader(inputPath)
	case "gdrive":
		var err error
		reader, err = googledrive.New(ctx, gdriveKeyPath, inputPath)
		if err != nil {
			return fmt.Errorf("failed to initialize Google Drive reader: %w", err)
		}
	}

	// LLM client + extractor
	llmClient, err := geminiclient.New(ctx, geminiApiKey)
	if err != nil {
		return fmt.Errorf("failed to initialize Gemini client: %w", err)
	}
	extractor := ocrextractor.New(llmClient, geminiModel)

	// Catalog
	appsheetClient, err := appsheet.NewClient(appsheetBaseUrl, appsheetAppId, appsheetApiKey)
	if err != nil {
		return fmt.Errorf("failed to initialize AppSheet client: %w", err)
	}
	lister := appsheetcatalog.New(appsheetClient)

	// Categorizer
	categorizer := llmcategorizer.New(llmClient, geminiModel)

	// Result writer
	var writer app.ResultWriter
	switch outputTarget {
	case "file":
		w, err := jsonfilewriter.New(outputPath)
		if err != nil {
			return fmt.Errorf("failed to initialize json file writer: %w", err)
		}
		writer = w
	case "appsheet":
		writer = appsheetwriter.New(appsheetClient)
	}

	return app.Process(ctx, reader, extractor, lister, categorizer, writer)
}
