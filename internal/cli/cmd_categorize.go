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
	"github.com/totge/invoice-oc/go_item_categorizer/internal/jsonfilereader"
	"github.com/totge/invoice-oc/go_item_categorizer/internal/jsonfilewriter"
	"github.com/totge/invoice-oc/go_item_categorizer/internal/llmcategorizer"
)

type categorizeCommand struct {
	baseConfig
	Source       string
	InputPath    string
	OutputTarget string
	OutputPath   string
}

// Ensure categorizeCommand satisfies the command interface.
var _ command = (*categorizeCommand)(nil)

func (c *categorizeCommand) GetName() string {
	return "categorize"
}

func (c *categorizeCommand) SetDefaults() {
	c.Source = "local"
	c.InputPath = ""
	c.OutputTarget = "file"
	c.OutputPath = "output.json"
}

func (c *categorizeCommand) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(&c.Source, "source", c.Source, "Source system")
	fs.StringVar(&c.InputPath, "input", c.InputPath, "Path to the input receipt file (JSON)")
	fs.StringVar(&c.OutputTarget, "target", c.OutputTarget, "Output target: 'file' or 'appsheet'")
	fs.StringVar(&c.OutputPath, "output", c.OutputPath, "Output file path (only used if target is 'file')")
}

func (c *categorizeCommand) ValidateOptions() error {

	if c.InputPath == "" {
		return fmt.Errorf("--input is required")
	}

	switch c.Source {
	case "local":
		if info, err := os.Stat(c.InputPath); os.IsNotExist(err) {
			return fmt.Errorf("input file does not exist: %s", c.InputPath)
		} else if info.IsDir() {
			return fmt.Errorf("input path is a directory: %s", c.InputPath)
		} else if err != nil {
			return fmt.Errorf("failed to check input file %s: %w", c.InputPath, err)
		}
	case "gdrive":
		// No preemptive validation
	default:
		return fmt.Errorf("invalid source '%s'. Must be 'local' or 'gdrive'", c.OutputTarget)
	}

	// validate target type
	switch c.OutputTarget {
	case "file", "appsheet": // , "terminal" to be added later
		// Valid
	default:
		return fmt.Errorf("invalid target '%s'. Must be 'file', 'appsheet', or 'terminal'", c.OutputTarget)
	}

	// validate output path (for file target)
	if c.OutputTarget == "file" {
		if c.OutputPath == "" {
			return fmt.Errorf("--output is required when target is 'file'")
		}
		if c.InputPath == c.OutputPath {
			return fmt.Errorf("--input and --output paths cannot be the same")
		}
	}

	return nil
}

// RunCategorizeCommand handles the 'categorize' CLI command.
// args: os.Args[2:] (arguments after 'categorize')
func RunCategorizeCommand(args []string) error {
	// 1. Initialize and load command options
	var options categorizeCommand
	err := parseCommandOptions(&options, args)
	if err != nil {
		return fmt.Errorf("failed running 'extract' command: %w", err)
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
	var reader app.ReceiptReader
	switch options.Source {
	case "local":
		reader = jsonfilereader.NewReader(options.InputPath)
	case "gdrive":
		reader, err = googledrive.New(ctx, options.GetConfig().GDriveKeyPath, options.InputPath)
		if err != nil {
			return fmt.Errorf("failed to initialize Google Drive reader: %w", err)
		}
	}

	// initialize catalog
	appsheetClient, err := appsheet.NewClient(options.GetConfig().AppSheetBaseUrl, options.GetConfig().AppSheetAppId, options.GetConfig().AppSheetApiKey)
	if err != nil {
		return fmt.Errorf("failed to initialize Appsheet client: %w", err)
	}

	lister := appsheetcatalog.New(appsheetClient)

	// initialize categorizer
	llmClient, err := geminiclient.New(ctx, options.GetConfig().GeminiApiKey)
	if err != nil {
		return fmt.Errorf("failed to initialize Gemini client: %w", err)
	}
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
	// fmt.Printf("Processing %s -> %s...\n", *inputPath, *outputTarget)
	if err := app.Categorize(ctx, reader, lister, categorizer, writer); err != nil {
		return err
	}

	return nil
}
