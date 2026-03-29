// Package examples provides embedded study and collection template files.
package examples

import "embed"

//go:embed *.json *.yaml
var FS embed.FS
