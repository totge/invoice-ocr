package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCategorizeOptions_Validate(t *testing.T) {
	createTempFile := func(t *testing.T) string {
		t.Helper()
		dir := t.TempDir()
		path := filepath.Join(dir, "receipt.json")
		if err := os.WriteFile(path, []byte(`{}`), 0644); err != nil {
			t.Fatal(err)
		}
		return path
	}

	validOpts := func(t *testing.T) CategorizeOptions {
		t.Helper()
		return CategorizeOptions{
			Source:         "local",
			Input:          createTempFile(t),
			Target:         "file",
			Output:         "output.json",
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
		if err := opts.Validate(); err != nil {
			t.Errorf("expected no error, got: %v", err)
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
		opts.Input = "/nonexistent/receipt.json"
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
}
