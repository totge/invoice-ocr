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
	"github.com/totge/invoice-oc/go_item_categorizer/internal/geminiclient"
	"github.com/totge/invoice-oc/go_item_categorizer/internal/googledrive"
	"github.com/totge/invoice-oc/go_item_categorizer/internal/imagereader"
	"github.com/totge/invoice-oc/go_item_categorizer/internal/jsonfilewriter"
	"github.com/totge/invoice-oc/go_item_categorizer/internal/ocrextractor"
)

// NewExtractCmd creates the "extract" subcommand for OCR extraction.
func NewExtractCmd(v *viper.Viper) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "extract",
		Short: "Extract data from a receipt image using OCR",
		Long:  `Extract uses LLM vision capabilities to read a receipt image and output structured data as JSON.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runExtract(cmd, v)
		},
	}

	cmd.Flags().String("input", "", "path to the input receipt image")
	cmd.Flags().String("output", "", "output file path")
	cmd.Flags().String("source", "local", "source system: 'local' or 'gdrive'")

	cmd.MarkFlagRequired("input")
	cmd.MarkFlagRequired("output")

	return cmd
}

func runExtract(cmd *cobra.Command, v *viper.Viper) error {
	// Read flags
	inputPath, _ := cmd.Flags().GetString("input")
	outputPath, _ := cmd.Flags().GetString("output")
	source, _ := cmd.Flags().GetString("source")

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

	// Read config from viper
	geminiApiKey := v.GetString("gemini-api-key")
	geminiModel := v.GetString("gemini-model")
	gdriveKeyPath := v.GetString("google-drive-key-path")

	if geminiApiKey == "" {
		return fmt.Errorf("Gemini API key is required (set INVOICE_CATEGORIZER_GEMINI_API_KEY or --gemini-api-key)")
	}
	if geminiModel == "" {
		geminiModel = "gemini-2.0-flash"
	}

	slog.Debug("Extract command configuration",
		"source", source,
		"input", inputPath,
		"output", outputPath,
		"model", geminiModel,
	)

	ctx, cancel := context.WithTimeout(cmd.Context(), 3*time.Minute)
	defer cancel()

	// Construct dependencies
	var reader app.ReceiptImageReader
	switch source {
	case "local":
		reader = imagereader.NewReader(inputPath)
	case "gdrive":
		var err error
		reader, err = googledrive.New(ctx, gdriveKeyPath, inputPath)
		if err != nil {
			return fmt.Errorf("failed to initialize Google Drive client: %w", err)
		}
	}

	llmClient, err := geminiclient.New(ctx, geminiApiKey)
	if err != nil {
		return fmt.Errorf("failed to initialize Gemini client: %w", err)
	}
	extractor := ocrextractor.New(llmClient, geminiModel)

	writer, err := jsonfilewriter.New(outputPath)
	if err != nil {
		return fmt.Errorf("failed to initialize json file writer: %w", err)
	}

	return app.Extract(ctx, reader, extractor, writer)
}
