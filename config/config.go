// Package config provides default configuration values for the Prolific CLI,
// including base URLs for the Prolific application and API.
package config

import (
	"strings"

	"github.com/spf13/viper"
)

// DefaultApplicationURL is the default Prolific application URL.
const DefaultApplicationURL = "https://app.prolific.com"

// GetApplicationURL will return the Application URL. This can be overridden
// using the PROLIFIC_APPLICATION_URL environment variable.
func GetApplicationURL() string {
	viper.SetDefault("PROLIFIC_APPLICATION_URL", DefaultApplicationURL)
	return strings.TrimRight(viper.GetString("PROLIFIC_APPLICATION_URL"), "/")
}

// GetAPIURL will return the API URL. This is the default API, but we allow
// users to override this with an environment variable.
func GetAPIURL() string {
	return "https://api.prolific.com"
}
