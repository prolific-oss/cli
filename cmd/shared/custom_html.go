package shared

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

// Supports L&D content in TaskBuilder.

// Copy of CUSTOM_HTML_BLOCK_MAX_BYTES in the data-collection tool.
// That constant is the source of truth. The CLI copies it so an oversized
// file fails before the request is sent.
const CustomHTMLBlockMaxBytes = 300_000

var errCustomHTMLFileTooLarge = fmt.Errorf("the file is more than %d bytes, the limit per block", CustomHTMLBlockMaxBytes)

const (
	ErrCustomHTMLBothSources = "custom_html item must set html_file or html_content, not both"
	ErrCustomHTMLNoSource    = "custom_html item must set html_file or html_content"
	ErrCustomHTMLFileEmpty   = "html_file is empty"
	ErrCustomHTMLNotUTF8     = "html_file is not valid UTF-8; save the file as UTF-8"
	ErrCustomHTMLOutsideDir  = "html_file must be a relative path inside the html folder"
)

// ResolveCustomHTML returns the HTML to send for a custom_html item.
// htmlFile must stay inside the folder after symbolic links are followed, so a
// shared template can read only files from its own folder.
// The path check makes the error clear. The open folder handle enforces the
// folder for the whole run, so a later change to a link or to the folder
// cannot redirect the read.
func ResolveCustomHTML(folder *HTMLFolder, htmlFile, htmlContent string) (string, error) {
	if htmlFile != "" && htmlContent != "" {
		return "", errors.New(ErrCustomHTMLBothSources)
	}
	if htmlFile == "" && htmlContent == "" {
		return "", errors.New(ErrCustomHTMLNoSource)
	}
	if htmlContent != "" {
		return htmlContent, nil
	}

	if err := checkHTMLPath(folder.path, htmlFile); err != nil {
		return "", err
	}

	root, err := folder.open()
	if err != nil {
		return "", fmt.Errorf("failed to read html_file %s: %w", htmlFile, err)
	}
	data, err := readInsideRoot(root, htmlFile)
	if errors.Is(err, errCustomHTMLFileTooLarge) {
		return "", fmt.Errorf("html_file %s: %w", htmlFile, err)
	}
	if err != nil {
		return "", fmt.Errorf("failed to read html_file %s: %w", htmlFile, err)
	}
	if len(data) == 0 {
		return "", fmt.Errorf("html_file %s: %s", htmlFile, ErrCustomHTMLFileEmpty)
	}
	if !utf8.Valid(data) {
		return "", fmt.Errorf("html_file %s: %s", htmlFile, ErrCustomHTMLNotUTF8)
	}

	return string(data), nil
}

// HTMLFolder is the folder html_file paths are read from.
// After open it keeps one handle, so replacing the folder or a link in its path
// cannot redirect later reads. Close it when the command finishes.
// It is not safe for concurrent use.
type HTMLFolder struct {
	path string
	root *os.Root
}

