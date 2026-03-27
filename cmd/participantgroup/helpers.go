package participantgroup

import (
	"fmt"
	"os"
	"strings"
)

// parseParticipantFile reads a file and returns a slice of participant IDs.
// Each non-empty line is treated as one participant ID.
func parseParticipantFile(filePath string) ([]string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("unable to read file: %w", err)
	}

	content := strings.TrimRight(string(data), "\n\r ")
	if content == "" {
		return nil, fmt.Errorf("file is empty: %s", filePath)
	}

	var ids []string
	for i, line := range strings.Split(content, "\n") {
		id := strings.TrimSpace(line)
		if id == "" {
			continue
		}
		if strings.Contains(id, ",") {
			return nil, fmt.Errorf("line %d: unexpected comma — expected one participant ID per line", i+1)
		}
		ids = append(ids, id)
	}

	if len(ids) == 0 {
		return nil, fmt.Errorf("no participant IDs found in file: %s", filePath)
	}

	return ids, nil
}
