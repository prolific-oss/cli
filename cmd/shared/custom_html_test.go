package shared

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveCustomHTML(t *testing.T) {
	dir := t.TempDir()
	html := "<h1>Café</h1>"
	fileName := "course.html"
	if err := os.WriteFile(filepath.Join(dir, fileName), []byte(html), 0600); err != nil {
		t.Fatalf("failed to write html file: %s", err)
	}
	emptyName := "empty.html"
	if err := os.WriteFile(filepath.Join(dir, emptyName), nil, 0600); err != nil {
		t.Fatalf("failed to write empty file: %s", err)
	}
	latin1Name := "latin1.html"
	if err := os.WriteFile(filepath.Join(dir, latin1Name), []byte("caf\xe9"), 0600); err != nil {
		t.Fatalf("failed to write latin1 file: %s", err)
	}
	if err := os.Mkdir(filepath.Join(dir, "pages"), 0700); err != nil {
		t.Fatalf("failed to make subfolder: %s", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "pages", fileName), []byte(html), 0600); err != nil {
		t.Fatalf("failed to write html file in subfolder: %s", err)
	}
	tests := []struct {
		name        string
		baseDir     string
		htmlFile    string
		htmlContent string
		want        string
		wantErr     string
		chdir       string
		prepare     func(t *testing.T)
	}{
		{
			name:     "relative file",
			baseDir:  dir,
			htmlFile: fileName,
			want:     html,
		},
		{
			name:     "file in a subfolder",
			baseDir:  dir,
			htmlFile: filepath.Join("pages", fileName),
			want:     html,
		},
		{
			name:     "absolute file",
			baseDir:  t.TempDir(),
			htmlFile: filepath.Join(dir, fileName),
			wantErr:  ErrCustomHTMLOutsideDir,
		},
		{
			name:     "file above the template folder",
			baseDir:  filepath.Join(dir, "pages"),
			htmlFile: filepath.Join("..", fileName),
			wantErr:  ErrCustomHTMLOutsideDir,
		},
		{
			name:     "absolute link inside a relative folder",
			baseDir:  ".",
			htmlFile: "alias.html",
			wantErr:  ErrCustomHTMLOutsideDir,
			chdir:    dir,
			prepare: func(t *testing.T) {
				symlinkOrSkip(t, filepath.Join(dir, fileName), filepath.Join(dir, "alias.html"))
			},
		},
		{
			name:     "relative link inside a relative folder",
			baseDir:  ".",
			htmlFile: "relative-alias.html",
			want:     html,
			chdir:    dir,
			prepare: func(t *testing.T) {
				symlinkOrSkip(t, fileName, filepath.Join(dir, "relative-alias.html"))
			},
		},
		{
			name:     "absolute link outside a relative folder",
			baseDir:  ".",
			htmlFile: "outside-link.html",
			wantErr:  ErrCustomHTMLOutsideDir,
			chdir:    dir,
			prepare: func(t *testing.T) {
				outside := filepath.Join(t.TempDir(), "secret.html")
				if err := os.WriteFile(outside, []byte("<p>Outside page</p>"), 0600); err != nil {
					t.Fatalf("failed to write outside file: %s", err)
				}
				symlinkOrSkip(t, outside, filepath.Join(dir, "outside-link.html"))
			},
		},
		{
			name:     "relative link outside the folder",
			baseDir:  dir,
			htmlFile: "escape.html",
			wantErr:  ErrCustomHTMLOutsideDir,
			prepare: func(t *testing.T) {
				outside := filepath.Join(t.TempDir(), "secret.html")
				if err := os.WriteFile(outside, []byte("<p>secret</p>"), 0600); err != nil {
					t.Fatalf("failed to write outside file: %s", err)
				}
				target, err := filepath.Rel(dir, outside)
				if err != nil {
					t.Fatalf("failed to make a relative link target: %s", err)
				}
				if filepath.IsAbs(target) {
					t.Fatalf("link target %q is absolute, so this case would not reach the real-path check", target)
				}
				symlinkOrSkip(t, target, filepath.Join(dir, "escape.html"))
			},
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
			wantErr:  "html_file " + emptyName + ": " + ErrCustomHTMLFileEmpty,
		},
		{
			name:     "missing file",
			baseDir:  dir,
			htmlFile: "missing.html",
			wantErr:  "failed to read html_file missing.html",
		},
		{
			name:     "non-utf-8 file",
			baseDir:  dir,
			htmlFile: latin1Name,
			wantErr:  "html_file " + latin1Name + ": " + ErrCustomHTMLNotUTF8,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.chdir != "" {
				t.Chdir(tt.chdir)
			}
			if tt.prepare != nil {
				tt.prepare(t)
			}
			got, err := ResolveCustomHTML(openTestFolder(t, tt.baseDir, ""), tt.htmlFile, tt.htmlContent)
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

func TestOpenHTMLFolder(t *testing.T) {
	templateDir := t.TempDir()
	htmlDir := t.TempDir()
	file := filepath.Join(templateDir, "course.html")
	if err := os.WriteFile(file, []byte("<p>x</p>"), 0600); err != nil {
		t.Fatalf("failed to write file: %s", err)
	}

	folder, err := OpenHTMLFolder(templateDir, "")
	if err != nil || folder.path != templateDir {
		t.Fatalf("expected template folder %q, got %+v (err %v)", templateDir, folder, err)
	}
	if err := folder.Close(); err != nil {
		t.Fatalf("unexpected error closing the template folder: %s", err)
	}

	folder, err = OpenHTMLFolder(templateDir, htmlDir)
	if err != nil || folder.path != htmlDir {
		t.Fatalf("expected html-dir %q, got %+v (err %v)", htmlDir, folder, err)
	}
	if folder.root == nil {
		t.Fatal("expected the html-dir to be open")
	}
	if err := folder.Close(); err != nil {
		t.Fatalf("unexpected error closing the html-dir: %s", err)
	}

	if _, err := OpenHTMLFolder(templateDir, filepath.Join(htmlDir, "missing")); err == nil {
		t.Fatal("expected an error for a missing html-dir")
	}

	if _, err := OpenHTMLFolder(templateDir, file); err == nil || !strings.Contains(err.Error(), "is not a folder") {
		t.Fatalf("expected a not-a-folder error, got %v", err)
	}
}

// A folder opened once must still be read after that folder is replaced.
func TestHTMLFolderStaysOnTheOpenedFolder(t *testing.T) {
	root := t.TempDir()
	htmlDir := filepath.Join(root, "pages")
	if err := os.Mkdir(htmlDir, 0700); err != nil {
		t.Fatalf("failed to make folder: %s", err)
	}
	original := "<h1>Original</h1>"
	if err := os.WriteFile(filepath.Join(htmlDir, "course.html"), []byte(original), 0600); err != nil {
		t.Fatalf("failed to write html file: %s", err)
	}

	folder := openTestFolder(t, t.TempDir(), htmlDir)

	if err := os.Rename(htmlDir, filepath.Join(root, "pages-moved")); err != nil {
		t.Skipf("the operating system does not allow moving an open folder: %s", err)
	}
	if err := os.Mkdir(htmlDir, 0700); err != nil {
		t.Fatalf("failed to make replacement folder: %s", err)
	}
	if err := os.WriteFile(filepath.Join(htmlDir, "course.html"), []byte("<h1>Replacement</h1>"), 0600); err != nil {
		t.Fatalf("failed to write replacement file: %s", err)
	}

	got, err := ResolveCustomHTML(folder, "course.html", "")
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if got != original {
		t.Fatalf("expected the opened folder to be read, got %q", got)
	}
}

// These tests call the read directly and skip the path check.
// The open folder must still refuse a path that leaves it.
// The parent-path case uses no symbolic link, so it runs where links are unavailable.
func TestReadInsideRootRefusesEscapes(t *testing.T) {
	dir := t.TempDir()
	outsideDir := t.TempDir()
	secret := filepath.Join(outsideDir, "secret.html")
	if err := os.WriteFile(secret, []byte("<p>secret</p>"), 0600); err != nil {
		t.Fatalf("failed to write outside file: %s", err)
	}

	t.Run("symbolic link to a file outside", func(t *testing.T) {
		symlinkOrSkip(t, secret, filepath.Join(dir, "file-link.html"))
		data, err := readInsideRoot(openTestRoot(t, dir), "file-link.html")
		if err == nil {
			t.Fatalf("expected an error, read %q", data)
		}
	})

	t.Run("parent path", func(t *testing.T) {
		data, err := readInsideRoot(openTestRoot(t, dir), filepath.Join("..", filepath.Base(outsideDir), "secret.html"))
		if err == nil {
			t.Fatalf("expected an error, read %q", data)
		}
	})
}

func TestResolveCustomHTMLFileSizeLimit(t *testing.T) {
	dir := t.TempDir()
	atLimit := strings.Repeat("a", CustomHTMLBlockMaxBytes)
	if err := os.WriteFile(filepath.Join(dir, "at-limit.html"), []byte(atLimit), 0600); err != nil {
		t.Fatalf("failed to write file: %s", err)
	}
	overLimit := strings.Repeat("a", CustomHTMLBlockMaxBytes+1)
	if err := os.WriteFile(filepath.Join(dir, "over-limit.html"), []byte(overLimit), 0600); err != nil {
		t.Fatalf("failed to write file: %s", err)
	}

	folder := openTestFolder(t, dir, "")
	got, err := ResolveCustomHTML(folder, "at-limit.html", "")
	if err != nil {
		t.Fatalf("unexpected error at the limit: %s", err)
	}
	if got != atLimit {
		t.Fatalf("expected the full file at the limit, got %d bytes", len(got))
	}

	_, err = ResolveCustomHTML(folder, "over-limit.html", "")
	if err == nil {
		t.Fatal("expected an error one byte over the limit")
	}
	if !errors.Is(err, errCustomHTMLFileTooLarge) {
		t.Fatalf("expected size error, got %s", err)
	}
	if strings.Contains(err.Error(), "failed to read") {
		t.Fatalf("expected a size error, not a read error, got %s", err)
	}
}

func TestCheckCustomHTMLSize(t *testing.T) {
	tests := []struct {
		name    string
		size    int
		wantErr bool
	}{
		{name: "at the limit", size: CustomHTMLBlockMaxBytes},
		{name: "one byte over the limit", size: CustomHTMLBlockMaxBytes + 1, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CheckCustomHTMLSize("page at index 0, item at index 2", strings.Repeat("a", tt.size))
			if !tt.wantErr {
				if err != nil {
					t.Fatalf("unexpected error: %s", err)
				}
				return
			}
			if err == nil {
				t.Fatal("expected an error")
			}
			if !strings.Contains(err.Error(), "page at index 0, item at index 2") {
				t.Fatalf("expected error to name the location, got %q", err.Error())
			}
			if !strings.Contains(err.Error(), "300001") {
				t.Fatalf("expected error to contain the length, got %q", err.Error())
			}
		})
	}
}

func openTestFolder(t *testing.T, templateDir, htmlDir string) *HTMLFolder {
	t.Helper()
	folder, err := OpenHTMLFolder(templateDir, htmlDir)
	if err != nil {
		t.Fatalf("failed to open the html folder: %s", err)
	}
	t.Cleanup(func() { _ = folder.Close() })
	return folder
}

func openTestRoot(t *testing.T, dir string) *os.Root {
	t.Helper()
	root, err := os.OpenRoot(dir)
	if err != nil {
		t.Fatalf("failed to open the folder: %s", err)
	}
	t.Cleanup(func() { _ = root.Close() })
	return root
}

// symlinkOrSkip creates a symbolic link. It skips the test when the operating system
// refuses the link, as Windows does without Developer Mode or an elevated process.
// An existing name fails the test. That is a test mistake.
func symlinkOrSkip(t *testing.T, target, name string) {
	t.Helper()
	if err := os.Symlink(target, name); err != nil {
		if errors.Is(err, fs.ErrExist) {
			t.Fatalf("failed to make link: %s", err)
		}
		t.Skipf("symbolic links are not available: %s", err)
	}
}
