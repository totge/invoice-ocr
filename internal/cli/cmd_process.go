package cli

import (
	"flag"
	"fmt"
	"os"
)

type processCommand struct {
	baseConfig
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
	p.InputPath = ""
	p.OutputTarget = "file"
	p.OutputPath = "output.json"
}

func (p *processCommand) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(&p.InputPath, "input", p.InputPath, "Path to the input receipt image")
	fs.StringVar(&p.OutputTarget, "target", p.OutputTarget, "Output target: 'file' or 'appsheet'")
	fs.StringVar(&p.OutputPath, "output", p.OutputPath, "Output file path (only used if target is 'file')")
}

func (p *processCommand) ValidateOptions() error {

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
