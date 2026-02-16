package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestProcessOptions_Validate(t *testing.T) {
	createTempFile := func(t *testing.T) string {
		t.Helper()
		dir := t.TempDir()
		path := filepath.Join(dir, "receipt.png")
		if err := os.WriteFile(path, []byte("fake image"), 0644); err != nil {
			t.Fatal(err)
		}
		return path
	}

	// validOpts returns a fully valid ProcessOptions for local+file target.
	validOpts := func(t *testing.T) ProcessOptions {
		t.Helper()
		return ProcessOptions{
			Source:         "local",
			Input:          createTempFile(t),
			Target:         "file",
			Output:         "output.json",
			Catalog:        "appsheet",
			GeminiAPIKey:   "test-key",
			GeminiModel:    "gemini-2.0-flash",
			AppSheetAPIKey: "test-appsheet-key",
			AppSheetAppID:  "test-app-id",
		}
	}

	t.Run("valid local + file target", func(t *testing.T) {
		opts := validOpts(t)
		if err := opts.Validate(); err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})

	t.Run("valid local + appsheet target", func(t *testing.T) {
		opts := validOpts(t)
		opts.Target = "appsheet"
		if err := opts.Validate(); err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})

	t.Run("valid gdrive source", func(t *testing.T) {
		opts := validOpts(t)
		opts.Source = "gdrive"
		opts.Input = "some-file-id"
		opts.GDriveKeyPath = "/path/to/key.json"
		if err := opts.Validate(); err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})

	t.Run("gdrive missing key path", func(t *testing.T) {
		opts := validOpts(t)
		opts.Source = "gdrive"
		opts.Input = "some-file-id"
		opts.GDriveKeyPath = ""
		err := opts.Validate()
		if err == nil {
			t.Fatal("expected error for missing gdrive key path")
		}
		if !strings.Contains(err.Error(), "google drive key is required") {
			t.Errorf("expected error about missing gdrive key, got: %v", err)
		}
	})

	t.Run("invalid source", func(t *testing.T) {
		opts := validOpts(t)
		opts.Source = "ftp"
		err := opts.Validate()
		if err == nil {
			t.Fatal("expected error for invalid source")
		}
		if !strings.Contains(err.Error(), "invalid --source") {
			t.Errorf("expected error about invalid source, got: %v", err)
		}
	})

	t.Run("local file does not exist", func(t *testing.T) {
		opts := validOpts(t)
		opts.Input = "/nonexistent/file.png"
		err := opts.Validate()
		if err == nil {
			t.Fatal("expected error for nonexistent file")
		}
		if !strings.Contains(err.Error(), "does not exist") {
			t.Errorf("expected error about file not existing, got: %v", err)
		}
	})

	t.Run("local input is a directory", func(t *testing.T) {
		opts := validOpts(t)
		opts.Input = t.TempDir()
		err := opts.Validate()
		if err == nil {
			t.Fatal("expected error for directory input")
		}
		if !strings.Contains(err.Error(), "is a directory") {
			t.Errorf("expected error about directory, got: %v", err)
		}
	})

	t.Run("invalid target", func(t *testing.T) {
		opts := validOpts(t)
		opts.Target = "database"
		err := opts.Validate()
		if err == nil {
			t.Fatal("expected error for invalid target")
		}
		if !strings.Contains(err.Error(), "invalid --target") {
			t.Errorf("expected error about invalid target, got: %v", err)
		}
	})

	t.Run("file target missing output", func(t *testing.T) {
		opts := validOpts(t)
		opts.Target = "file"
		opts.Output = ""
		err := opts.Validate()
		if err == nil {
			t.Fatal("expected error for missing output")
		}
		if !strings.Contains(err.Error(), "--output is required") {
			t.Errorf("expected error about required output, got: %v", err)
		}
	})

	t.Run("file target input equals output", func(t *testing.T) {
		opts := validOpts(t)
		opts.Target = "file"
		opts.Output = opts.Input
		err := opts.Validate()
		if err == nil {
			t.Fatal("expected error for input == output")
		}
		if !strings.Contains(err.Error(), "cannot be the same") {
			t.Errorf("expected error about same paths, got: %v", err)
		}
	})

	t.Run("missing gemini API key", func(t *testing.T) {
		opts := validOpts(t)
		opts.GeminiAPIKey = ""
		err := opts.Validate()
		if err == nil {
			t.Fatal("expected error for missing Gemini key")
		}
		if !strings.Contains(err.Error(), "Gemini API key is required") {
			t.Errorf("expected error about missing Gemini key, got: %v", err)
		}
	})

	t.Run("missing appsheet API key", func(t *testing.T) {
		opts := validOpts(t)
		opts.AppSheetAPIKey = ""
		err := opts.Validate()
		if err == nil {
			t.Fatal("expected error for missing AppSheet key")
		}
		if !strings.Contains(err.Error(), "AppSheet API key is required") {
			t.Errorf("expected error about missing AppSheet key, got: %v", err)
		}
	})

	t.Run("missing appsheet app ID", func(t *testing.T) {
		opts := validOpts(t)
		opts.AppSheetAppID = ""
		err := opts.Validate()
		if err == nil {
			t.Fatal("expected error for missing AppSheet App ID")
		}
		if !strings.Contains(err.Error(), "AppSheet App ID is required") {
			t.Errorf("expected error about missing AppSheet App ID, got: %v", err)
		}
	})

	t.Run("gemini model defaults when empty", func(t *testing.T) {
		opts := validOpts(t)
		opts.GeminiModel = ""
		if err := opts.Validate(); err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
		if opts.GeminiModel != "gemini-2.0-flash" {
			t.Errorf("expected GeminiModel to default to 'gemini-2.0-flash', got: %s", opts.GeminiModel)
		}
	})

	t.Run("catalog csv with valid path", func(t *testing.T) {
		opts := validOpts(t)
		opts.Catalog = "csv"
		opts.CatalogCSVPath = createTempFile(t) // reuse helper, any file works for validation
		opts.AppSheetAPIKey = ""
		opts.AppSheetAppID = ""
		if err := opts.Validate(); err != nil {
			t.Errorf("expected no error for csv catalog without appsheet keys, got: %v", err)
		}
	})

	t.Run("catalog csv missing path", func(t *testing.T) {
		opts := validOpts(t)
		opts.Catalog = "csv"
		opts.CatalogCSVPath = ""
		err := opts.Validate()
		if err == nil {
			t.Fatal("expected error for missing catalog CSV path")
		}
		if !strings.Contains(err.Error(), "--catalog-csv-path is required") {
			t.Errorf("expected error about missing csv path, got: %v", err)
		}
	})

	t.Run("catalog csv path does not exist", func(t *testing.T) {
		opts := validOpts(t)
		opts.Catalog = "csv"
		opts.CatalogCSVPath = "/nonexistent/catalog.csv"
		err := opts.Validate()
		if err == nil {
			t.Fatal("expected error for nonexistent catalog CSV")
		}
		if !strings.Contains(err.Error(), "does not exist") {
			t.Errorf("expected error about file not existing, got: %v", err)
		}
	})

	t.Run("catalog csv path is a directory", func(t *testing.T) {
		opts := validOpts(t)
		opts.Catalog = "csv"
		opts.CatalogCSVPath = t.TempDir()
		err := opts.Validate()
		if err == nil {
			t.Fatal("expected error for directory catalog CSV path")
		}
		if !strings.Contains(err.Error(), "is a directory") {
			t.Errorf("expected error about directory, got: %v", err)
		}
	})

	t.Run("catalog csv + appsheet target still requires appsheet keys", func(t *testing.T) {
		opts := validOpts(t)
		opts.Catalog = "csv"
		opts.CatalogCSVPath = createTempFile(t)
		opts.Target = "appsheet"
		opts.AppSheetAPIKey = ""
		err := opts.Validate()
		if err == nil {
			t.Fatal("expected error for missing AppSheet key with appsheet target")
		}
		if !strings.Contains(err.Error(), "AppSheet API key is required") {
			t.Errorf("expected error about missing AppSheet key, got: %v", err)
		}
	})

	t.Run("invalid catalog value", func(t *testing.T) {
		opts := validOpts(t)
		opts.Catalog = "database"
		err := opts.Validate()
		if err == nil {
			t.Fatal("expected error for invalid catalog")
		}
		if !strings.Contains(err.Error(), "invalid --catalog") {
			t.Errorf("expected error about invalid catalog, got: %v", err)
		}
	})
}
