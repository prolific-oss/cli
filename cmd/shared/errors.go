package shared

import (
	"net/http"
	"strings"

	"github.com/prolific-oss/cli/client"
)

// ErrCustomHTMLNotEnabled is shown when the workspace gate refuses a custom_html block.
const ErrCustomHTMLNotEnabled = "custom_html is not enabled for your workspace yet"

const customHTMLNotEnabledText = "custom html content blocks are not enabled"

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

// IsCustomHTMLNotEnabledError reports a 403 from the custom_html workspace gate.
// The API owns the refusal text. Delete this check when that gate is removed.
func IsCustomHTMLNotEnabledError(err error) bool {
	if err == nil || !client.IsHTTPStatusError(err, http.StatusForbidden) {
		return false
	}
	return strings.Contains(strings.ToLower(err.Error()), customHTMLNotEnabledText)
}
