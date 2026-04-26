package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/totge/invoice-oc/go_item_categorizer/internal/config"
)

// NewConfigCmd creates the "config" command with init, set, get, list subcommands.
func NewConfigCmd(cfgFile *string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage configuration",
		Long:  "View and modify the TOML configuration file.",
	}

	cmd.AddCommand(newConfigInitCmd())
	cmd.AddCommand(newConfigSetCmd(cfgFile))
	cmd.AddCommand(newConfigGetCmd(cfgFile))
	cmd.AddCommand(newConfigListCmd(cfgFile))

	return cmd
}

func newConfigInitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Create a default configuration file",
		Long:  "Creates a config.toml with all available keys set to their default values.",
		RunE:  runConfigInit,
	}

	cmd.Flags().String("path", "", "config file path (default: ~/.invoice_ocr/config.toml)")
	cmd.Flags().Bool("force", false, "overwrite existing config file")

	return cmd
}

func newConfigGetCmd(cfgFile *string) *cobra.Command {
	return &cobra.Command{
		Use:   "get <key>",
		Short: "Get a configuration value",
		Long:  "Read a value from the TOML config file. Keys use dotted paths (e.g., ai.gemini_api_key).",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runConfigGet(cmd, args, cfgFile)
		},
	}
}

func newConfigSetCmd(cfgFile *string) *cobra.Command {
	return &cobra.Command{
		Use:   "set <key> <value>",
		Short: "Set a configuration value",
		Long:  "Set a value in the TOML config file. Keys use dotted paths (e.g., ai.gemini_api_key).",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runConfigSet(cmd, args, cfgFile)
		},
	}
}

func newConfigListCmd(cfgFile *string) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all configuration values from the config file",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runConfigList(cmd, cfgFile)
		},
	}
}

func runConfigInit(cmd *cobra.Command, args []string) error {
	filePath, _ := cmd.Flags().GetString("path")
	force, _ := cmd.Flags().GetBool("force")

	if filePath == "" {
		var err error
		filePath, err = defaultConfigPath()
		if err != nil {
			return err
		}
	} else {
		var err error
		filePath, err = expandHome(filePath)
		if err != nil {
			return err
		}
	}

	if !force {
		if _, err := os.Stat(filePath); err == nil {
			return fmt.Errorf("config file already exists: %s (use --force to overwrite)", filePath)
		}
	}

	if err := config.InitConfigFile(filePath); err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Config file created at %s\n", filePath)
	return nil
}

func runConfigGet(cmd *cobra.Command, args []string, cfgFile *string) error {
	path, err := resolveConfigPath(cfgFile)
	if err != nil {
		return err
	}

	value, err := config.GetValue(path, args[0])
	if err != nil {
		return err
	}

	fmt.Fprintf(cmd.OutOrStdout(), "%s = %s\n", args[0], value)
	return nil
}

func runConfigSet(cmd *cobra.Command, args []string, cfgFile *string) error {
	path, err := resolveConfigPath(cfgFile)
	if err != nil {
		return err
	}

	if err := config.SetValue(path, args[0], args[1]); err != nil {
		return err
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Set %s = %s\n", args[0], args[1])
	return nil
}

func runConfigList(cmd *cobra.Command, cfgFile *string) error {
	path, err := resolveConfigPath(cfgFile)
	if err != nil {
		return err
	}

	entries, err := config.ListValues(path)
	if err != nil {
		return err
	}

	w := cmd.OutOrStdout()
	for _, kv := range entries {
		fmt.Fprintf(w, "%-25s = %s\n", kv.Key, kv.Value)
	}
	return nil
}

// defaultConfigPath returns the default config file path: ~/.invoice_ocr/config.toml.
func defaultConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine home directory: %w", err)
	}
	return filepath.Join(home, ".invoice_ocr", "config.toml"), nil
}

// resolveConfigPath returns the config file path from the flag or the default.
func resolveConfigPath(cfgFile *string) (string, error) {
	if cfgFile != nil && *cfgFile != "" {
		return *cfgFile, nil
	}
	return defaultConfigPath()
}

// expandHome replaces a leading ~ with the user's home directory.
func expandHome(path string) (string, error) {
	if path == "~" || strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("cannot determine home directory: %w", err)
		}
		return filepath.Join(home, path[1:]), nil
	}
	return path, nil
}
