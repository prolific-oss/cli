package collection

import (
	"fmt"

	"github.com/prolific-oss/cli/cmd/shared"
)

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
