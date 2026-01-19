// Package config provides default configuration values for the Prolific CLI,
// including base URLs for the Prolific application and API.
package config

// GetApplicationURL will return the Application URL. This could be updated
// to understand different environments based on API URL perhaps?
func GetApplicationURL() string {
	return "https://app.prolific.com"
}

// GetAPIURL will return the API URL. This is the default API, but we allow
// users to override this with an environment variable.
func GetAPIURL() string {
	return "https://api.prolific.com"
}
