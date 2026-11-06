package shared

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"unicode/utf8"
)

// This code is to facilitate the L&D content via AITB initiative

// One DynamoDB item holds at most 400 KB. 300,000 bytes of HTML leaves room
// for the other fields on that item. A participant response returns every
// block, so a collection or batch holds at most 1,000,000 bytes. The names
// say chars to match DCT, but the checks below count bytes.
const (
	CustomHTMLBlockMaxChars             = 300_000
	CustomHTMLCollectionOrBatchMaxChars = 1_000_000
)

const (
	ErrCustomHTMLBothSources = "custom_html item must set html_file or html_content, not both"
	ErrCustomHTMLNoSource    = "custom_html item must set html_file or html_content"
	ErrCustomHTMLEmpty       = "custom_html item has empty html_content"
	ErrCustomHTMLNotUTF8     = "html_file is not valid UTF-8; save the file as UTF-8"
)

// ResolveCustomHTML returns the HTML to send for a custom_html item.
// A relative htmlFile is resolved against baseDir, so a template can refer
// to an HTML file beside it from any working directory.
func ResolveCustomHTML(baseDir, htmlFile, htmlContent string) (string, error) {
	if htmlFile != "" && htmlContent != "" {
		return "", errors.New(ErrCustomHTMLBothSources)
	}
	if htmlFile == "" && htmlContent == "" {
		return "", errors.New(ErrCustomHTMLNoSource)
	}
	if htmlContent != "" {
		return htmlContent, nil
	}

	path := htmlFile
	if !filepath.IsAbs(path) {
		path = filepath.Join(baseDir, path)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read html_file %s: %w", path, err)
	}
	if len(data) == 0 {
		return "", errors.New(ErrCustomHTMLEmpty)
	}
	if !utf8.Valid(data) {
		return "", fmt.Errorf("html_file %s: %s", path, ErrCustomHTMLNotUTF8)
	}

	return string(data), nil
}

// CustomHTMLSizeCheck checks custom_html blocks against the per-block and
// total size limits. Sizes are UTF-8 bytes, which are never fewer than the
// UTF-16 code units that the server counts.
type CustomHTMLSizeCheck struct {
	total int
}

// Add checks one block. location names the block, for example "page 0, item 2".
func (c *CustomHTMLSizeCheck) Add(location, html string) error {
	size := len(html)
	if size > CustomHTMLBlockMaxChars {
		return fmt.Errorf("%s: custom_html is %d bytes, the limit per block is %d", location, size, CustomHTMLBlockMaxChars)
	}

	c.total += size
	if c.total > CustomHTMLCollectionOrBatchMaxChars {
		return fmt.Errorf("%s: custom_html total is %d bytes, the limit is %d", location, c.total, CustomHTMLCollectionOrBatchMaxChars)
	}

	return nil
}
