package cli

import (
	"flag"
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

// Config struct to hold application configuration
type Config struct {
	GeminiApiKey    string
	GeminiModel     string
	AppSheetApiKey  string
	AppSheetAppId   string
	AppSheetBaseUrl string
}

func NewConfig() *Config {
	// 1. Load .env file
	_ = godotenv.Load()

	// 2. Initialize with Defaults
	c := &Config{
		GeminiModel:     "gemini-2.0-flash",
		AppSheetBaseUrl: "https://www.appsheet.com",
	}

	// 3. Overwrite with Environment Variables
	get := func(key string) string {
		return os.Getenv("INVOICE_CATEGORIZER_" + key)
	}

	// getting the environment variables
	c.GeminiApiKey = get("GEMINI_API_KEY")
	c.AppSheetApiKey = get("APPSHEET_API_KEY")
	c.AppSheetAppId = get("APPSHEET_APP_ID")
	if model := get("GEMINI_MODEL"); model != "" {
		c.GeminiModel = model
	}
	if url := get("APPSHEET_BASE_URL"); url != "" {
		c.AppSheetBaseUrl = url
	}

	return c
}

// RegisterFlags binds the config fields to CLI flags on the provided FlagSet.
// This allows any command to "inherit" these standard config flags.
func (c *Config) RegisterConfigFlags(fs *flag.FlagSet) {
	fs.StringVar(&c.GeminiApiKey, "gemini-key", c.GeminiApiKey, "Gemini API Key")
	fs.StringVar(&c.GeminiModel, "gemini-model", c.GeminiModel, "Gemini model name")
	fs.StringVar(&c.AppSheetApiKey, "appsheet-key", c.AppSheetApiKey, "AppSheet API Key")
	fs.StringVar(&c.AppSheetAppId, "appsheet-id", c.AppSheetAppId, "AppSheet App ID")
	fs.StringVar(&c.AppSheetBaseUrl, "appsheet-url", c.AppSheetBaseUrl, "AppSheet Base URL")
}

// Validate checks if the final configuration (after Env and Flags) is valid.
func (c *Config) Validate() error {
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
	return nil
}
