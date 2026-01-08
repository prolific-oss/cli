package collection

import (
	"fmt"
	"testing"

	"github.com/prolific-oss/cli/ui"
)

func TestIsFeatureNotEnabledError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error returns false",
			err:      nil,
			expected: false,
		},
		{
			name:     "feature not enabled error with 404 and feature not enabled fragment returns true",
			err:      fmt.Errorf("request failed with status 404: %s", ui.FeatureNotEnabledErrorFragment),
			expected: true,
		},
		{
			name:     "feature not enabled error with mixed case returns true",
			err:      fmt.Errorf("request failed with status 404: You do not currently have permission to access to this feature."),
			expected: true,
		},
		{
			name:     "generic 404 without feature message returns false",
			err:      fmt.Errorf("request failed with status 404: not found"),
			expected: false,
		},
		{
			name:     "HTTP 403 status code error returns false",
			err:      fmt.Errorf("request failed with status 403: forbidden"),
			expected: false,
		},
		{
			name:     "generic error is not a feature access error",
			err:      fmt.Errorf("something went wrong"),
			expected: false,
		},
		{
			name:     "API error with status 404 but no feature message returns false",
			err:      fmt.Errorf("unable to fulfil request /api/v1/data-collection/collections: request failed with status 404: {\"detail\":\"Not found.\"}"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isFeatureNotEnabledError(tt.err)
			if result != tt.expected {
				t.Errorf("isFeatureNotEnabledError(%v) = %v, expected %v", tt.err, result, tt.expected)
			}
		})
	}
}
