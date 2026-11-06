package shared

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveCustomHTML(t *testing.T) {
	dir := t.TempDir()
	html := "<h1>Course</h1>"
	fileName := "course.html"
	if err := os.WriteFile(filepath.Join(dir, fileName), []byte(html), 0600); err != nil {
		t.Fatalf("failed to write html file: %s", err)
	}
	emptyName := "empty.html"
	if err := os.WriteFile(filepath.Join(dir, emptyName), nil, 0600); err != nil {
		t.Fatalf("failed to write empty file: %s", err)
	}
	utf8Name := "utf8.html"
	utf8HTML := "<p>café</p>"
	if err := os.WriteFile(filepath.Join(dir, utf8Name), []byte(utf8HTML), 0600); err != nil {
		t.Fatalf("failed to write utf-8 file: %s", err)
	}
	latin1Name := "latin1.html"
	if err := os.WriteFile(filepath.Join(dir, latin1Name), []byte("caf\xe9"), 0600); err != nil {
		t.Fatalf("failed to write latin1 file: %s", err)
	}

	tests := []struct {
		name        string
		baseDir     string
		htmlFile    string
		htmlContent string
		want        string
		wantErr     string
	}{
		{
			name:     "relative file",
			baseDir:  dir,
			htmlFile: fileName,
			want:     html,
		},
		{
			name:     "absolute file",
			baseDir:  t.TempDir(),
			htmlFile: filepath.Join(dir, fileName),
			want:     html,
		},
		{
			name:        "inline content",
			baseDir:     dir,
			htmlContent: html,
			want:        html,
		},
		{
			name:        "both sources",
			baseDir:     dir,
			htmlFile:    fileName,
			htmlContent: html,
			wantErr:     ErrCustomHTMLBothSources,
		},
		{
			name:    "neither source",
			baseDir: dir,
			wantErr: ErrCustomHTMLNoSource,
		},
		{
			name:     "empty file",
			baseDir:  dir,
			htmlFile: emptyName,
			wantErr:  ErrCustomHTMLEmpty,
		},
		{
			name:     "missing file",
			baseDir:  dir,
			htmlFile: "missing.html",
			wantErr:  filepath.Join(dir, "missing.html"),
		},
		{
			name:     "utf-8 file",
			baseDir:  dir,
			htmlFile: utf8Name,
			want:     utf8HTML,
		},
		{
			name:     "non-utf-8 file",
			baseDir:  dir,
			htmlFile: latin1Name,
			wantErr:  filepath.Join(dir, latin1Name) + ": " + ErrCustomHTMLNotUTF8,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ResolveCustomHTML(tt.baseDir, tt.htmlFile, tt.htmlContent)
			if tt.wantErr != "" {
				if err == nil {
					t.Fatal("expected an error")
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("expected error to contain %q, got %q", tt.wantErr, err.Error())
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			if got != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, got)
			}
		})
	}
}

func TestCustomHTMLSizeCheckBlockLimit(t *testing.T) {
	var check CustomHTMLSizeCheck
	err := check.Add("page 0, item 2", strings.Repeat("a", CustomHTMLBlockMaxChars+1))
	if err == nil {
		t.Fatal("expected an error")
	}
	if !strings.Contains(err.Error(), "page 0, item 2") {
		t.Fatalf("expected error to name the location, got %q", err.Error())
	}
	if !strings.Contains(err.Error(), "300001") {
		t.Fatalf("expected error to contain the length, got %q", err.Error())
	}
}

func TestCustomHTMLSizeCheckTotalLimit(t *testing.T) {
	var check CustomHTMLSizeCheck
	block := strings.Repeat("a", 250_001)
	for i := 0; i < 3; i++ {
		if err := check.Add("page 0, item 0", block); err != nil {
			t.Fatalf("block %d: unexpected error: %s", i, err)
		}
	}

	err := check.Add("page 3, item 0", block)
	if err == nil {
		t.Fatal("expected the fourth block to fail")
	}
	if !strings.Contains(err.Error(), "page 3, item 0") {
		t.Fatalf("expected error to name the fourth block, got %q", err.Error())
	}
	if !strings.Contains(err.Error(), "1000004") {
		t.Fatalf("expected error to contain the total, got %q", err.Error())
	}
}
