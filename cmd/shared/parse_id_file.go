package shared

import (
	"fmt"
	"os"
	"strings"
)

// ParseIDFile reads a file containing one ID per line.
// It trims whitespace, skips blank lines, and returns an error
// if the file is empty or contains no valid entries.
func ParseIDFile(filePath string) ([]string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("unable to read file: %w", err)
	}

	content := strings.TrimRight(string(data), "\n\r ")
	if content == "" {
		return nil, fmt.Errorf("file is empty: %s", filePath)
	}

	var ids []string
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		ids = append(ids, line)
	}

	if len(ids) == 0 {
		return nil, fmt.Errorf("no valid entries found in file: %s", filePath)
	}

	return ids, nil
}
