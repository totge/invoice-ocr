package cli

import (
	"errors"
	"log/slog"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/totge/invoice-oc/go_item_categorizer/internal/config"
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
	cmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file path (default is ~/.invoice_ocr/config.toml)")
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
	cmd.AddCommand(NewConfigCmd(v))

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

func initConfig(v *viper.Viper, cfgPath string) error {
	// 1. Determine Path
	if cfgPath == "" {
		var err error
		cfgPath, err = config.DefaultConfigPath()
		if err != nil {
			return err
		}
	}

	// 2. Point Viper to the config file
	v.SetConfigFile(cfgPath)
	v.SetConfigType("toml")

	// 3. Configure Environment Variables
	v.SetEnvPrefix("INVOICE_CATEGORIZER")
	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_", ".", "_"))
	v.AutomaticEnv()

	// 4. Read the file
	if err := v.ReadInConfig(); err != nil {
		// It's perfectly fine if the config file doesn't exist yet!
		// The user might be relying entirely on flags/env vars.
		if !errors.Is(err, os.ErrNotExist) {
			// Only log an error if the file exists but is corrupted/unreadable
			slog.Warn("Failed to read config file", "error", err)
		}
	} else {
		slog.Debug("Config loaded", "config_file", v.ConfigFileUsed())
	}

	slog.Debug("Config loaded", "config_file", cfgPath)

	return nil
}
