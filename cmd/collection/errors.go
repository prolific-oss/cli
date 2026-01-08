package collection

import "strings"

// isFeatureNotEnabledError checks if the error indicates that the Collections feature
// is not enabled for the user. The backend converts 403 (feature flag off) → 404 for
// non-staff users, so we detect 404 errors with the feature not enabled fragment in the response.
func isFeatureNotEnabledError(err error) bool {
	if err == nil {
		return false
	}

	errMsg := strings.ToLower(err.Error())

	// Check for HTTP 404 status code patterns from API responses
	if strings.Contains(errMsg, "status 404") && strings.Contains(errMsg, FeatureNotEnabledErrorFragment) {
		return true
	}

	return false
}