// OpenHTMLFolder returns the folder html_file paths are read from.
// An empty htmlDir uses the template folder and opens it on the first html_file,
// so a template with no html_file does not open a folder. --html-dir is opened at once.
func OpenHTMLFolder(templateDir, htmlDir string) (*HTMLFolder, error) {
	folder := &HTMLFolder{path: templateDir}
	if htmlDir == "" {
		return folder, nil
	}

	info, err := os.Stat(htmlDir)
	if err != nil {
		return nil, fmt.Errorf("html-dir %s: %w", htmlDir, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("html-dir %s is not a folder", htmlDir)
	}

	folder.path = htmlDir
	if _, err := folder.open(); err != nil {
		return nil, fmt.Errorf("html-dir %s: %w", htmlDir, err)
	}
	return folder, nil
}

// Close releases the folder handle, including when the folder was never opened.
func (f *HTMLFolder) Close() error {
	if f.root == nil {
		return nil
	}
	err := f.root.Close()
	f.root = nil
	return err
}

func (f *HTMLFolder) open() (*os.Root, error) {
	if f.root != nil {
		return f.root, nil
	}
	root, err := os.OpenRoot(f.path)
	if err != nil {
		return nil, err
	}
	f.root = root
	return root, nil
}

// checkHTMLPath refuses a path that leaves baseDir, so the error is clear.
// The open folder handle is the real limit. This check runs before the open
// and can be raced.
func checkHTMLPath(baseDir, htmlFile string) error {
	if filepath.IsAbs(htmlFile) {
		return fmt.Errorf("html_file %s: %s", htmlFile, ErrCustomHTMLOutsideDir)
	}

	path := filepath.Join(baseDir, htmlFile)
	if !isInsideDir(baseDir, path) {
		return fmt.Errorf("html_file %s: %s", htmlFile, ErrCustomHTMLOutsideDir)
	}

	// os.Root rejects an absolute link target even inside the folder.
	// Reject it here so the error stays clear.
	if err := absoluteSymlink(baseDir, htmlFile); err != nil {
		return err
	}

	realBase, err := filepath.EvalSymlinks(baseDir)
	if err != nil {
		return fmt.Errorf("failed to read html_file %s: %w", htmlFile, err)
	}
	realPath, err := filepath.EvalSymlinks(path)
	if err != nil {
		return fmt.Errorf("failed to read html_file %s: %w", htmlFile, err)
	}
	if !isInsideDir(realBase, realPath) {
		return fmt.Errorf("html_file %s: %s", htmlFile, ErrCustomHTMLOutsideDir)
	}

	return nil
}

// readInsideRoot reads name through the open folder handle. The operating system
// refuses a step that leaves the folder, including a link swapped in after the check.
// A link that stays inside the folder still works. Reading one byte past the limit
// refuses an oversized file before it is loaded in full.
func readInsideRoot(root *os.Root, name string) ([]byte, error) {
	file, err := root.Open(name)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	data, err := io.ReadAll(io.LimitReader(file, CustomHTMLBlockMaxBytes+1))
	if err != nil {
		return nil, err
	}
	if len(data) > CustomHTMLBlockMaxBytes {
		return nil, errCustomHTMLFileTooLarge
	}
	return data, nil
}

// absoluteSymlink reports an absolute link target in htmlFile.
// os.Root refuses an absolute target even when it points inside the folder, so this check copies that rule to keep the error clear.
// It checks each path part without following the link, so the target itself is tested.
func absoluteSymlink(baseDir, htmlFile string) error {
	current := baseDir
	for _, part := range strings.Split(filepath.FromSlash(htmlFile), string(filepath.Separator)) {
		if part == "" || part == "." {
			continue
		}
		current = filepath.Join(current, part)
		info, err := os.Lstat(current)
		if err != nil {
			return fmt.Errorf("failed to read html_file %s: %w", htmlFile, err)
		}
		if info.Mode()&os.ModeSymlink == 0 {
			continue
		}
		target, err := os.Readlink(current)
		if err != nil {
			return fmt.Errorf("failed to read html_file %s: %w", htmlFile, err)
		}
		if filepath.IsAbs(target) {
			return fmt.Errorf("html_file %s: %s", htmlFile, ErrCustomHTMLOutsideDir)
		}
	}
	return nil
}

func isInsideDir(dir, path string) bool {
	rel, err := filepath.Rel(dir, path)
	if err != nil {
		return false
	}
	return rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

// CheckCustomHTMLSize checks one block against the per-block UTF-8 byte limit.
// location names the block, for example "page at index 0, item at index 2".
func CheckCustomHTMLSize(location, html string) error {
	size := len(html)
	if size > CustomHTMLBlockMaxBytes {
		return fmt.Errorf("%s: custom_html is %d bytes, the limit per block is %d", location, size, CustomHTMLBlockMaxBytes)
	}
	return nil
}
