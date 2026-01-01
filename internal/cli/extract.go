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

type extractOptions struct {
	InputPath  string
	OutputPath string
	*Config
}

func (p *extractOptions) registerFlags(fs *flag.FlagSet) {
	fs.StringVar(&p.InputPath, "input", p.InputPath, "Path to the input receipt image")
	fs.StringVar(&p.OutputPath, "output", p.OutputPath, "Output file path")
	p.registerConfigFlags(fs)
}

func (p *extractOptions) validate() error {

	// validate app config
	err := p.Config.validate()
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

	return nil
}

// Sets the default arguments for the command
func newExtractDefaultOptions() extractOptions {
	cfg := newConfig()
	return extractOptions{
		InputPath:  "",
		OutputPath: "",
		Config:     cfg,
	}
}

func getExtractCommandOptions(args []string) (extractOptions, error) {
	options := newExtractDefaultOptions()
	fs := flag.NewFlagSet("extract", flag.ExitOnError)
	options.registerFlags(fs)
	err := fs.Parse(args)

	if err != nil {
		return options, fmt.Errorf("failed parsing cli arguments for 'extract' command: %w", err)
	}

	err = options.validate()
	if err != nil {
		return options, fmt.Errorf("invalid arguments for 'process' command: %w", err)
	}

	return options, nil
}

// RunProcessCommand handles the 'process' CLI command.
// args: os.Args[2:] (arguments after 'process')
func RunExtractCommand(args []string) error {
	// 1. Initialize Config (Loads Defaults + Env)
	options, err := getExtractCommandOptions(args)
	if err != nil {
		return fmt.Errorf("failed running 'extract' command: %w", err)
	}

	// 4. Setup Context (Cancellation)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	// 5. Construct Dependencies

	// initialize input reader
	reader := imagereader.NewReader(options.InputPath)

	// initialize extractor
	llmClient, err := geminiclient.New(ctx, options.GeminiApiKey)
	if err != nil {
		return fmt.Errorf("failed to initialize Gemini client: %w", err)
	}
	extractor := ocrextractor.New(llmClient, options.GeminiModel)

	// initialize result writer
	writer, err := jsonfilewriter.New(options.OutputPath)
	if err != nil {
		return fmt.Errorf("failed to initialize json file writer: %w", err)
	}
	// // 6. Run App Logic
	// fmt.Printf("Processing %s -> %s...\n", *inputPath, *outputTarget)
	if err := app.Extract(ctx, reader, extractor, writer); err != nil {
		return err
	}

	return nil
}
