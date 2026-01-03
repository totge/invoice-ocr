package cli

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/totge/invoice-oc/go_item_categorizer/internal/app"
	"github.com/totge/invoice-oc/go_item_categorizer/internal/appsheet"
	"github.com/totge/invoice-oc/go_item_categorizer/internal/appsheetcatalog"
	"github.com/totge/invoice-oc/go_item_categorizer/internal/appsheetwriter"
	"github.com/totge/invoice-oc/go_item_categorizer/internal/geminiclient"
	"github.com/totge/invoice-oc/go_item_categorizer/internal/jsonfilereader"
	"github.com/totge/invoice-oc/go_item_categorizer/internal/jsonfilewriter"
	"github.com/totge/invoice-oc/go_item_categorizer/internal/llmcategorizer"
)

type processOptions struct {
	baseConfig
	InputPath    string
	OutputTarget string
	OutputPath   string
}

// Ensure extractCommand satisfies the command interface.
var _ command = (*processOptions)(nil)

func (p *processOptions) GetName() string {
	return "process"
}

func (p *processOptions) SetDefaults() {
	p.InputPath = ""
	p.OutputTarget = "file"
	p.OutputPath = "output.json"
}

func (p *processOptions) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(&p.InputPath, "input", p.InputPath, "Path to the input receipt file (JSON)")
	fs.StringVar(&p.OutputTarget, "target", p.OutputTarget, "Output target: 'file' or 'appsheet'")
	fs.StringVar(&p.OutputPath, "output", p.OutputPath, "Output file path (only used if target is 'file')")
}

func (p *processOptions) ValidateOptions() error {

	if p.InputPath == "" {
		return fmt.Errorf("--input is required")
	}

	if info, err := os.Stat(p.InputPath); os.IsNotExist(err) {
		return fmt.Errorf("input file does not exist: %s", p.InputPath)
	} else if info.IsDir() {
		return fmt.Errorf("input path is a directory: %s", p.InputPath)
	} else if err != nil {
		return fmt.Errorf("failed to check input file %s: %w", p.InputPath, err)
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

// RunProcessCommand handles the 'process' CLI command.
// args: os.Args[2:] (arguments after 'process')
func RunProcessCommand(args []string) error {
	// 1. Initialize and load command options
	var options processOptions
	err := parseCommandOptions(&options, args)
	if err != nil {
		return fmt.Errorf("failed running 'extract' command: %w", err)
	}

	// 2. Setup Context (Cancellation)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	// 3. Construct Dependencies

	// initialize input reader
	reader := jsonfilereader.NewReader(options.InputPath)

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
	if err := app.Process(ctx, reader, lister, categorizer, writer); err != nil {
		return err
	}

	return nil
}
