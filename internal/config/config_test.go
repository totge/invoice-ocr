package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfigPath(t *testing.T) {
	path, err := DefaultConfigPath()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	home, _ := os.UserHomeDir()
	expected := filepath.Join(home, ".invoice_ocr", "config.toml")
	if path != expected {
		t.Errorf("got %s, want %s", path, expected)
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.AI.GeminiModel != "gemini-2.0-flash" {
		t.Errorf("GeminiModel = %q, want %q", cfg.AI.GeminiModel, "gemini-2.0-flash")
	}
	if cfg.AppSheet.BaseURL != "https://www.appsheet.com" {
		t.Errorf("BaseURL = %q, want %q", cfg.AppSheet.BaseURL, "https://www.appsheet.com")
	}
	if cfg.Catalog.Type != "appsheet" {
		t.Errorf("Catalog.Type = %q, want %q", cfg.Catalog.Type, "appsheet")
	}
}

func writeTOML(t *testing.T, dir, content string) string {
	t.Helper()
	path := filepath.Join(dir, "config.toml")
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestLoadFromFile_ValidFull(t *testing.T) {
	dir := t.TempDir()
	path := writeTOML(t, dir, `
[ai]
gemini_api_key = 'key-123'
gemini_model = 'gemini-pro'

[appsheet]
api_key = 'as-key'
app_id = 'as-app'
base_url = 'https://custom.appsheet.com'

[gdrive]
key_path = '/keys/sa.json'

[catalog]
type = 'csv'
csv_path = '/data/catalog.csv'

[defaults]
log_level = 'debug'
`)

	cfg, err := LoadFromFile(path)
	if err != nil {
		t.Fatalf("LoadFromFile error: %v", err)
	}

	checks := map[string]string{
		"AI.GeminiAPIKey":    cfg.AI.GeminiAPIKey,
		"AI.GeminiModel":     cfg.AI.GeminiModel,
		"AppSheet.APIKey":    cfg.AppSheet.APIKey,
		"AppSheet.AppID":     cfg.AppSheet.AppID,
		"AppSheet.BaseURL":   cfg.AppSheet.BaseURL,
		"GDrive.KeyPath":     cfg.GDrive.KeyPath,
		"Catalog.Type":       cfg.Catalog.Type,
		"Catalog.CSVPath":    cfg.Catalog.CSVPath,
		"Defaults.LogLevel":  cfg.Defaults.LogLevel,
	}
	expected := map[string]string{
		"AI.GeminiAPIKey":    "key-123",
		"AI.GeminiModel":     "gemini-pro",
		"AppSheet.APIKey":    "as-key",
		"AppSheet.AppID":     "as-app",
		"AppSheet.BaseURL":   "https://custom.appsheet.com",
		"GDrive.KeyPath":     "/keys/sa.json",
		"Catalog.Type":       "csv",
		"Catalog.CSVPath":    "/data/catalog.csv",
		"Defaults.LogLevel":  "debug",
	}

	for field, got := range checks {
		if got != expected[field] {
			t.Errorf("%s = %q, want %q", field, got, expected[field])
		}
	}
}

func TestLoadFromFile_MissingFile(t *testing.T) {
	_, err := LoadFromFile("/nonexistent/config.toml")
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
}

func TestLoadFromFile_MalformedTOML(t *testing.T) {
	dir := t.TempDir()
	path := writeTOML(t, dir, `[ai
broken toml`)

	_, err := LoadFromFile(path)
	if err == nil {
		t.Fatal("expected error for malformed TOML")
	}
}

func TestInitConfigFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sub", "nested", "config.toml")

	if err := InitConfigFile(path); err != nil {
		t.Fatalf("InitConfigFile error: %v", err)
	}

	// Verify file was created and is readable
	cfg, err := LoadFromFile(path)
	if err != nil {
		t.Fatalf("LoadFromFile error: %v", err)
	}

	// Should contain defaults
	if cfg.AI.GeminiModel != "gemini-2.0-flash" {
		t.Errorf("GeminiModel = %q, want default", cfg.AI.GeminiModel)
	}

	// Verify file permissions
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat error: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0600 {
		t.Errorf("file permissions = %o, want 0600", perm)
	}

	// Verify directory permissions
	dirInfo, err := os.Stat(filepath.Dir(path))
	if err != nil {
		t.Fatalf("Stat dir error: %v", err)
	}
	if perm := dirInfo.Mode().Perm(); perm != 0700 {
		t.Errorf("directory permissions = %o, want 0700", perm)
	}
}

func TestFlatMap(t *testing.T) {
	cfg := Config{
		AI:       AIConfig{GeminiAPIKey: "k1", GeminiModel: "m1"},
		AppSheet: AppSheetConfig{APIKey: "k2", AppID: "a1", BaseURL: "u1"},
		GDrive:   GDriveConfig{KeyPath: "p1"},
		Catalog:  CatalogConfig{Type: "csv", CSVPath: "c1"},
		Defaults: DefaultsConfig{LogLevel: "debug"},
	}

	fm := cfg.FlatMap()

	expected := map[string]string{
		"gemini-api-key":        "k1",
		"gemini-model":          "m1",
		"appsheet-api-key":      "k2",
		"appsheet-app-id":       "a1",
		"appsheet-base-url":     "u1",
		"google-drive-key-path": "p1",
		"catalog":               "csv",
		"catalog-csv-path":      "c1",
		"log-level":             "debug",
	}

	for key, want := range expected {
		got, ok := fm[key]
		if !ok {
			t.Errorf("FlatMap missing key %q", key)
			continue
		}
		if got != want {
			t.Errorf("FlatMap[%q] = %q, want %q", key, got, want)
		}
	}

	if len(fm) != len(expected) {
		t.Errorf("FlatMap has %d keys, want %d", len(fm), len(expected))
	}
}

func TestIsValidKey(t *testing.T) {
	valid := []string{
		"ai.gemini_api_key", "ai.gemini_model",
		"appsheet.api_key", "appsheet.app_id", "appsheet.base_url",
		"gdrive.key_path",
		"catalog.type", "catalog.csv_path",
		"defaults.log_level",
	}
	for _, key := range valid {
		if !IsValidKey(key) {
			t.Errorf("IsValidKey(%q) = false, want true", key)
		}
	}

	invalid := []string{"unknown", "ai.unknown", "gemini-api-key", ""}
	for _, key := range invalid {
		if IsValidKey(key) {
			t.Errorf("IsValidKey(%q) = true, want false", key)
		}
	}
}
