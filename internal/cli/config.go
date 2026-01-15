package cli

import (
	"flag"
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

// config struct to hold application configuration
type config struct {
	GeminiApiKey    string
	GeminiModel     string
	AppSheetApiKey  string
	AppSheetAppId   string
	AppSheetBaseUrl string
	GDriveKeyPath   string
	LogLevel        string
	Verbose         bool
}

func newConfig() *config {
	// 1. Load .env file
	_ = godotenv.Load()

	// 2. Initialize with Defaults
	c := &config{
		GeminiModel:     "gemini-2.0-flash",
		AppSheetBaseUrl: "https://www.appsheet.com",
		LogLevel:        "error",
	}

	// 3. Overwrite with Environment Variables
	get := func(key string) string {
		return os.Getenv("INVOICE_CATEGORIZER_" + key)
	}

	// getting the environment variables or setting defaults
	c.GeminiApiKey = get("GEMINI_API_KEY")
	c.AppSheetApiKey = get("APPSHEET_API_KEY")
	c.AppSheetAppId = get("APPSHEET_APP_ID")
	c.GDriveKeyPath = get("GOOGLE_DRIVE_KEY_PATH")
	if model := get("GEMINI_MODEL"); model != "" {
		c.GeminiModel = model
	}
	if url := get("APPSHEET_BASE_URL"); url != "" {
		c.AppSheetBaseUrl = url
	}
	if logLevel := get("LOG_LEVEL"); logLevel != "" {
		c.LogLevel = logLevel
	}

	return c
}

// RegisterFlags binds the config fields to CLI flags on the provided FlagSet.
// This allows any command to "inherit" these standard config flags.
func (c *config) registerConfigFlags(fs *flag.FlagSet) {
	fs.StringVar(&c.GeminiApiKey, "gemini-key", c.GeminiApiKey, "Gemini API Key")
	fs.StringVar(&c.GeminiModel, "gemini-model", c.GeminiModel, "Gemini model name")
	fs.StringVar(&c.AppSheetApiKey, "appsheet-key", c.AppSheetApiKey, "AppSheet API Key")
	fs.StringVar(&c.AppSheetAppId, "appsheet-id", c.AppSheetAppId, "AppSheet App ID")
	fs.StringVar(&c.AppSheetBaseUrl, "appsheet-url", c.AppSheetBaseUrl, "AppSheet Base URL")
	fs.StringVar(&c.GDriveKeyPath, "gdrive-key-path", c.GDriveKeyPath, "Path to Google Drive key")
	fs.BoolVar(&c.Verbose, "verbose", c.Verbose, "Enable verbose debug logging")
}

// validate checks if the final configuration (after Env and Flags) is valid.
func (c *config) validate() error {
	// checking required config values
	var missing []string
	if c.GeminiApiKey == "" {
		missing = append(missing, "Gemini API Key")
	}
	if c.AppSheetApiKey == "" {
		missing = append(missing, "AppSheet API Key")
	}
	if c.AppSheetAppId == "" {
		missing = append(missing, "AppSheet App ID")
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required configuration: %v. \nSet via environment variables (INVOICE_CATEGORIZER_...) or CLI flags", missing)
	}

	// validate loglevel
	switch c.LogLevel {
	case "debug", "info", "error":
		// valid
	default:
		return fmt.Errorf("invalid log level: %s. \nIt must be one of [\"debug\" \"info\" \"error\"] set via environment variable (INVOICE_CATEGORIZER_LOG_LEVEL) or CLI flag --verbose", c.LogLevel)
	}

	return nil
}

// --- Helper type for embedded config ---
type baseConfig struct {
	cfg *config // Pointer to the shared config instance
}

// EmbedConfig sets the internal config reference.
func (b *baseConfig) EmbedConfig(c *config) {
	b.cfg = c
}

// GetConfig returns the embedded config.
func (b *baseConfig) GetConfig() *config {
	return b.cfg
}
