package participantgroup

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseParticipantFile_ValidCSV(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "participants.csv")
	content := "abc123\ndef456\n"
	if err := os.WriteFile(filePath, []byte(content), 0600); err != nil {
		t.Fatalf("failed to create temp file: %s", err)
	}

	result, err := parseParticipantFile(filePath)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	expected := []string{"abc123", "def456"}
	if len(result) != len(expected) {
		t.Fatalf("expected %d IDs, got %d", len(expected), len(result))
	}
	for i, id := range expected {
		if result[i] != id {
			t.Fatalf("expected result[%d]=%s, got %s", i, id, result[i])
		}
	}
}

func TestParseParticipantFile_SingleID(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "single.csv")
	if err := os.WriteFile(filePath, []byte("abc123\n"), 0600); err != nil {
		t.Fatalf("failed to create temp file: %s", err)
	}

	result, err := parseParticipantFile(filePath)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if len(result) != 1 || result[0] != "abc123" {
		t.Fatalf("expected [abc123], got %v", result)
	}
}

func TestParseParticipantFile_SkipsBlankLines(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "blanks.csv")
	content := "abc123\n\ndef456\n\n"
	if err := os.WriteFile(filePath, []byte(content), 0600); err != nil {
		t.Fatalf("failed to create temp file: %s", err)
	}

	result, err := parseParticipantFile(filePath)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if len(result) != 2 {
		t.Fatalf("expected 2 IDs after skipping blank lines, got %d: %v", len(result), result)
	}
}

func TestParseParticipantFile_TrimsWhitespace(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "whitespace.csv")
	content := "  abc123  \n  def456  \n"
	if err := os.WriteFile(filePath, []byte(content), 0600); err != nil {
		t.Fatalf("failed to create temp file: %s", err)
	}

	result, err := parseParticipantFile(filePath)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if result[0] != "abc123" || result[1] != "def456" {
		t.Fatalf("expected whitespace trimmed, got %v", result)
	}
}

func TestParseParticipantFile_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "empty.csv")
	if err := os.WriteFile(filePath, []byte(""), 0600); err != nil {
		t.Fatalf("failed to create temp file: %s", err)
	}

	_, err := parseParticipantFile(filePath)
	if err == nil {
		t.Fatal("expected error for empty file")
	}
}

func TestParseParticipantFile_NonExistentFile(t *testing.T) {
	_, err := parseParticipantFile("/non/existent/file.csv")
	if err == nil {
		t.Fatal("expected error for non-existent file")
	}
}

func TestParseParticipantFile_CommaInLine(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "comma.csv")
	content := "abc123,extrafield\n"
	if err := os.WriteFile(filePath, []byte(content), 0600); err != nil {
		t.Fatalf("failed to create temp file: %s", err)
	}

	_, err := parseParticipantFile(filePath)
	if err == nil {
		t.Fatal("expected error for line containing a comma")
	}
}

func TestParseParticipantFile_OnlyBlankLines(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "onlyblanks.csv")
	content := "\n\n\n"
	if err := os.WriteFile(filePath, []byte(content), 0600); err != nil {
		t.Fatalf("failed to create temp file: %s", err)
	}

	_, err := parseParticipantFile(filePath)
	if err == nil {
		t.Fatal("expected error when file contains only blank lines")
	}
}
