package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExtractOptions_Validate(t *testing.T) {
	// Helper to create a temp file for local source tests.
	createTempFile := func(t *testing.T) string {
		t.Helper()
		dir := t.TempDir()
		path := filepath.Join(dir, "receipt.png")
		if err := os.WriteFile(path, []byte("fake image"), 0644); err != nil {
			t.Fatal(err)
		}
		return path
	}

	t.Run("valid local source", func(t *testing.T) {
		opts := ExtractOptions{
			Source:       "local",
			Input:        createTempFile(t),
			Output:       "out.json",
			GeminiAPIKey: "test-key",
			GeminiModel:  "gemini-2.0-flash",
		}
		if err := opts.Validate(); err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})

	t.Run("valid gdrive source", func(t *testing.T) {
		opts := ExtractOptions{
			Source:       "gdrive",
			Input:        "some-file-id",
			Output:       "out.json",
			GeminiAPIKey: "test-key",
			GeminiModel:  "gemini-2.0-flash",
		}
		if err := opts.Validate(); err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})

	t.Run("invalid source", func(t *testing.T) {
		opts := ExtractOptions{
			Source:       "s3",
			Input:        "some-path",
			Output:       "out.json",
			GeminiAPIKey: "test-key",
		}
		err := opts.Validate()
		if err == nil {
			t.Fatal("expected error for invalid source")
		}
		if !strings.Contains(err.Error(), "invalid --source") {
			t.Errorf("expected error about invalid source, got: %v", err)
		}
	})

	t.Run("local file does not exist", func(t *testing.T) {
		opts := ExtractOptions{
			Source:       "local",
			Input:        "/nonexistent/file.png",
			Output:       "out.json",
			GeminiAPIKey: "test-key",
		}
		err := opts.Validate()
		if err == nil {
			t.Fatal("expected error for nonexistent file")
		}
		if !strings.Contains(err.Error(), "does not exist") {
			t.Errorf("expected error about file not existing, got: %v", err)
		}
	})

	t.Run("local input is a directory", func(t *testing.T) {
		opts := ExtractOptions{
			Source:       "local",
			Input:        t.TempDir(),
			Output:       "out.json",
			GeminiAPIKey: "test-key",
		}
		err := opts.Validate()
		if err == nil {
			t.Fatal("expected error for directory input")
		}
		if !strings.Contains(err.Error(), "is a directory") {
			t.Errorf("expected error about directory, got: %v", err)
		}
	})

	t.Run("missing gemini API key", func(t *testing.T) {
		opts := ExtractOptions{
			Source:       "gdrive",
			Input:        "some-id",
			Output:       "out.json",
			GeminiAPIKey: "",
		}
		err := opts.Validate()
		if err == nil {
			t.Fatal("expected error for missing API key")
		}
		if !strings.Contains(err.Error(), "Gemini API key is required") {
			t.Errorf("expected error about missing Gemini key, got: %v", err)
		}
	})

	t.Run("gemini model defaults when empty", func(t *testing.T) {
		opts := ExtractOptions{
			Source:       "gdrive",
			Input:        "some-id",
			Output:       "out.json",
			GeminiAPIKey: "test-key",
			GeminiModel:  "",
		}
		if err := opts.Validate(); err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
		if opts.GeminiModel != "gemini-2.0-flash" {
			t.Errorf("expected GeminiModel to default to 'gemini-2.0-flash', got: %s", opts.GeminiModel)
		}
	})
}
