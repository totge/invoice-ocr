package cli

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// NewRootCmd creates the entry point of the CLI application.
func NewRootCmd() *cobra.Command {
	v := viper.New()
	var (
		cfgFile  string
		verbose  bool
		logLevel string
	)

	cmd := &cobra.Command{
		Use:   "categorizer",
		Short: "A tool to categorize receipts using AI",
		Long: `Categorizer is a CLI tool that uses an LLM to extract data from
receipts (OCR) and categorize products into different categories.`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			setupLogger(logLevel, verbose)
			return initConfig(v, cfgFile)
		},
	}

	// Register Global Flags
	cmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ./.env)")
	cmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "enable debug logging")
	cmd.PersistentFlags().StringVar(&logLevel, "log-level", "", "set log level (debug, info, error)")
	cmd.PersistentFlags().String("gemini-api-key", "", "Gemini API Key")
	cmd.PersistentFlags().String("gemini-model", "", "Gemini model")
	cmd.PersistentFlags().String("appsheet-api-key", "", "AppSheet API Key")
	cmd.PersistentFlags().String("appsheet-app-id", "", "AppSheet App ID")
	cmd.PersistentFlags().String("appsheet-base-url", "https://www.appsheet.com", "AppSheet Base URL")
	cmd.PersistentFlags().String("google-drive-key-path", "", "path to Google Drive key file")

	v.BindPFlags(cmd.PersistentFlags())

	// Add Subcommands
	cmd.AddCommand(NewExtractCmd(v))
	cmd.AddCommand(NewProcessCmd(v))
	cmd.AddCommand(NewCategorizeCmd(v))
	cmd.AddCommand(NewListCmd(v))

	return cmd
}

func setupLogger(logLevel string, isVerbose bool) {
	level := slog.LevelError
	switch logLevel {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "error":
		level = slog.LevelError
	}

	if isVerbose {
		level = slog.LevelDebug
	}

	opts := &slog.HandlerOptions{
		Level:     level,
		AddSource: true,
	}
	logger := slog.New(slog.NewTextHandler(os.Stderr, opts))
	slog.SetDefault(logger)
}

func initConfig(v *viper.Viper, cfgFile string) error {
	// Load .env file into OS environment. godotenv does not overwrite
	// existing env vars, so real environment always takes precedence.
	if cfgFile != "" {
		if err := godotenv.Load(cfgFile); err != nil {
			return fmt.Errorf("error loading config file %s: %w", cfgFile, err)
		}
	} else {
		// Ignore error if default .env doesn't exist — env vars may suffice.
		_ = godotenv.Load()
	}

	v.SetEnvPrefix("INVOICE_CATEGORIZER")
	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	v.AutomaticEnv()

	slog.Debug("Config loaded")

	return nil
}
