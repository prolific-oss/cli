package client

import (
	"testing"
)

func TestFormatBatchErrorBody(t *testing.T) {
	tests := []struct {
		name     string
		body     []byte
		expected string
	}{
		{
			name:     "non-JSON body returned as-is",
			body:     []byte("internal server error"),
			expected: "internal server error",
		},
		{
			name:     "non-INVALID_BATCH_ITEMS error returned as-is",
			body:     []byte(`{"type":"SOME_OTHER_ERROR","message":"something went wrong"}`),
			expected: `{"type":"SOME_OTHER_ERROR","message":"something went wrong"}`,
		},
		{
			name:     "INVALID_BATCH_ITEMS with no issues",
			body:     []byte(`{"type":"INVALID_BATCH_ITEMS","issues":[]}`),
			expected: "batch_items validation failed:",
		},
		{
			name:     "INVALID_BATCH_ITEMS with single issue, no field",
			body:     []byte(`{"type":"INVALID_BATCH_ITEMS","issues":[{"page":0,"row":0,"column":0,"item":0,"type":"free_text","message":"description is required"}]}`),
			expected: "batch_items validation failed:\n  Page 1, Row 1, Column 1, Item 1 (free_text): description is required",
		},
		{
			name:     "INVALID_BATCH_ITEMS with field reference",
			body:     []byte(`{"type":"INVALID_BATCH_ITEMS","issues":[{"page":0,"row":1,"column":0,"item":2,"type":"dataset_field","field":"missing_col","message":"Field does not exist in the dataset schema"}]}`),
			expected: "batch_items validation failed:\n  Page 1, Row 2, Column 1, Item 3 (dataset_field) \"missing_col\": Field does not exist in the dataset schema",
		},
		{
			name:     "INVALID_BATCH_ITEMS with multiple issues",
			body:     []byte(`{"type":"INVALID_BATCH_ITEMS","issues":[{"page":0,"row":0,"column":0,"item":0,"type":"free_text","message":"description is required"},{"page":1,"row":2,"column":1,"item":3,"type":"multiple_choice","message":"answer_limit exceeds number of options"}]}`),
			expected: "batch_items validation failed:\n  Page 1, Row 1, Column 1, Item 1 (free_text): description is required\n  Page 2, Row 3, Column 2, Item 4 (multiple_choice): answer_limit exceeds number of options",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatBatchErrorBody(tt.body)
			if got != tt.expected {
				t.Fatalf("expected:\n%q\ngot:\n%q", tt.expected, got)
			}
		})
	}
}
