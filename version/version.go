// Package version holds build-time version information for the CLI.
package version

import (
	"runtime/debug"
	"strings"
)

// GITCOMMIT is injected at build time via -ldflags (e.g. "0.0.62").
// Defaults to "dev" when not set.
var GITCOMMIT string = "dev"

// Get returns the CLI version. It prefers the value injected via ldflags,
// then falls back to the Go module version embedded in the binary (e.g. when
// installed via `go install github.com/prolific-oss/cli/...@vX.Y.Z`).
func Get() string {
	if GITCOMMIT != "dev" {
		return GITCOMMIT
	}
	if info, ok := debug.ReadBuildInfo(); ok {
		version := strings.TrimPrefix(info.Main.Version, "v")
		if version != "" && !strings.Contains(version, "-") {
			return version
		}
	}
	return GITCOMMIT
}
