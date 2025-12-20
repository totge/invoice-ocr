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
	InputPath    string
	OutputTarget string
	OutputPath   string
	*Config
}

func (p *processOptions) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(&p.InputPath, "input", p.InputPath, "Path to the input receipt file (JSON)")
	fs.StringVar(&p.OutputTarget, "target", p.OutputTarget, "Output target: 'file' or 'appsheet'")
	fs.StringVar(&p.OutputPath, "output", p.OutputPath, "Output file path (only used if target is 'file')")
	p.RegisterConfigFlags(fs)
}

func (p *processOptions) Validate() error {

	// validate app config
	err := p.Config.Validate()
	if err != nil {
		// handle error
	}

	// validate input path
	if p.InputPath == "" {
		return fmt.Errorf("--input is required")
	}

	info, err := os.Stat(p.InputPath)
	if os.IsNotExist(err) {
		return fmt.Errorf("input file does not exist: %s", p.InputPath)
	}
	if info.IsDir() {
		return fmt.Errorf("input path is a directory, not a file: %s", p.InputPath)
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
			return fmt.Errorf("output path is required when target is 'file'")
		}
		if p.InputPath == p.OutputPath {
			return fmt.Errorf("input and output paths cannot be the same")
		}
	}

	return nil
}

// Sets the default arguments for the command
func NewProcessDefaultOptions() processOptions {
	cfg := NewConfig()
	return processOptions{
		InputPath:    "",
		OutputTarget: "file",
		OutputPath:   "",
		Config:       cfg,
	}
}

func GetProcessCommandOptions(args []string) (processOptions, error) {
	options := NewProcessDefaultOptions()
	fs := flag.NewFlagSet("process", flag.ExitOnError)
	options.RegisterFlags(fs)
	err := fs.Parse(args)

	if err != nil {
		return options, fmt.Errorf("failed parsing cli arguments for 'process' command: %w", err)
	}

	err = options.Validate()
	if err != nil {
		return options, fmt.Errorf("invalid arguments for 'process' command: %w", err)
	}

	return options, nil
}

// RunProcessCommand handles the 'process' CLI command.
// args: os.Args[2:] (arguments after 'process')
func RunProcessCommand(args []string) error {
	// 1. Initialize Config (Loads Defaults + Env)
	options, err := GetProcessCommandOptions(args)
	if err != nil {
		return fmt.Errorf("failed running 'process' command: %w", err)
	}

	// 4. Setup Context (Cancellation)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	// 5. Construct Dependencies

	// initialize input reader
	reader := jsonfilereader.NewReader(options.InputPath)

	// initialize catalog
	appsheetClient, err := appsheet.NewClient(options.AppSheetBaseUrl, options.AppSheetAppId, options.AppSheetApiKey)
	if err != nil {
		return fmt.Errorf("failed to initialize Appsheet client: %w", err)
	}

	lister := appsheetcatalog.New(appsheetClient)

	// initialize categorizer
	llmClient, err := geminiclient.New(ctx, options.GeminiApiKey)
	if err != nil {
		return fmt.Errorf("failed to initialize Gemini client: %w", err)
	}
	categorizer := llmcategorizer.New(llmClient, "gemini-2.0-flash")

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

	// // 6. Run App Logic
	// fmt.Printf("Processing %s -> %s...\n", *inputPath, *outputTarget)
	if err := app.Process(ctx, reader, lister, categorizer, writer); err != nil {
		return err
	}

	return nil
}
