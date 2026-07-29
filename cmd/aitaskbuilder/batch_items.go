package aitaskbuilder

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

// parseBatchItemsInput reads and validates batch_items JSON from a file path or inline string.
// Returns nil if neither is provided (omit from payload).
func parseBatchItemsInput(filePath, jsonStr string) (json.RawMessage, error) {
	if filePath != "" && jsonStr != "" {
		return nil, errors.New(ErrBothBatchItemsInputsProvided)
	}

	var raw []byte
	if filePath != "" {
		var err error
		raw, err = os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read batch-items file: %w", err)
		}
	} else if jsonStr != "" {
		raw = []byte(jsonStr)
	} else {
		return nil, nil
	}

	var pages []json.RawMessage
	if err := json.Unmarshal(raw, &pages); err != nil {
		return nil, errors.New(ErrBatchItemsMustBeArray)
	}
	if len(pages) == 0 {
		return nil, errors.New(ErrBatchItemsMustBeNonEmpty)
	}

	return json.RawMessage(raw), nil
}
