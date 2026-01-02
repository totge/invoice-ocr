package cli

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/totge/invoice-oc/go_item_categorizer/internal/app"
	"github.com/totge/invoice-oc/go_item_categorizer/internal/geminiclient"
	"github.com/totge/invoice-oc/go_item_categorizer/internal/imagereader"
	"github.com/totge/invoice-oc/go_item_categorizer/internal/jsonfilewriter"
	"github.com/totge/invoice-oc/go_item_categorizer/internal/ocrextractor"
)

type extractCommand struct {
	baseConfig
	InputPath  string
	OutputPath string
}

// Ensure extractCommand satisfies the command interface.
var _ command = (*extractCommand)(nil)

func (e *extractCommand) GetName() string {
	return "extract"
}

func (e *extractCommand) SetDefaults() {
	e.InputPath = ""
	e.OutputPath = ""
}

func (e *extractCommand) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(&e.InputPath, "input", e.InputPath, "Path to the input receipt image")
	fs.StringVar(&e.OutputPath, "output", e.OutputPath, "Output file path")
}

func (e *extractCommand) ValidateOptions() error {

	if e.InputPath == "" {
		return fmt.Errorf("--input is required")
	}

	if info, err := os.Stat(e.InputPath); os.IsNotExist(err) {
		return fmt.Errorf("input file does not exist: %s", e.InputPath)
	} else if info.IsDir() {
		return fmt.Errorf("input path is a directory: %s", e.InputPath)
	} else if err != nil {
		return fmt.Errorf("failed to check input file %s: %w", e.InputPath, err)
	}

	if e.OutputPath == "" {
		return fmt.Errorf("--output is required")
	}

	return nil
}

// RunProcessCommand handles the 'process' CLI command.
// args: os.Args[2:] (arguments after 'process')
func RunExtractCommand(args []string) error {
	// 1. Initialize and load command options
	var options extractCommand
	err := parseCommandOptions(&options, args)
	if err != nil {
		return fmt.Errorf("failed running 'extract' command: %w", err)
	}

	// 2. Setup Context (Cancellation)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	// 3. Construct Dependencies

	// initialize input reader
	reader := imagereader.NewReader(options.InputPath)

	// initialize extractor
	llmClient, err := geminiclient.New(ctx, options.GetConfig().GeminiApiKey)
	if err != nil {
		return fmt.Errorf("failed to initialize Gemini client: %w", err)
	}
	extractor := ocrextractor.New(llmClient, options.GetConfig().GeminiModel)

	// initialize result writer
	writer, err := jsonfilewriter.New(options.OutputPath)
	if err != nil {
		return fmt.Errorf("failed to initialize json file writer: %w", err)
	}
	// // 4. Run App Logic
	// fmt.Printf("Processing %s -> %s...\n", *inputPath, *outputTarget)
	if err := app.Extract(ctx, reader, extractor, writer); err != nil {
		return err
	}

	return nil
}
