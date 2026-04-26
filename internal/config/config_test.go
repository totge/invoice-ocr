package config

import (
	"os"
	"path/filepath"
	"testing"
)

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

func TestGetValue(t *testing.T) {
	dir := t.TempDir()
	path := writeTOML(t, dir, `
[ai]
gemini_api_key = 'test-key'
gemini_model = 'gemini-pro'
`)

	val, err := GetValue(path, "ai.gemini_api_key")
	if err != nil {
		t.Fatalf("GetValue error: %v", err)
	}
	if val != "test-key" {
		t.Errorf("got %q, want %q", val, "test-key")
	}
}

func TestGetValue_InvalidKey(t *testing.T) {
	dir := t.TempDir()
	path := writeTOML(t, dir, `[ai]
gemini_api_key = 'x'
`)

	_, err := GetValue(path, "unknown.key")
	if err == nil {
		t.Fatal("expected error for invalid key")
	}
}

func TestGetValue_MissingFile(t *testing.T) {
	_, err := GetValue("/nonexistent/config.toml", "ai.gemini_api_key")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestSetValue(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")

	// Init a config file first
	if err := InitConfigFile(path); err != nil {
		t.Fatalf("InitConfigFile error: %v", err)
	}

	// Set a value
	if err := SetValue(path, "ai.gemini_api_key", "new-key"); err != nil {
		t.Fatalf("SetValue error: %v", err)
	}

	// Read it back
	val, err := GetValue(path, "ai.gemini_api_key")
	if err != nil {
		t.Fatalf("GetValue error: %v", err)
	}
	if val != "new-key" {
		t.Errorf("got %q, want %q", val, "new-key")
	}

	// Verify other defaults are preserved
	val, err = GetValue(path, "ai.gemini_model")
	if err != nil {
		t.Fatalf("GetValue error: %v", err)
	}
	if val != "gemini-2.0-flash" {
		t.Errorf("default not preserved: got %q, want %q", val, "gemini-2.0-flash")
	}
}

func TestSetValue_InvalidKey(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	if err := InitConfigFile(path); err != nil {
		t.Fatalf("InitConfigFile error: %v", err)
	}

	err := SetValue(path, "unknown.key", "value")
	if err == nil {
		t.Fatal("expected error for invalid key")
	}
}

func TestSetValue_CreatesFileIfMissing(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sub", "config.toml")

	if err := SetValue(path, "ai.gemini_api_key", "created-key"); err != nil {
		t.Fatalf("SetValue error: %v", err)
	}

	val, err := GetValue(path, "ai.gemini_api_key")
	if err != nil {
		t.Fatalf("GetValue error: %v", err)
	}
	if val != "created-key" {
		t.Errorf("got %q, want %q", val, "created-key")
	}
}

func TestListValues(t *testing.T) {
	dir := t.TempDir()
	path := writeTOML(t, dir, `
[ai]
gemini_api_key = 'k1'
gemini_model = 'm1'

[appsheet]
api_key = 'k2'
app_id = 'a1'
base_url = 'u1'

[gdrive]
key_path = 'p1'

[catalog]
type = 'csv'
csv_path = 'c1'

[defaults]
log_level = 'debug'
`)

	entries, err := ListValues(path)
	if err != nil {
		t.Fatalf("ListValues error: %v", err)
	}

	if len(entries) != 9 {
		t.Fatalf("got %d entries, want 9", len(entries))
	}

	// Verify order and values
	if entries[0].Key != "ai.gemini_api_key" || entries[0].Value != "k1" {
		t.Errorf("first entry = {%q, %q}, want {%q, %q}", entries[0].Key, entries[0].Value, "ai.gemini_api_key", "k1")
	}
	if entries[8].Key != "defaults.log_level" || entries[8].Value != "debug" {
		t.Errorf("last entry = {%q, %q}, want {%q, %q}", entries[8].Key, entries[8].Value, "defaults.log_level", "debug")
	}
}

func TestSaveToFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nested", "config.toml")

	cfg := defaultConfig()
	cfg.AI.GeminiAPIKey = "save-test"

	if err := saveToFile(&cfg, path); err != nil {
		t.Fatalf("saveToFile error: %v", err)
	}

	loaded, err := LoadFromFile(path)
	if err != nil {
		t.Fatalf("LoadFromFile error: %v", err)
	}
	if loaded.AI.GeminiAPIKey != "save-test" {
		t.Errorf("got %q, want %q", loaded.AI.GeminiAPIKey, "save-test")
	}

	// Verify permissions
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat error: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0600 {
		t.Errorf("file permissions = %o, want 0600", perm)
	}
}
