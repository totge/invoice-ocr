package main

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

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
		// PersistentPreRunE runs AFTER flags are parsed but BEFORE the subcommand runs.
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			setupLogger(logLevel, verbose)
			return initConfig(v, cfgFile)
		},
	}

	// Register Global Flags
	cmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ./.env)")
	cmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "enable debug logging")
	cmd.PersistentFlags().StringVar(&logLevel, "log-level", "", "set log level")
	cmd.PersistentFlags().String("gemini-api-key", "", "Gemini API Key")
	cmd.PersistentFlags().String("gemini-model", "", "Gemini model")
	cmd.PersistentFlags().String("appsheet-api-key", "", "AppSheet API Key")
	cmd.PersistentFlags().String("appsheet-app-id", "", "AppSheet App ID")
	cmd.PersistentFlags().String("appsheet-base-url", "https://www.appsheet.com", "AppSheet Base URL")
	cmd.PersistentFlags().String("google-drive-key-path", "", "path to Google Drive key file")

	// Bind flags to viper so they can be read via v.GetString("flag-name")
	v.BindPFlags(cmd.PersistentFlags())

	// Add Subcommands
	cmd.AddCommand(NewExtractCmd(v))
	cmd.AddCommand(NewProcessCmd(v))
	cmd.AddCommand(NewCategorizeCmd(v))
	cmd.AddCommand(NewListCmd(v))

	return cmd
}

// setupLogger configures slog with the given level.
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

// initConfig loads configuration from .env file and environment variables.
func initConfig(v *viper.Viper, cfgFile string) error {
	if cfgFile != "" {
		v.SetConfigFile(cfgFile)
	} else {
		v.AddConfigPath(".")
		v.SetConfigName(".env")
		v.SetConfigType("env")
	}

	v.SetEnvPrefix("INVOICE_CATEGORIZER")
	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return fmt.Errorf("error reading config file: %w", err)
		}
	}

	slog.Debug("Config loaded", "file", v.ConfigFileUsed())

	return nil
}
