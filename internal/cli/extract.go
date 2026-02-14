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
	"github.com/totge/invoice-oc/go_item_categorizer/internal/geminiclient"
	"github.com/totge/invoice-oc/go_item_categorizer/internal/googledrive"
	"github.com/totge/invoice-oc/go_item_categorizer/internal/imagereader"
	"github.com/totge/invoice-oc/go_item_categorizer/internal/jsonfilewriter"
	"github.com/totge/invoice-oc/go_item_categorizer/internal/ocrextractor"
)

// ExtractOptions holds all inputs for the extract command.
type ExtractOptions struct {
	Source        string
	Input         string
	Output        string
	GeminiAPIKey  string
	GeminiModel   string
	GDriveKeyPath string
}

// Validate checks that all options are valid. It applies defaults where needed.
func (o *ExtractOptions) Validate() error {
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

	if o.GeminiAPIKey == "" {
		return fmt.Errorf("Gemini API key is required (set INVOICE_CATEGORIZER_GEMINI_API_KEY or --gemini-api-key)")
	}
	if o.GeminiModel == "" {
		o.GeminiModel = "gemini-2.0-flash"
	}

	return nil
}

func newExtractOptions(cmd *cobra.Command, v *viper.Viper) ExtractOptions {
	source, _ := cmd.Flags().GetString("source")
	input, _ := cmd.Flags().GetString("input")
	output, _ := cmd.Flags().GetString("output")

	return ExtractOptions{
		Source:        source,
		Input:         input,
		Output:        output,
		GeminiAPIKey:  v.GetString("gemini-api-key"),
		GeminiModel:   v.GetString("gemini-model"),
		GDriveKeyPath: v.GetString("google-drive-key-path"),
	}
}

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

	cmd.Flags().StringP("source", "s", "local", "input source: 'local' or 'gdrive'")
	cmd.Flags().StringP("input", "i", "", "path to the receipt image")
	cmd.Flags().StringP("output", "o", "", "output file path")

	cmd.MarkFlagRequired("input")
	cmd.MarkFlagRequired("output")

	return cmd
}

func runExtract(cmd *cobra.Command, v *viper.Viper) error {
	opts := newExtractOptions(cmd, v)
	if err := opts.Validate(); err != nil {
		return err
	}

	slog.Debug("Extract command configuration",
		"source", opts.Source,
		"input", opts.Input,
		"output", opts.Output,
		"model", opts.GeminiModel,
	)

	ctx, cancel := context.WithTimeout(cmd.Context(), 3*time.Minute)
	defer cancel()

	// Construct dependencies
	var reader app.ReceiptImageReader
	switch opts.Source {
	case "local":
		reader = imagereader.NewReader(opts.Input)
	case "gdrive":
		var err error
		reader, err = googledrive.New(ctx, opts.GDriveKeyPath, opts.Input)
		if err != nil {
			return fmt.Errorf("failed to initialize Google Drive client: %w", err)
		}
	}

	llmClient, err := geminiclient.New(ctx, opts.GeminiAPIKey)
	if err != nil {
		return fmt.Errorf("failed to initialize Gemini client: %w", err)
	}
	extractor := ocrextractor.New(llmClient, opts.GeminiModel)

	writer, err := jsonfilewriter.New(opts.Output)
	if err != nil {
		return fmt.Errorf("failed to initialize json file writer: %w", err)
	}

	return app.Extract(ctx, reader, extractor, writer)
}
