package shared

import (
	"fmt"
	"testing"
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
			name:     "feature not enabled error with structured JSON response returns true",
			err:      fmt.Errorf("request failed: You do not currently have permission to access to this feature."),
			expected: true,
		},
		{
			name:     "feature not enabled error with status code format returns true",
			err:      fmt.Errorf("request failed with status 404: You do not currently have permission to access to this feature."),
			expected: true,
		},
		{
			name:     "feature not enabled error without typo returns true",
			err:      fmt.Errorf("request failed: You do not currently have permission to access this feature."),
			expected: true,
		},
		{
			name:     "feature not enabled error with mixed case returns true",
			err:      fmt.Errorf("request failed: you DO NOT have PERMISSION to access this FEATURE"),
			expected: true,
		},
		{
			name:     "request failed without feature message returns false",
			err:      fmt.Errorf("request failed: not found"),
			expected: false,
		},
		{
			name:     "request failed with only permission keyword returns false",
			err:      fmt.Errorf("request failed: you do not have permission"),
			expected: false,
		},
		{
			name:     "request failed with only feature keyword returns false",
			err:      fmt.Errorf("request failed: this feature is unavailable"),
			expected: false,
		},
		{
			name:     "generic error without request failed returns false",
			err:      fmt.Errorf("something went wrong with permission and feature"),
			expected: false,
		},
		{
			name:     "API error with not found detail returns false",
			err:      fmt.Errorf("unable to fulfil request: request failed with status 404: {\"detail\":\"Not found.\"}"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsFeatureNotEnabledError(tt.err)
			if result != tt.expected {
				t.Errorf("IsFeatureNotEnabledError(%v) = %v, expected %v", tt.err, result, tt.expected)
			}
		})
	}
}
