package cli

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

// ProcessOptions holds all inputs for the process command.
type ProcessOptions struct {
	Source          string
	Input           string
	Target          string
	Output          string
	GeminiAPIKey    string
	GeminiModel     string
	GDriveKeyPath   string
	AppSheetAPIKey  string
	AppSheetAppID   string
	AppSheetBaseURL string
}

// Validate checks that all options are valid. It applies defaults where needed.
func (o *ProcessOptions) Validate() error {
	switch o.Source {
	case "local":
		info, err := os.Stat(o.Input)
		if os.IsNotExist(err) {
			return fmt.Errorf("input file does not exist: %s", o.Input)
		} else if err != nil {
			return fmt.Errorf("failed to check input file %s: %w", o.Input, err)
		} else if info.IsDir() {
			return fmt.Errorf("input path is a directory: %s", o.Input)
		}
	case "gdrive":
		if o.GDriveKeyPath == "" {
			return fmt.Errorf("path to google drive key is required.")
		}
	default:
		return fmt.Errorf("invalid --source: %s (must be 'local' or 'gdrive')", o.Source)
	}

	switch o.Target {
	case "file", "appsheet":
		// valid
	default:
		return fmt.Errorf("invalid --target: %s (must be 'file' or 'appsheet')", o.Target)
	}

	if o.Target == "file" {
		if o.Output == "" {
			return fmt.Errorf("--output is required when target is 'file'")
		}
		if o.Input == o.Output {
			return fmt.Errorf("--input and --output paths cannot be the same")
		}
	}

	if o.GeminiAPIKey == "" {
		return fmt.Errorf("Gemini API key is required (set INVOICE_CATEGORIZER_GEMINI_API_KEY or --gemini-api-key)")
	}
	if o.GeminiModel == "" {
		o.GeminiModel = "gemini-2.0-flash"
	}
	if o.AppSheetAPIKey == "" {
		return fmt.Errorf("AppSheet API key is required (set INVOICE_CATEGORIZER_APPSHEET_API_KEY or --appsheet-api-key)")
	}
	if o.AppSheetAppID == "" {
		return fmt.Errorf("AppSheet App ID is required (set INVOICE_CATEGORIZER_APPSHEET_APP_ID or --appsheet-app-id)")
	}

	return nil
}

func newProcessOptions(cmd *cobra.Command, v *viper.Viper) ProcessOptions {
	source, _ := cmd.Flags().GetString("source")
	input, _ := cmd.Flags().GetString("input")
	target, _ := cmd.Flags().GetString("target")
	output, _ := cmd.Flags().GetString("output")

	return ProcessOptions{
		Source:          source,
		Input:           input,
		Target:          target,
		Output:          output,
		GeminiAPIKey:    v.GetString("gemini-api-key"),
		GeminiModel:     v.GetString("gemini-model"),
		GDriveKeyPath:   v.GetString("google-drive-key-path"),
		AppSheetAPIKey:  v.GetString("appsheet-api-key"),
		AppSheetAppID:   v.GetString("appsheet-app-id"),
		AppSheetBaseURL: v.GetString("appsheet-base-url"),
	}
}

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
	opts := newProcessOptions(cmd, v)
	if err := opts.Validate(); err != nil {
		return err
	}

	slog.Debug("Process command configuration",
		"source", opts.Source,
		"input", opts.Input,
		"target", opts.Target,
		"output", opts.Output,
		"model", opts.GeminiModel,
	)

	ctx, cancel := context.WithTimeout(cmd.Context(), 3*time.Minute)
	defer cancel()

	// Input reader
	var reader app.ReceiptImageReader
	switch opts.Source {
	case "local":
		reader = imagereader.NewReader(opts.Input)
	case "gdrive":
		var err error
		reader, err = googledrive.New(ctx, opts.GDriveKeyPath, opts.Input)
		if err != nil {
			return fmt.Errorf("failed to initialize Google Drive reader: %w", err)
		}
	}

	// LLM client + extractor
	llmClient, err := geminiclient.New(ctx, opts.GeminiAPIKey)
	if err != nil {
		return fmt.Errorf("failed to initialize Gemini client: %w", err)
	}
	extractor := ocrextractor.New(llmClient, opts.GeminiModel)

	// Catalog
	appsheetClient, err := appsheet.NewClient(opts.AppSheetBaseURL, opts.AppSheetAppID, opts.AppSheetAPIKey)
	if err != nil {
		return fmt.Errorf("failed to initialize AppSheet client: %w", err)
	}
	lister := appsheetcatalog.New(appsheetClient)

	// Categorizer
	categorizer := llmcategorizer.New(llmClient, opts.GeminiModel)

	// Result writer
	var writer app.ResultWriter
	switch opts.Target {
	case "file":
		w, err := jsonfilewriter.New(opts.Output)
		if err != nil {
			return fmt.Errorf("failed to initialize json file writer: %w", err)
		}
		writer = w
	case "appsheet":
		writer = appsheetwriter.New(appsheetClient)
	}

	return app.Process(ctx, reader, extractor, lister, categorizer, writer)
}
