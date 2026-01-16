package cli

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"time"

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

type processCommand struct {
	baseConfig
	Source       string
	InputPath    string
	OutputTarget string
	OutputPath   string
}

// Ensure processCommand satisfies the command interface.
var _ command = (*processCommand)(nil)

func (p *processCommand) GetName() string {
	return "process"
}

func (p *processCommand) SetDefaults() {
	p.Source = "local"
	p.InputPath = ""
	p.OutputTarget = "file"
	p.OutputPath = "output.json"
}

func (p *processCommand) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(&p.Source, "source", p.Source, "Source system: 'local' or 'gdrive'")
	fs.StringVar(&p.InputPath, "input", p.InputPath, "Path to the input receipt image")
	fs.StringVar(&p.OutputTarget, "target", p.OutputTarget, "Output target: 'file' or 'appsheet'")
	fs.StringVar(&p.OutputPath, "output", p.OutputPath, "Output file path (only used if target is 'file')")
}

func (p *processCommand) ValidateOptions() error {

	if p.InputPath == "" {
		return fmt.Errorf("--input is required")
	}

	// validate source type
	switch p.Source {
	case "local": // , "terminal" to be added later
		if info, err := os.Stat(p.InputPath); os.IsNotExist(err) {
			return fmt.Errorf("input file does not exist: %s", p.InputPath)
		} else if info.IsDir() {
			return fmt.Errorf("input path is a directory: %s", p.InputPath)
		} else if err != nil {
			return fmt.Errorf("failed to check input file %s: %w", p.InputPath, err)
		}
	case "gdrive":
		// No preemptive validation
	default:
		return fmt.Errorf("invalid source '%s'. Must be 'local' or 'gdrive'", p.OutputTarget)
	}

	// validate target type
	switch p.OutputTarget {
	case "file", "appsheet": // , "terminal" to be added later
		// Valid
	default:
		return fmt.Errorf("invalid target '%s'. Must be 'file', 'appsheet', or 'terminal'", p.OutputTarget)
	}

	// validate output path (for file target)
	if p.OutputTarget == "file" {
		if p.OutputPath == "" {
			return fmt.Errorf("--output is required when target is 'file'")
		}
		if p.InputPath == p.OutputPath {
			return fmt.Errorf("--input and --output paths cannot be the same")
		}
	}

	return nil
}

func RunProcessCommand(args []string) error {
	// 1. Initialize and load command options
	var options processCommand
	err := parseCommandOptions(&options, args)
	if err != nil {
		return fmt.Errorf("failed running 'process' command: %w", err)
	}

	slog.Debug("Categorize command configuration",
		"input", options.InputPath,
		"target", options.OutputTarget,
		"output", options.OutputPath,
	)

	// 2. Setup Context (Cancellation)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	// 3. Construct Dependencies

	// initialize input reader
	var reader app.ReceiptImageReader

	switch options.Source {
	case "local":
		reader = imagereader.NewReader(options.InputPath)
	case "gdrive":
		reader, err = googledrive.New(ctx, options.GetConfig().GDriveKeyPath, options.InputPath)
		if err != nil {
			return fmt.Errorf("failed to initialize Google Drive reader: %w", err)
		}
	}

	// initialize llm client
	llmClient, err := geminiclient.New(ctx, options.GetConfig().GeminiApiKey)
	if err != nil {
		return fmt.Errorf("failed to initialize Gemini client: %w", err)
	}

	// initialize extractor
	extractor := ocrextractor.New(llmClient, options.GetConfig().GeminiModel)

	// initialize catalog
	appsheetClient, err := appsheet.NewClient(options.GetConfig().AppSheetBaseUrl, options.GetConfig().AppSheetAppId, options.GetConfig().AppSheetApiKey)
	if err != nil {
		return fmt.Errorf("failed to initialize Appsheet client: %w", err)
	}

	lister := appsheetcatalog.New(appsheetClient)

	// initialize categorizer
	categorizer := llmcategorizer.New(llmClient, options.GetConfig().GeminiModel)

	// initialize result writer
	var writer app.ResultWriter
	switch options.OutputTarget {
	case "file":
		w, err := jsonfilewriter.New(options.OutputPath)
		if err != nil {
			return err
		}

		writer = w
	case "appsheet":
		writer = appsheetwriter.New(appsheetClient)

	default:
		return fmt.Errorf("unknown target: %s", options.OutputTarget)
	}

	// 4. Run App Logic
	if err := app.Process(ctx, reader, extractor, lister, categorizer, writer); err != nil {
		return err
	}

	return nil
}
