package shared

import (
	"strings"
)

// IsFeatureNotEnabledError checks if the error indicates that a feature
// is not enabled for the user. Uses semantic matching on key phrases for
// robustness against minor wording changes.
func IsFeatureNotEnabledError(err error) bool {
	if err == nil {
		return false
	}

	errMsg := strings.ToLower(err.Error())

	// Check for API request failure with feature access denial
	// This matches messages like "request failed: you do not currently have permission to access [to] this feature"
	hasRequestFailed := strings.Contains(errMsg, "request failed")
	hasPermission := strings.Contains(errMsg, "permission")
	hasFeature := strings.Contains(errMsg, "feature")

	return hasRequestFailed && hasPermission && hasFeature
}
