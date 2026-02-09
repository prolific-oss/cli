package bonus

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseBonusEntries_ValidInline(t *testing.T) {
	entries := []string{"pid1,4.25", "pid2,3.50"}
	result, err := parseBonusEntries(entries)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	expected := "pid1,4.25\npid2,3.50"
	if result != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%s'", expected, result)
	}
}

func TestParseBonusEntries_SingleEntry(t *testing.T) {
	entries := []string{"pid1,10.00"}
	result, err := parseBonusEntries(entries)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	expected := "pid1,10.00"
	if result != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%s'", expected, result)
	}
}

func TestParseBonusEntries_EmptyID(t *testing.T) {
	entries := []string{",4.25"}
	_, err := parseBonusEntries(entries)
	if err == nil {
		t.Fatal("expected error for empty ID")
	}
}

func TestParseBonusEntries_NegativeAmount(t *testing.T) {
	entries := []string{"pid1,-4.25"}
	_, err := parseBonusEntries(entries)
	if err == nil {
		t.Fatal("expected error for negative amount")
	}
}

func TestParseBonusEntries_ZeroAmount(t *testing.T) {
	entries := []string{"pid1,0"}
	_, err := parseBonusEntries(entries)
	if err == nil {
		t.Fatal("expected error for zero amount")
	}
}

func TestParseBonusEntries_NaN(t *testing.T) {
	entries := []string{"pid1,NaN"}
	_, err := parseBonusEntries(entries)
	if err == nil {
		t.Fatal("expected error for NaN amount")
	}
}

func TestParseBonusEntries_Inf(t *testing.T) {
	entries := []string{"pid1,+Inf"}
	_, err := parseBonusEntries(entries)
	if err == nil {
		t.Fatal("expected error for Inf amount")
	}
}

func TestParseBonusEntries_NonNumericAmount(t *testing.T) {
	entries := []string{"pid1,abc"}
	_, err := parseBonusEntries(entries)
	if err == nil {
		t.Fatal("expected error for non-numeric amount")
	}
}

func TestParseBonusEntries_MissingComma(t *testing.T) {
	entries := []string{"pid1"}
	_, err := parseBonusEntries(entries)
	if err == nil {
		t.Fatal("expected error for missing comma separator")
	}
}

func TestParseBonusFile_ValidCSV(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "bonuses.csv")
	content := "pid1,4.25\npid2,3.50\n"
	if err := os.WriteFile(filePath, []byte(content), 0600); err != nil {
		t.Fatalf("failed to create temp file: %s", err)
	}

	result, err := parseBonusFile(filePath)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	expected := "pid1,4.25\npid2,3.50"
	if result != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%s'", expected, result)
	}
}

func TestParseBonusFile_NonExistentFile(t *testing.T) {
	_, err := parseBonusFile("/non/existent/file.csv")
	if err == nil {
		t.Fatal("expected error for non-existent file")
	}
}

func TestParseBonusFile_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "empty.csv")
	if err := os.WriteFile(filePath, []byte(""), 0600); err != nil {
		t.Fatalf("failed to create temp file: %s", err)
	}

	_, err := parseBonusFile(filePath)
	if err == nil {
		t.Fatal("expected error for empty file")
	}
}

func TestParseBonusFile_MalformedLines(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "bad.csv")
	content := "pid1\npid2,3.50\n"
	if err := os.WriteFile(filePath, []byte(content), 0600); err != nil {
		t.Fatalf("failed to create temp file: %s", err)
	}

	_, err := parseBonusFile(filePath)
	if err == nil {
		t.Fatal("expected error for malformed lines")
	}
}

func TestParseBonusFile_ExtraWhitespace(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "whitespace.csv")
	content := "  pid1 , 4.25 \n pid2 , 3.50 \n"
	if err := os.WriteFile(filePath, []byte(content), 0600); err != nil {
		t.Fatalf("failed to create temp file: %s", err)
	}

	result, err := parseBonusFile(filePath)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	// Should have trimmed whitespace
	if strings.Contains(result, " ") {
		t.Fatalf("expected whitespace to be trimmed, got: '%s'", result)
	}
}

func TestConfirmPayment_Yes(t *testing.T) {
	reader := strings.NewReader("y\n")
	var buf bytes.Buffer

	result, err := confirmPayment("bonus-123", false, reader, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if !result {
		t.Fatal("expected confirmation to return true")
	}

	output := buf.String()
	if !strings.Contains(output, "bonus-123") {
		t.Fatalf("expected prompt to contain bonus ID, got: '%s'", output)
	}
}

func TestConfirmPayment_No(t *testing.T) {
	reader := strings.NewReader("n\n")
	var buf bytes.Buffer

	result, err := confirmPayment("bonus-123", false, reader, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if result {
		t.Fatal("expected confirmation to return false")
	}
}

func TestConfirmPayment_NonInteractive(t *testing.T) {
	var buf bytes.Buffer

	result, err := confirmPayment("bonus-123", true, nil, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if !result {
		t.Fatal("expected non-interactive to return true")
	}
}
