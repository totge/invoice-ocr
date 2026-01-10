package cli

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/totge/invoice-oc/go_item_categorizer/internal/app"
	"github.com/totge/invoice-oc/go_item_categorizer/internal/localfilelister"
	"github.com/totge/invoice-oc/go_item_categorizer/internal/sourcefilter"
	"github.com/totge/invoice-oc/go_item_categorizer/internal/terminalwriter"
)

type listCommand struct {
	baseConfig
	SourceType string
	Path       string
	Limit      int
	Target     string
}

// Ensure listCommand satisfies the command interface.
var _ command = (*listCommand)(nil)

func (l *listCommand) GetName() string {
	return "extract"
}

func (l *listCommand) SetDefaults() {
	l.SourceType = "local"
	l.Path = ""
	l.Limit = 10
	l.Target = "terminal"
}

func (l *listCommand) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(&l.SourceType, "source", l.SourceType, "Source system")
	fs.StringVar(&l.Path, "path", l.Path, "Path to source files")
	fs.IntVar(&l.Limit, "output", l.Limit, "Max. number of files in putput")
	fs.StringVar(&l.Target, "target", l.Target, "Target for command output")
}

func (l *listCommand) ValidateOptions() error {

	if l.Path == "" {
		return fmt.Errorf("--path is required")
	}

	switch l.SourceType {
	case "local":
		if info, err := os.Stat(l.Path); os.IsNotExist(err) {
			return fmt.Errorf("input directory does not exist: %s", l.Path)
		} else if !info.IsDir() {
			return fmt.Errorf("input path is not a directory: %s", l.Path)
		} else if err != nil {
			return fmt.Errorf("failed to check input path %s: %w", l.Path, err)
		}
	default:
		// pass validation for external systems
	}

	switch l.Target {
	case "terminal":
		// valid
	default:
		return fmt.Errorf("invalid --target: %s\nmust be one of [terminal]", l.Target)
	}

	return nil
}

// RunProcessCommand handles the 'process' CLI command.
// args: os.Args[2:] (arguments after 'process')
func RunlistCommand(args []string) error {
	// 1. Initialize and load command options
	var options listCommand
	err := parseCommandOptions(&options, args)
	if err != nil {
		return fmt.Errorf("failed running 'list' command: %w", err)
	}

	slog.Debug("List command configuration",
		"path", options.Path,
		"source", options.SourceType,
		"target", options.Target,
	)

	// 2. Setup Context (Cancellation)
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	// 3. Construct Dependencies

	// initialize local file lister
	lister := localfilelister.New(options.Path)

	// initialize filter for files
	filter := sourcefilter.New(options.Limit)

	// initialize writer
	var writer app.SourceListWriter
	switch options.Target {
	case "terminal":
		writer = terminalwriter.New()
	}

	// // 4. Run App Logic
	if err := app.ListInputs(ctx, lister, filter, writer); err != nil {
		return err
	}

	return nil
}
