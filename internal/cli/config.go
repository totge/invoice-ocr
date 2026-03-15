package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/totge/invoice-oc/go_item_categorizer/internal/config"
)

type ConfigInitOptions struct {
	Path  string
	Force bool
}

type ConfigGetOptions struct {
	Key string
}

type ConfigSetOptions struct {
	Key string
}

// NewConfigCmd creates the "config" command with init, set, get, list subcommands.
func NewConfigCmd(v *viper.Viper) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage configuration",
		Long:  "View and modify the TOML configuration file.",
	}

	cmd.AddCommand(newConfigInitCmd())
	cmd.AddCommand(newConfigSetCmd(v))
	cmd.AddCommand(newConfigGetCmd(v))
	cmd.AddCommand(newConfigListCmd(v))

	return cmd
}

// TODO: Is it ok to provide the default config path like that
func newConfigInitCmd() *cobra.Command {

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Create a default configuration file",
		Long:  "Creates a config.toml with all available keys set to their default values.",
		RunE:  runConfigInit,
	}

	cmd.Flags().String("path", "~/.invoice_ocr", "base directory for the config file (default: ~/.invoice_ocr)")
	cmd.Flags().Bool("force", false, "overwrite existing config file")

	return cmd
}

func newConfigGetCmd(v *viper.Viper) *cobra.Command {
	return &cobra.Command{
		Use:   "get <key>",
		Short: "Get a configuration value",
		Long:  "Read a value from the TOML config file. Keys use dotted paths (e.g., ai.gemini_api_key).",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {

			key := args[0]

			// 1. Create a "Clean" Viper instance, to only read values from the config file
			fileViper := viper.New()
			fileViper.SetConfigFile(v.ConfigFileUsed()) // Use the known file path

			// 2. ONLY read from the file.
			// Do NOT call AutomaticEnv() or BindPFlags() on this instance!
			if err := fileViper.ReadInConfig(); err != nil {
				// If the file doesn't exist yet, that's okay, we'll create it on Write
				if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
					return fmt.Errorf("failed to read config file: %w", err)
				}
			}

			value := fileViper.Get(key)

			if value == nil {
				keys := fileViper.AllKeys()
				fmt.Printf("%v", keys)
			}

			fmt.Printf("%s = %s\n", key, value)
			return nil
		},
	}
}

func newConfigSetCmd(v *viper.Viper) *cobra.Command {
	return &cobra.Command{
		Use:   "set <key> <value>",
		Short: "Set a configuration value",
		Long:  "Set a value in the TOML config file. Keys use dotted paths (e.g., ai.gemini_api_key).",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			key, value := args[0], args[1]

			//1. Validate key
			if !config.IsValidKey(key) {
				return fmt.Errorf("Invalid key '%s'", key)
			}

			// 2. Create a "Clean" Viper instance, to prevent modification of other options
			fileViper := viper.New()
			fileViper.SetConfigFile(v.ConfigFileUsed()) // Use the known file path

			// 2. ONLY read from the file.
			// Do NOT call AutomaticEnv() or BindPFlags() on this instance!
			if err := fileViper.ReadInConfig(); err != nil {
				// If the file doesn't exist yet, that's okay, we'll create it on Write
				if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
					return fmt.Errorf("failed to read config file: %w", err)
				}
			}

			// 3. Set the value in this clean, file-only state
			fileViper.Set(key, value)

			// 4. Write the clean state back to the file
			if err := fileViper.WriteConfig(); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}

			fmt.Printf("Set %s = %s\n", key, value)
			return nil

		},
	}
}

func newConfigListCmd(v *viper.Viper) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all configuration values and their values in the config file",
		RunE: func(cmd *cobra.Command, args []string) error {

			// 1. Create a "Clean" Viper instance, to only read values from the config file
			fileViper := viper.New()
			fileViper.SetConfigFile(v.ConfigFileUsed()) // Use the known file path

			if err := fileViper.ReadInConfig(); err != nil {
				// If the file doesn't exist yet, that's okay, we'll create it on Write
				if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
					return fmt.Errorf("failed to read config file: %w", err)
				}
			}

			w := cmd.OutOrStdout()

			return fileViper.WriteConfigTo(w)
		},
	}
}

func runConfigInit(cmd *cobra.Command, args []string) error {
	// TODO: does this need error handling
	path, _ := cmd.Flags().GetString("path")
	force, _ := cmd.Flags().GetBool("force")

	if path == "" {
		return fmt.Errorf("--path cannot be empty string")
	}

	// check if config file already exists
	if !force {
		if _, err := os.Stat(path); err == nil {
			return fmt.Errorf("config file already exists: %s (use --force to overwrite)", path)
		}
	}

	config.InitConfigFile(path, force)

	return nil
}

// func runConfigSet(cmd *cobra.Command, v *viper.Viper) error {

// }

// func resolveConfigPath(cfgFile *string) string {
// 	if cfgFile != nil && *cfgFile != "" {
// 		return *cfgFile
// 	}
// 	path, err := config.DefaultConfigPath()
// 	if err != nil {
// 		home, _ := os.UserHomeDir()
// 		return config.ConfigPathIn(home)
// 	}
// 	return path
// }

// func detectSource(viperKey, tomlPath string, fileCfg *config.Config) string {
// 	envKey := "INVOICE_CATEGORIZER_" + strings.ToUpper(strings.ReplaceAll(viperKey, "-", "_"))
// 	if _, exists := os.LookupEnv(envKey); exists {
// 		return "env"
// 	}

// 	if fileCfg != nil {
// 		fileValue, err := config.GetValue(fileCfg, tomlPath)
// 		if err == nil && fileValue != "" {
// 			return "file"
// 		}
// 	}

// 	return "default"
// }

// TODO: Add masking later

// func isSensitive(tomlPath string) bool {
// 	return strings.Contains(tomlPath, "api_key") || strings.Contains(tomlPath, "key_path")
// }

// func maskValue(s string) string {
// 	if len(s) <= 8 {
// 		return "****"
// 	}
// 	return s[:4] + "..." + s[len(s)-4:]
// }
