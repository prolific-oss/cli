package shared

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseIDFileReturnsIDs(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "ids.csv")
	if err := os.WriteFile(filePath, []byte("id-1\nid-2\nid-3\n"), 0600); err != nil {
		t.Fatalf("failed to create temp file: %s", err)
	}

	ids, err := ParseIDFile(filePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []string{"id-1", "id-2", "id-3"}
	if len(ids) != len(expected) {
		t.Fatalf("expected %d ids, got %d", len(expected), len(ids))
	}
	for i, id := range ids {
		if id != expected[i] {
			t.Errorf("expected ids[%d] = %q, got %q", i, expected[i], id)
		}
	}
}

func TestParseIDFileSkipsBlankLines(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "ids.csv")
	if err := os.WriteFile(filePath, []byte("id-1\n\n  \nid-2\n"), 0600); err != nil {
		t.Fatalf("failed to create temp file: %s", err)
	}

	ids, err := ParseIDFile(filePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(ids) != 2 {
		t.Fatalf("expected 2 ids, got %d: %v", len(ids), ids)
	}
}

func TestParseIDFileErrorsOnEmptyFile(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "empty.csv")
	if err := os.WriteFile(filePath, []byte(""), 0600); err != nil {
		t.Fatalf("failed to create temp file: %s", err)
	}

	_, err := ParseIDFile(filePath)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "file is empty: "+filePath {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestParseIDFileErrorsOnMissingFile(t *testing.T) {
	_, err := ParseIDFile("/nonexistent/path/file.csv")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestParseIDFileErrorsOnWhitespaceOnlyFile(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "whitespace.csv")
	if err := os.WriteFile(filePath, []byte("   \n  \n"), 0600); err != nil {
		t.Fatalf("failed to create temp file: %s", err)
	}

	_, err := ParseIDFile(filePath)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
