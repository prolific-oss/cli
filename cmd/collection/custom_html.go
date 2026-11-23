package collection

import (
	"fmt"

	"github.com/prolific-oss/cli/cmd/shared"
)

// html_file is omitted from the request. A path on another item type would be dropped with no error.
func rejectHTMLFileOnOtherType(pageIndex, itemIndex int, htmlFile string) error {
	if htmlFile == "" {
		return nil
	}
	return fmt.Errorf("page at index %d, item at index %d: html_file is only valid on custom_html items", pageIndex, itemIndex)
}

func readCustomHTML(folder *shared.HTMLFolder, pageIndex, itemIndex int, htmlFile, htmlContent string) (string, error) {
	location := fmt.Sprintf("page at index %d, item at index %d", pageIndex, itemIndex)
	html, err := shared.ResolveCustomHTML(folder, htmlFile, htmlContent)
	if err != nil {
		return "", fmt.Errorf("%s: %w", location, err)
	}
	if err := shared.CheckCustomHTMLSize(location, html); err != nil {
		return "", err
	}
	return html, nil
}
