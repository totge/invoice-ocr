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

// KeyEntry maps a TOML dotted path to its flat Viper key.
type KeyEntry struct {
	TOMLPath string
	ViperKey string
}

// ValidKeys lists all accepted TOML dotted paths for config get/set.
var ValidKeys = map[string]bool{
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

func IsValidKey(key string) bool {
	return ValidKeys[key]
}

// FlatMap returns the config values keyed by the flat viper key names
// used by CLI flags and env vars. This bridges the nested TOML structure
// to the flat key namespace used everywhere else in the app.
func (c *Config) FlatMap() map[string]string {
	return map[string]string{
		"gemini-api-key":       c.AI.GeminiAPIKey,
		"gemini-model":         c.AI.GeminiModel,
		"appsheet-api-key":     c.AppSheet.APIKey,
		"appsheet-app-id":      c.AppSheet.AppID,
		"appsheet-base-url":    c.AppSheet.BaseURL,
		"google-drive-key-path": c.GDrive.KeyPath,
		"catalog":              c.Catalog.Type,
		"catalog-csv-path":     c.Catalog.CSVPath,
		"log-level":            c.Defaults.LogLevel,
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

// Cretes an empty config file with the defaults set
func InitConfigFile(path string, force bool) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory %s: %w", dir, err)
	}

	cfg := DefaultConfig()

	data, err := toml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal default config: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file %s: %w", path, err)
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

// ---------------------
// Old implementation below, should be deleted later
// ---------------------

// // rethink name
// type ConfigManager struct {
// 	path string
// }

// func NewConfigManager(path string) *ConfigManager {
// 	return &ConfigManager{path: path}
// }

// // reads the file bound to the instance into a Config struct
// func (cm *ConfigManager) readFile() (*Config, error) {
// 	data, err := os.ReadFile(cm.path)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to read config file %s: %w", cm.path, err)
// 	}

// 	var cfg Config
// 	if err := toml.Unmarshal(data, &cfg); err != nil {
// 		return nil, fmt.Errorf("failed to parse config file %s: %w", cm.path, err)
// 	}
// 	return &cfg, nil
// }

// // saves the Config provided into the file bound to the instance
// func (cm *ConfigManager) saveFile(cfg *Config) error {
// 	data, err := toml.Marshal(cfg)
// 	if err != nil {
// 		return fmt.Errorf("failed to marshal config: %w", err)
// 	}

// 	if err := os.WriteFile(cm.path, data, 0600); err != nil {
// 		return fmt.Errorf("failed to write config file %s: %w", cm.path, err)
// 	}

// 	return nil
// }

// // Check for folder (create if not exists)
// // Check for file, create if not exists
// // Create cf manager with file path
// // Create new default config struct
// // write to file
// func (cm *ConfigManager) InitConfigFile(force bool) error {
// 	dir := filepath.Dir(cm.path)
// 	if err := os.MkdirAll(dir, 0700); err != nil {
// 		return fmt.Errorf("failed to create config directory %s: %w", dir, err)
// 	}

// 	cfg := DefaultConfig()

// 	// file not exist or force
// 	if _, err := os.Stat(cm.path); errors.Is(err, os.ErrNotExist) || force {
// 		// run saveFile method and return result
// 		return cm.saveFile(&cfg)
// 	} else if err != nil {
// 		return fmt.Errorf("failed to save default config file: %w", err)
// 	}

// 	return nil
// }

// func (cm *ConfigManager) GetConfig() (*Config, error) {
// 	return cm.readFile()
// }

// func (cm *ConfigManager) GetValue(key string) (string, error) {
// 	cfg, err := cm.readFile()
// 	if err != nil {
// 		return "", err
// 	}

// 	switch key {
// 	case "ai.gemini_api_key":
// 		return cfg.AI.GeminiAPIKey, nil
// 	case "ai.gemini_model":
// 		return cfg.AI.GeminiModel, nil
// 	case "appsheet.api_key":
// 		return cfg.AppSheet.APIKey, nil
// 	case "appsheet.app_id":
// 		return cfg.AppSheet.AppID, nil
// 	case "appsheet.base_url":
// 		return cfg.AppSheet.BaseURL, nil
// 	case "gdrive.key_path":
// 		return cfg.GDrive.KeyPath, nil
// 	case "catalog.type":
// 		return cfg.Catalog.Type, nil
// 	case "catalog.csv_path":
// 		return cfg.Catalog.CSVPath, nil
// 	case "defaults.log_level":
// 		return cfg.Defaults.LogLevel, nil

// 	default:
// 		return "", fmt.Errorf("unknown config key: %s", key)
// 	}
// }

// func (cm *ConfigManager) SetValue(key string, value string) error {
// 	cfg, err := cm.readFile()
// 	if err != nil {
// 		return err
// 	}

// 	switch key {
// 	case "ai.gemini_api_key":
// 		cfg.AI.GeminiAPIKey = value
// 	case "ai.gemini_model":
// 		cfg.AI.GeminiModel = value
// 	case "appsheet.api_key":
// 		cfg.AppSheet.APIKey = value
// 	case "appsheet.app_id":
// 		cfg.AppSheet.AppID = value
// 	case "appsheet.base_url":
// 		cfg.AppSheet.BaseURL = value
// 	case "gdrive.key_path":
// 		cfg.GDrive.KeyPath = value
// 	case "catalog.type":
// 		cfg.Catalog.Type = value
// 	case "catalog.csv_path":
// 		cfg.Catalog.CSVPath = value
// 	case "defaults.log_level":
// 		cfg.Defaults.LogLevel = value
// 	default:
// 		return fmt.Errorf("unknown config key: %s", key)
// 	}
// 	return nil

// }

// // KeyMapping defines the mapping from TOML section.key paths to the flat Viper
// // keys used throughout the codebase (e.g., v.GetString("gemini-api-key")).
// var KeyMapping = []KeyEntry{
// 	{"ai.gemini_api_key", "gemini-api-key"},
// 	{"ai.gemini_model", "gemini-model"},
// 	{"appsheet.api_key", "appsheet-api-key"},
// 	{"appsheet.app_id", "appsheet-app-id"},
// 	{"appsheet.base_url", "appsheet-base-url"},
// 	{"gdrive.key_path", "google-drive-key-path"},
// 	{"catalog.type", "catalog"},
// 	{"catalog.csv_path", "catalog-csv-path"},
// 	{"defaults.log_level", "log-level"},
// }

// // ConfigPathIn returns <baseDir>/.invoice_ocr/config.toml.
// func ConfigPathIn(baseDir string) string {
// 	return filepath.Join(baseDir, configDir, configFile)
// }

// func applyToViper(v *viper.Viper, cfg *Config) {

// 	v.SetDefault("gemini-api-key", cfg.AI.GeminiAPIKey)
// 	v.SetDefault("gemini-model", cfg.AI.GeminiModel)
// 	v.SetDefault("appsheet-api-key", cfg.AppSheet.APIKey)
// 	v.SetDefault("appsheet-app-id", cfg.AppSheet.AppID)
// 	v.SetDefault("appsheet-base-url", cfg.AppSheet.BaseURL)
// 	v.SetDefault("google-drive-key-path", cfg.GDrive.KeyPath)
// 	v.SetDefault("catalog", cfg.Catalog.Type)
// 	v.SetDefault("catalog-csv-path", cfg.Catalog.CSVPath)
// 	v.SetDefault("log-level", cfg.Defaults.LogLevel)

// }

// func tomlToViperKey(tomlPath string) string {
// 	for _, entry := range KeyMapping {
// 		if entry.TOMLPath == tomlPath {
// 			return entry.ViperKey
// 		}
// 	}
// 	return ""
// }

// // LoadFromFile reads and unmarshals a TOML config file without touching Viper.
// func LoadFromFile(path string) (*Config, error) {
// 	data, err := os.ReadFile(path)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to read config file %s: %w", path, err)
// 	}

// 	var cfg Config
// 	if err := toml.Unmarshal(data, &cfg); err != nil {
// 		return nil, fmt.Errorf("failed to parse config file %s: %w", path, err)
// 	}
// 	return &cfg, nil
// }

// // LoadOrDefault loads the config from file. If the file does not exist,
// // it returns DefaultConfig().
// func LoadOrDefault(path string) (*Config, error) {
// 	cfg, err := LoadFromFile(path)
// 	if err != nil {
// 		if os.IsNotExist(err) || strings.Contains(err.Error(), "no such file") {
// 			d := DefaultConfig()
// 			return &d, nil
// 		}
// 		return nil, err
// 	}
// 	return cfg, nil
// }

// // GetValue retrieves a field from the Config by its TOML dotted path.
// func GetValue(cfg *Config, tomlPath string) (string, error) {
// 	switch tomlPath {
// 	case "ai.gemini_api_key":
// 		return cfg.AI.GeminiAPIKey, nil
// 	case "ai.gemini_model":
// 		return cfg.AI.GeminiModel, nil
// 	case "appsheet.api_key":
// 		return cfg.AppSheet.APIKey, nil
// 	case "appsheet.app_id":
// 		return cfg.AppSheet.AppID, nil
// 	case "appsheet.base_url":
// 		return cfg.AppSheet.BaseURL, nil
// 	case "gdrive.key_path":
// 		return cfg.GDrive.KeyPath, nil
// 	case "catalog.type":
// 		return cfg.Catalog.Type, nil
// 	case "catalog.csv_path":
// 		return cfg.Catalog.CSVPath, nil
// 	case "defaults.log_level":
// 		return cfg.Defaults.LogLevel, nil
// 	default:
// 		return "", fmt.Errorf("unknown config key: %s", tomlPath)
// 	}
// }
