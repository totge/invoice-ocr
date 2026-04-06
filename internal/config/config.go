package config

import (
	"fmt"
	"os"
	"path/filepath"

	toml "github.com/pelletier/go-toml/v2"
)

const (
	configDir  = ".invoice_ocr"
	configFile = "config.toml"
)

// Config represents the full TOML configuration file structure.
type Config struct {
	AI       AIConfig       `toml:"ai" mapstructure:"ai"`
	AppSheet AppSheetConfig `toml:"appsheet" mapstructure:"appsheet"`
	GDrive   GDriveConfig   `toml:"gdrive" mapstructure:"gdrive"`
	Catalog  CatalogConfig  `toml:"catalog" mapstructure:"catalog"`
	Defaults DefaultsConfig `toml:"defaults" mapstructure:"defaults"`
}

type AIConfig struct {
	GeminiAPIKey string `toml:"gemini_api_key" mapstructure:"gemini_api_key"`
	GeminiModel  string `toml:"gemini_model" mapstructure:"gemini_model"`
}

type AppSheetConfig struct {
	APIKey  string `toml:"api_key" mapstructure:"api_key"`
	AppID   string `toml:"app_id" mapstructure:"app_id"`
	BaseURL string `toml:"base_url" mapstructure:"base_url"`
}

type GDriveConfig struct {
	KeyPath string `toml:"key_path" mapstructure:"key_path"`
}

type CatalogConfig struct {
	Type    string `toml:"type" mapstructure:"type"`
	CSVPath string `toml:"csv_path" mapstructure:"csv_path"`
}

type DefaultsConfig struct {
	LogLevel string `toml:"log_level" mapstructure:"log_level"`
}

// validKeys lists all accepted TOML dotted paths for config get/set.
var validKeys = map[string]bool{
	"ai.gemini_api_key":  true,
	"ai.gemini_model":    true,
	"appsheet.api_key":   true,
	"appsheet.app_id":    true,
	"appsheet.base_url":  true,
	"gdrive.key_path":    true,
	"catalog.type":       true,
	"catalog.csv_path":   true,
	"defaults.log_level": true,
}

// IsValidKey reports whether key is an accepted TOML dotted path.
func IsValidKey(key string) bool {
	return validKeys[key]
}

// FlatMap returns the config values keyed by the flat viper key names
// used by CLI flags and env vars. This bridges the nested TOML structure
// to the flat key namespace used everywhere else in the app.
func (c *Config) FlatMap() map[string]string {
	return map[string]string{
		"gemini-api-key":        c.AI.GeminiAPIKey,
		"gemini-model":          c.AI.GeminiModel,
		"appsheet-api-key":      c.AppSheet.APIKey,
		"appsheet-app-id":       c.AppSheet.AppID,
		"appsheet-base-url":     c.AppSheet.BaseURL,
		"google-drive-key-path": c.GDrive.KeyPath,
		"catalog":               c.Catalog.Type,
		"catalog-csv-path":      c.Catalog.CSVPath,
		"log-level":             c.Defaults.LogLevel,
	}
}

// LoadFromFile reads and unmarshals a TOML config file.
func LoadFromFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file %s: %w", path, err)
	}
	return &cfg, nil
}

// DefaultConfig returns a Config populated with sensible defaults.
func DefaultConfig() Config {
	return Config{
		AI: AIConfig{
			GeminiModel: "gemini-2.0-flash",
		},
		AppSheet: AppSheetConfig{
			BaseURL: "https://www.appsheet.com",
		},
		Catalog: CatalogConfig{
			Type: "appsheet",
		},
	}
}

// InitConfigFile creates a default config file at the given path,
// creating parent directories as needed.
func InitConfigFile(filePath string) error {
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory %s: %w", dir, err)
	}

	cfg := DefaultConfig()

	data, err := toml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal default config: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file %s: %w", filePath, err)
	}

	return nil
}

// DefaultConfigPath returns the default config file path: ~/.invoice_ocr/config.toml.
func DefaultConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine home directory: %w", err)
	}
	return filepath.Join(home, configDir, configFile), nil
}
