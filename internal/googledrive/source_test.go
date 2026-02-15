package googledrive

import (
	"strings"
	"testing"
)

func TestParseURI(t *testing.T) {
	t.Run("valid uri", func(t *testing.T) {
		id, err := ParseURI("gdrive://abc123")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if id != "abc123" {
			t.Errorf("expected id %q, got %q", "abc123", id)
		}
	})

	t.Run("valid uri with path-like id", func(t *testing.T) {
		id, err := ParseURI("gdrive://folder/subfolder/file")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if id != "folder/subfolder/file" {
			t.Errorf("expected id %q, got %q", "folder/subfolder/file", id)
		}
	})

	t.Run("missing prefix", func(t *testing.T) {
		_, err := ParseURI("abc123")
		if err == nil {
			t.Fatal("expected error for missing prefix")
		}
		if !strings.Contains(err.Error(), "invalid gdrive uri") {
			t.Errorf("expected error about invalid uri, got: %v", err)
		}
	})

	t.Run("empty id after prefix", func(t *testing.T) {
		_, err := ParseURI("gdrive://")
		if err == nil {
			t.Fatal("expected error for empty id")
		}
		if !strings.Contains(err.Error(), "missing ID") {
			t.Errorf("expected error about missing ID, got: %v", err)
		}
	})

	t.Run("wrong prefix", func(t *testing.T) {
		_, err := ParseURI("s3://bucket/key")
		if err == nil {
			t.Fatal("expected error for wrong prefix")
		}
		if !strings.Contains(err.Error(), "invalid gdrive uri") {
			t.Errorf("expected error about invalid uri, got: %v", err)
		}
	})

	t.Run("empty string", func(t *testing.T) {
		_, err := ParseURI("")
		if err == nil {
			t.Fatal("expected error for empty string")
		}
	})
}
