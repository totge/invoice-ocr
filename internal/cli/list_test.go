package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestListOptions_Validate(t *testing.T) {
	// Helper to create a temp file (not a directory) for negative tests.
	createTempFile := func(t *testing.T) string {
		t.Helper()
		dir := t.TempDir()
		path := filepath.Join(dir, "not-a-dir.txt")
		if err := os.WriteFile(path, []byte("content"), 0644); err != nil {
			t.Fatal(err)
		}
		return path
	}

	t.Run("valid local source", func(t *testing.T) {
		opts := ListOptions{
			Source: "local",
			Path:   t.TempDir(),
			Target: "terminal",
			Limit:  10,
		}
		if err := opts.Validate(); err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})

	t.Run("valid gdrive source", func(t *testing.T) {
		opts := ListOptions{
			Source:        "gdrive",
			Path:          "some-folder-id",
			Target:        "terminal",
			Limit:         10,
			GDriveKeyPath: "/path/to/key.json",
		}
		if err := opts.Validate(); err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})

	t.Run("invalid source", func(t *testing.T) {
		opts := ListOptions{
			Source: "dropbox",
			Path:   "some-path",
			Target: "terminal",
			Limit:  10,
		}
		err := opts.Validate()
		if err == nil {
			t.Fatal("expected error for invalid source")
		}
		if !strings.Contains(err.Error(), "invalid --source") {
			t.Errorf("expected error about invalid source, got: %v", err)
		}
	})

	t.Run("local path does not exist", func(t *testing.T) {
		opts := ListOptions{
			Source: "local",
			Path:   "/nonexistent/directory",
			Target: "terminal",
			Limit:  10,
		}
		err := opts.Validate()
		if err == nil {
			t.Fatal("expected error for nonexistent path")
		}
		if !strings.Contains(err.Error(), "does not exist") {
			t.Errorf("expected error about path not existing, got: %v", err)
		}
	})

	t.Run("local path is a file not directory", func(t *testing.T) {
		opts := ListOptions{
			Source: "local",
			Path:   createTempFile(t),
			Target: "terminal",
			Limit:  10,
		}
		err := opts.Validate()
		if err == nil {
			t.Fatal("expected error for file input")
		}
		if !strings.Contains(err.Error(), "not a directory") {
			t.Errorf("expected error about not being a directory, got: %v", err)
		}
	})

	t.Run("invalid target", func(t *testing.T) {
		opts := ListOptions{
			Source: "local",
			Path:   t.TempDir(),
			Target: "json",
			Limit:  10,
		}
		err := opts.Validate()
		if err == nil {
			t.Fatal("expected error for invalid target")
		}
		if !strings.Contains(err.Error(), "invalid --target") {
			t.Errorf("expected error about invalid target, got: %v", err)
		}
	})

	t.Run("negative limit", func(t *testing.T) {
		opts := ListOptions{
			Source: "local",
			Path:   t.TempDir(),
			Target: "terminal",
			Limit:  -1,
		}
		err := opts.Validate()
		if err == nil {
			t.Fatal("expected error for negative limit")
		}
		if !strings.Contains(err.Error(), "limit") {
			t.Errorf("expected error about limit, got: %v", err)
		}
	})

	t.Run("zero limit is valid (unlimited)", func(t *testing.T) {
		opts := ListOptions{
			Source: "local",
			Path:   t.TempDir(),
			Target: "terminal",
			Limit:  0,
		}
		if err := opts.Validate(); err != nil {
			t.Errorf("expected no error for zero limit, got: %v", err)
		}
	})
}
