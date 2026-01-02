package cli

import (
	"flag"
	"fmt"
	"os"
)

type command interface {
	GetName() string
	SetDefaults() 
	RegisterFlags(*flag.FlagSet) // Binds command-specific flags
	ValidateOptions() error      // Validates command-specific options (not global config)
	EmbedConfig(*config)         // Method to inject the parsed global config
	GetConfig() *config          // Method to retrieve the config for validation
}

func parseCommandOptions(cmd command, args []string) error {
	// 1. Set default values
	cmd.SetDefaults()

	// 2. Initialize and Load Global Config (defaults + env vars)
	cfg := newConfig()
	cmd.EmbedConfig(cfg) // Inject the loaded config into the command's options

	// 3. Create FlagSet
	fs := flag.NewFlagSet(cmd.GetName(), flag.ContinueOnError)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s %s:\n", os.Args[0], cmd.GetName())
		fs.PrintDefaults()
	}

	// 4. Register Global Config Flags
	cfg.registerConfigFlags(fs)

	// 5. Register Command-Specific Flags
	cmd.RegisterFlags(fs)

	// 6. Parse Flags
	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("failed to parse flags for command '%s': %w", cmd.GetName(), err)
	}

	// 7. Validate Global Config
	if err := cmd.GetConfig().ValidateConfig(); err != nil {
		return fmt.Errorf("invalid global configuration for command '%s': %w", cmd.GetName(), err)
	}

	// 8. Validate Command-Specific Options
	if err := cmd.ValidateOptions(); err != nil {
		return fmt.Errorf("invalid options for command '%s': %w", cmd.GetName(), err)
	}

	return nil
}
