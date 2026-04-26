package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	toml "github.com/pelletier/go-toml/v2"
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

// KeyValue represents a single config entry for listing.
type KeyValue struct {
	Key   string
	Value string
}

func defaultConfig() Config {
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

func getValue(cfg *Config, key string) (string, error) {
	switch key {
	case "ai.gemini_api_key":
		return cfg.AI.GeminiAPIKey, nil
	case "ai.gemini_model":
		return cfg.AI.GeminiModel, nil
	case "appsheet.api_key":
		return cfg.AppSheet.APIKey, nil
	case "appsheet.app_id":
		return cfg.AppSheet.AppID, nil
	case "appsheet.base_url":
		return cfg.AppSheet.BaseURL, nil
	case "gdrive.key_path":
		return cfg.GDrive.KeyPath, nil
	case "catalog.type":
		return cfg.Catalog.Type, nil
	case "catalog.csv_path":
		return cfg.Catalog.CSVPath, nil
	case "defaults.log_level":
		return cfg.Defaults.LogLevel, nil
	default:
		return "", fmt.Errorf("invalid config key: %q", key)
	}
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

// InitConfigFile creates a default config file at the given path,
// creating parent directories as needed.
func InitConfigFile(filePath string) error {
	cfg := defaultConfig()
	return saveToFile(&cfg, filePath)
}

// GetValue reads the config file and returns the value for a specific key.
func GetValue(path string, key string) (string, error) {
	cfg, err := LoadFromFile(path)
	if err != nil {
		return "", err
	}
	return getValue(cfg, key)
}

// SetValue reads the config file, sets a key to a value, and writes it back.
func SetValue(path string, key string, value string) error {
	cfg, err := LoadFromFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			d := defaultConfig()
			cfg = &d
		} else {
			return err
		}
	}

	switch key {
	case "ai.gemini_api_key":
		cfg.AI.GeminiAPIKey = value
	case "ai.gemini_model":
		cfg.AI.GeminiModel = value
	case "appsheet.api_key":
		cfg.AppSheet.APIKey = value
	case "appsheet.app_id":
		cfg.AppSheet.AppID = value
	case "appsheet.base_url":
		cfg.AppSheet.BaseURL = value
	case "gdrive.key_path":
		cfg.GDrive.KeyPath = value
	case "catalog.type":
		cfg.Catalog.Type = value
	case "catalog.csv_path":
		cfg.Catalog.CSVPath = value
	case "defaults.log_level":
		cfg.Defaults.LogLevel = value
	default:
		return fmt.Errorf("invalid config key: %q", key)
	}

	return saveToFile(cfg, path)
}

// ListValues reads the config file and returns all key-value pairs in order.
func ListValues(path string) ([]KeyValue, error) {
	cfg, err := LoadFromFile(path)
	if err != nil {
		return nil, err
	}

	return []KeyValue{
		{"ai.gemini_api_key", cfg.AI.GeminiAPIKey},
		{"ai.gemini_model", cfg.AI.GeminiModel},
		{"appsheet.api_key", cfg.AppSheet.APIKey},
		{"appsheet.app_id", cfg.AppSheet.AppID},
		{"appsheet.base_url", cfg.AppSheet.BaseURL},
		{"gdrive.key_path", cfg.GDrive.KeyPath},
		{"catalog.type", cfg.Catalog.Type},
		{"catalog.csv_path", cfg.Catalog.CSVPath},
		{"defaults.log_level", cfg.Defaults.LogLevel},
	}, nil
}

func saveToFile(cfg *Config, path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory %s: %w", dir, err)
	}

	data, err := toml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file %s: %w", path, err)
	}
	return nil
}
